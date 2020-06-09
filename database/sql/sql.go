/**------------------------------------------------------------**
 * @filename sql/
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-07-19 11:29
 * @desc     go.jd100.com - sql -
 **------------------------------------------------------------**/
package sql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"go.jd100.com/medusa/common/constants"
	_ "go.jd100.com/medusa/database/sql/mysql"
	"go.jd100.com/medusa/log"

	"go.jd100.com/medusa/ctime"
	"go.jd100.com/medusa/errors"
	"go.jd100.com/medusa/net/netutil/breaker"
	"go.jd100.com/medusa/net/trace"
	"go.jd100.com/medusa/stat"
)

const (
	_mysql  = "mysql"
	_family = "sql_client"
)

var (
	// ErrStmtNil prepared stmt error
	ErrStmtNil = errors.New("sql: prepare failed and stmt nil")
	// ErrNoMaster is returned by Master when call master multiple times.
	ErrNoMaster = errors.New("sql: no master instance")
	// ErrNoRows is returned by Scan when QueryRow doesn't return a row.
	// In such a case, QueryRow returns a placeholder *Row value that defers
	// this error until a Scan.
	ErrNoRows = sql.ErrNoRows
	// ErrTxDone transaction done.
	ErrTxDone = sql.ErrTxDone

	stats = stat.DB
)

type Result sql.Result

type DB struct {
	write  *conn
	read   []*conn
	idx    int64
	master *DB
}

type Config struct {
	Addr         string          // for trace
	DSN          string          // write data source name.
	ReadDSN      []string        // read data source name.
	Active       int             // pool
	Idle         int             // pool
	IdleTimeout  ctime.Duration  // connect max life time.
	QueryTimeout ctime.Duration  // query sql timeout
	ExecTimeout  ctime.Duration  // execute sql timeout
	TranTimeout  ctime.Duration  // transaction sql timeout
	Breaker      *breaker.Config // breaker
}

type conn struct {
	*sql.DB

	conf    *Config
	breaker breaker.Breaker
}

// Tx transaction.
type Tx struct {
	db     *conn
	tx     *sql.Tx
	t      trace.Trace
	c      context.Context
	cancel func()
}

// Row row.
type Row struct {
	err error
	*sql.Row
	db     *conn
	query  string
	args   []interface{}
	t      trace.Trace
	cancel func()
}

// Scan copies the columns from the matched row into the values pointed at by dest.
func (r *Row) Scan(dest ...interface{}) (err error) {
	if r.t != nil {
		defer r.t.Finish(&err)
	}
	if r.err != nil {
		err = r.err
	} else if r.Row == nil {
		err = ErrStmtNil
	}
	if err != nil {
		return
	}
	err = r.Row.Scan(dest...)
	if r.cancel != nil {
		r.cancel()
	}
	r.db.onBreaker(&err)
	if err != ErrNoRows {
		err = errors.Wrapf(err, "query %s args %+v", r.query, r.args)
	}
	return
}

// Rows rows.
type Rows struct {
	*sql.Rows
	cancel func()
}

// Close closes the Rows, preventing further enumeration. If Next is called
// and returns false and there are no further result sets,
// the Rows are closed automatically and it will suffice to check the
// result of Err. Close is idempotent and does not affect the result of Err.
func (rs *Rows) Close() (err error) {
	err = errors.WithStack(rs.Rows.Close())
	if rs.cancel != nil {
		rs.cancel()
	}
	return
}

// Stmt prepared stmt.
type Stmt struct {
	db    *conn
	tx    bool
	query string
	stmt  atomic.Value
	t     trace.Trace
}

func newDB(adapter string, c *Config) (db *DB) {
	if c.QueryTimeout == 0 || c.ExecTimeout == 0 || c.TranTimeout == 0 {
		panic("db config must be set query/execute/transaction timeout")
	}
	var err error
	db, err = Open(adapter, c)
	if err != nil {
		log.Errorf("open %s error(%v)", adapter, err)
		panic(err)
	}
	return
}

// Open opens a database specified by its database driver name and a
// driver-specific data source name, usually consisting of at least a database
// name and connection information.
func Open(adapter string, c *Config) (*DB, error) {
	db := new(DB)
	d, err := connect(adapter, c, c.DSN)
	if err != nil {
		return nil, err
	}
	brkGroup := breaker.NewGroup(c.Breaker)
	brk := brkGroup.Get(c.Addr)
	w := &conn{DB: d, breaker: brk, conf: c}
	rs := make([]*conn, 0, len(c.ReadDSN))
	for _, rd := range c.ReadDSN {
		d, err := connect(adapter, c, rd)
		if err != nil {
			return nil, err
		}
		brk := brkGroup.Get(parseDSNAddr(rd))
		r := &conn{DB: d, breaker: brk, conf: c}
		rs = append(rs, r)
	}
	db.write = w
	db.read = rs
	db.master = &DB{write: db.write}
	return db, nil
}

func connect(adapter string, c *Config, dataSourceName string) (*sql.DB, error) {
	d, err := sql.Open(adapter, dataSourceName)
	if err != nil {
		err = errors.WithStack(err)
		return nil, err
	}
	d.SetMaxOpenConns(c.Active)
	d.SetMaxIdleConns(c.Idle)
	d.SetConnMaxLifetime(time.Duration(c.IdleTimeout))
	return d, nil
}

// Begin starts a transaction. The isolation level is dependent on the driver.
func (db *DB) Begin(c context.Context) (tx *Tx, err error) {
	return db.write.begin(c)
}

// Exec executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
func (db *DB) Exec(c context.Context, query string, args ...interface{}) (res sql.Result, err error) {
	return db.write.exec(c, query, args...)
}

// Prepare creates a prepared statement for later queries or executions.
// Multiple queries or executions may be run concurrently from the returned
// statement. The caller must call the statement's Close method when the
// statement is no longer needed.
func (db *DB) Prepare(query string) (*Stmt, error) {
	return db.write.prepare(query)
}

// Prepared creates a prepared statement for later queries or executions.
// Multiple queries or executions may be run concurrently from the returned
// statement. The caller must call the statement's Close method when the
// statement is no longer needed.
func (db *DB) Prepared(query string) (stmt *Stmt) {
	return db.write.prepared(query)
}

// Query executes a query that returns rows, typically a SELECT. The args are
// for any placeholder parameters in the query.
func (db *DB) Query(c context.Context, query string, args ...interface{}) (rows *Rows, err error) {
	idx := db.readIndex()
	for i := range db.read {
		if rows, err = db.read[(idx+i)%len(db.read)].query(c, query, args...); !errors.EqualError(errors.ServiceUnavailable, err) {
			return
		}
	}
	return db.write.query(c, query, args...)
}

// QueryRow executes a query that is expected to return at most one row.
// QueryRow always returns a non-nil value. Errors are deferred until Row's
// Scan method is called.
func (db *DB) QueryRow(c context.Context, query string, args ...interface{}) *Row {
	idx := db.readIndex()
	for i := range db.read {
		if row := db.read[(idx+i)%len(db.read)].queryRow(c, query, args...); !errors.EqualError(errors.ServiceUnavailable, row.err) {
			return row
		}
	}
	return db.write.queryRow(c, query, args...)
}

func (db *DB) readIndex() int {
	if len(db.read) == 0 {
		return 0
	}
	v := atomic.AddInt64(&db.idx, 1)
	return int(v) % len(db.read)
}

// Close closes the write and read database, releasing any open resources.
func (db *DB) Close() (err error) {
	if e := db.write.Close(); e != nil {
		err = errors.WithStack(e)
	}
	for _, rd := range db.read {
		if e := rd.Close(); e != nil {
			err = errors.WithStack(e)
		}
	}
	return
}

// Ping verifies a connection to the database is still alive, establishing a
// connection if necessary.
func (db *DB) Ping(c context.Context) (err error) {
	if err = db.write.ping(c); err != nil {
		return
	}
	for _, rd := range db.read {
		if err = rd.ping(c); err != nil {
			return
		}
	}
	return
}

// Master return *DB instance direct use master conn
// use this *DB instance only when you have some reason need to get result without any delay.
func (db *DB) Master() *DB {
	if db.master == nil {
		panic(ErrNoMaster)
	}
	return db.master
}

func (db *conn) onBreaker(err *error) {
	if err != nil && *err != nil && *err != sql.ErrNoRows && *err != sql.ErrTxDone {
		db.breaker.MarkFailed()
	} else {
		db.breaker.MarkSuccess()
	}
}

func (db *conn) begin(c context.Context) (tx *Tx, err error) {
	now := time.Now()
	t, ok := trace.FromContext(c)
	if ok {
		t = t.Fork(_family, "begin")
		t.SetTag(trace.String(trace.TagAddress, db.conf.Addr), trace.String(trace.TagComment, ""))
		defer func() {
			if err != nil {
				t.Finish(&err)
			}
		}()
	}
	if err = db.breaker.Allow(); err != nil {
		stats.Incr("mysql:begin", "breaker")
		return
	}
	_, c, cancel := db.conf.TranTimeout.Shrink(c)
	rtx, err := db.BeginTx(c, nil)
	stats.Timing("mysql:begin", int64(time.Since(now)/time.Millisecond))
	if err != nil {
		err = errors.WithStack(err)
		cancel()
		return
	}
	tx = &Tx{tx: rtx, t: t, db: db, c: c, cancel: cancel}
	return
}

func (db *conn) exec(c context.Context, query string, args ...interface{}) (res sql.Result, err error) {
	now := time.Now()
	if t, ok := trace.FromContext(c); ok {
		t = t.Fork(_family, "exec")
		t.SetTag(trace.String(trace.TagAddress, db.conf.Addr), trace.String(trace.TagComment, query))
		defer t.Finish(&err)
	}
	if err = db.breaker.Allow(); err != nil {
		stats.Incr("mysql:exec", "breaker")
		return
	}
	_, c, cancel := db.conf.ExecTimeout.Shrink(c)
	res, err = db.ExecContext(c, query, args...)
	cancel()
	db.onBreaker(&err)
	stats.Timing("mysql:exec", int64(time.Since(now)/time.Millisecond))
	if err != nil {
		err = errors.Wrapf(err, "exec:%s, args:%+v", query, args)
	}
	return
}

func (db *conn) ping(c context.Context) (err error) {
	now := time.Now()
	if t, ok := trace.FromContext(c); ok {
		t = t.Fork(_family, "ping")
		t.SetTag(trace.String(trace.TagAddress, db.conf.Addr), trace.String(trace.TagComment, ""))
		defer t.Finish(&err)
	}
	if err = db.breaker.Allow(); err != nil {
		stats.Incr("mysql:ping", "breaker")
		return
	}
	_, c, cancel := db.conf.ExecTimeout.Shrink(c)
	err = db.PingContext(c)
	cancel()
	db.onBreaker(&err)
	stats.Timing("mysql:ping", int64(time.Since(now)/time.Millisecond))
	if err != nil {
		err = errors.WithStack(err)
	}
	return
}

func (db *conn) prepare(query string) (*Stmt, error) {
	stmt, err := db.Prepare(query)
	if err != nil {
		err = errors.Wrapf(err, "prepare %s", query)
		return nil, err
	}
	st := &Stmt{query: query, db: db}
	st.stmt.Store(stmt)
	return st, nil
}

func (db *conn) prepared(query string) (stmt *Stmt) {
	stmt = &Stmt{query: query, db: db}
	s, err := db.Prepare(query)
	if err == nil {
		stmt.stmt.Store(s)
		return
	}
	go func() {
		for {
			s, err := db.Prepare(query)
			if err != nil {
				time.Sleep(time.Second)
				continue
			}
			stmt.stmt.Store(s)
			return
		}
	}()
	return
}

func (db *conn) query(c context.Context, query string, args ...interface{}) (rows *Rows, err error) {
	now := time.Now()
	if t, ok := trace.FromContext(c); ok {
		t = t.Fork(_family, "query")
		t.SetTag(trace.String(trace.TagAddress, db.conf.Addr), trace.String(trace.TagComment, query))
		defer t.Finish(&err)
	}
	if err = db.breaker.Allow(); err != nil {
		stats.Incr("mysql:query", "breaker")
		return
	}
	_, c, cancel := db.conf.QueryTimeout.Shrink(c)
	rs, err := db.DB.QueryContext(c, query, args...)
	db.onBreaker(&err)
	stats.Timing("mysql:query", int64(time.Since(now)/time.Millisecond))
	if err != nil {
		err = errors.Wrapf(err, "query:%s, args:%+v", query, args)
		cancel()
		return
	}
	rows = &Rows{Rows: rs, cancel: cancel}
	return
}

func (db *conn) queryRow(c context.Context, query string, args ...interface{}) *Row {
	now := time.Now()
	t, ok := trace.FromContext(c)
	if ok {
		t = t.Fork(_family, "queryrow")
		t.SetTag(trace.String(trace.TagAddress, db.conf.Addr), trace.String(trace.TagComment, query))
	}
	if err := db.breaker.Allow(); err != nil {
		stats.Incr("mysql:queryrow", "breaker")
		return &Row{db: db, t: t, err: err}
	}
	_, c, cancel := db.conf.QueryTimeout.Shrink(c)
	r := db.DB.QueryRowContext(c, query, args...)
	stats.Timing("mysql:queryrow", int64(time.Since(now)/time.Millisecond))
	return &Row{db: db, Row: r, query: query, args: args, t: t, cancel: cancel}
}

// Close closes the statement.
func (s *Stmt) Close() (err error) {
	if s == nil {
		err = ErrStmtNil
		return
	}
	stmt, ok := s.stmt.Load().(*sql.Stmt)
	if ok {
		err = errors.WithStack(stmt.Close())
	}
	return
}

// Exec executes a prepared statement with the given arguments and returns a
// Result summarizing the effect of the statement.
func (s *Stmt) Exec(c context.Context, args ...interface{}) (res sql.Result, err error) {
	if s == nil {
		err = ErrStmtNil
		return
	}
	now := time.Now()
	if s.tx {
		if s.t != nil {
			s.t.SetTag(trace.String(trace.TagAnnotation, s.query))
		}
	} else if t, ok := trace.FromContext(c); ok {
		t = t.Fork(_family, "exec")
		t.SetTag(trace.String(trace.TagAddress, s.db.conf.Addr), trace.String(trace.TagComment, s.query))
		defer t.Finish(&err)
	}
	if err = s.db.breaker.Allow(); err != nil {
		stats.Incr("mysql:stmt:exec", "breaker")
		return
	}
	stmt, ok := s.stmt.Load().(*sql.Stmt)
	if !ok {
		err = ErrStmtNil
		return
	}
	_, c, cancel := s.db.conf.ExecTimeout.Shrink(c)
	res, err = stmt.ExecContext(c, args...)
	cancel()
	s.db.onBreaker(&err)
	stats.Timing("mysql:stmt:exec", int64(time.Since(now)/time.Millisecond))
	if err != nil {
		err = errors.Wrapf(err, "exec:%s, args:%+v", s.query, args)
	}
	return
}

// Query executes a prepared query statement with the given arguments and
// returns the query results as a *Rows.
func (s *Stmt) Query(c context.Context, args ...interface{}) (rows *Rows, err error) {
	if s == nil {
		err = ErrStmtNil
		return
	}
	if s.tx {
		if s.t != nil {
			s.t.SetTag(trace.String(trace.TagAnnotation, s.query))
		}
	} else if t, ok := trace.FromContext(c); ok {
		t = t.Fork(_family, "query")
		t.SetTag(trace.String(trace.TagAddress, s.db.conf.Addr), trace.String(trace.TagComment, s.query))
		defer t.Finish(&err)
	}
	if err = s.db.breaker.Allow(); err != nil {
		stats.Incr("mysql:stmt:query", "breaker")
		return
	}
	stmt, ok := s.stmt.Load().(*sql.Stmt)
	if !ok {
		err = ErrStmtNil
		return
	}
	now := time.Now()
	_, c, cancel := s.db.conf.QueryTimeout.Shrink(c)
	rs, err := stmt.QueryContext(c, args...)
	s.db.onBreaker(&err)
	stats.Timing("mysql:stmt:query", int64(time.Since(now)/time.Millisecond))
	if err != nil {
		err = errors.Wrapf(err, "query:%s, args:%+v", s.query, args)
		cancel()
		return
	}
	rows = &Rows{Rows: rs, cancel: cancel}
	return
}

// QueryRow executes a prepared query statement with the given arguments.
// If an error occurs during the execution of the statement, that error will
// be returned by a call to Scan on the returned *Row, which is always non-nil.
// If the query selects no rows, the *Row's Scan will return ErrNoRows.
// Otherwise, the *Row's Scan scans the first selected row and discards the rest.
func (s *Stmt) QueryRow(c context.Context, args ...interface{}) (row *Row) {
	now := time.Now()
	row = &Row{db: s.db, query: s.query, args: args}
	if s == nil {
		row.err = ErrStmtNil
		return
	}
	if s.tx {
		if s.t != nil {
			s.t.SetTag(trace.String(trace.TagAnnotation, s.query))
		}
	} else if t, ok := trace.FromContext(c); ok {
		t = t.Fork(_family, "queryrow")
		t.SetTag(trace.String(trace.TagAddress, s.db.conf.Addr), trace.String(trace.TagComment, s.query))
		row.t = t
	}
	if row.err = s.db.breaker.Allow(); row.err != nil {
		stats.Incr("mysql:stmt:queryrow", "breaker")
		return
	}
	stmt, ok := s.stmt.Load().(*sql.Stmt)
	if !ok {
		return
	}
	_, c, cancel := s.db.conf.QueryTimeout.Shrink(c)
	row.Row = stmt.QueryRowContext(c, args...)
	row.cancel = cancel
	stats.Timing("mysql:stmt:queryrow", int64(time.Since(now)/time.Millisecond))
	return
}

// Commit commits the transaction.
func (tx *Tx) Commit() (err error) {
	err = tx.tx.Commit()
	tx.cancel()
	tx.db.onBreaker(&err)
	if tx.t != nil {
		tx.t.Finish(&err)
	}
	if err != nil {
		err = errors.WithStack(err)
	}
	return
}

// Rollback aborts the transaction.
func (tx *Tx) Rollback() (err error) {
	err = tx.tx.Rollback()
	tx.cancel()
	tx.db.onBreaker(&err)
	if tx.t != nil {
		tx.t.Finish(&err)
	}
	if err != nil {
		err = errors.WithStack(err)
	}
	return
}

// Exec executes a query that doesn't return rows. For example: an INSERT and
// UPDATE.
func (tx *Tx) Exec(query string, args ...interface{}) (res sql.Result, err error) {
	now := time.Now()
	if tx.t != nil {
		tx.t.SetTag(trace.String(trace.TagAnnotation, fmt.Sprintf("exec %s", query)))
	}
	res, err = tx.tx.ExecContext(tx.c, query, args...)
	stats.Timing("mysql:tx:exec", int64(time.Since(now)/time.Millisecond))
	if err != nil {
		err = errors.Wrapf(err, "exec:%s, args:%+v", query, args)
	}
	return
}

// Query executes a query that returns rows, typically a SELECT.
func (tx *Tx) Query(query string, args ...interface{}) (rows *Rows, err error) {
	if tx.t != nil {
		tx.t.SetTag(trace.String(trace.TagAnnotation, fmt.Sprintf("query %s", query)))
	}
	now := time.Now()
	defer func() {
		stats.Timing("mysql:tx:query", int64(time.Since(now)/time.Millisecond))
	}()
	rs, err := tx.tx.QueryContext(tx.c, query, args...)
	if err == nil {
		rows = &Rows{Rows: rs}
	} else {
		err = errors.Wrapf(err, "query:%s, args:%+v", query, args)
	}
	return
}

// QueryRow executes a query that is expected to return at most one row.
// QueryRow always returns a non-nil value. Errors are deferred until Row's
// Scan method is called.
func (tx *Tx) QueryRow(query string, args ...interface{}) *Row {
	if tx.t != nil {
		tx.t.SetTag(trace.String(trace.TagAnnotation, fmt.Sprintf("queryrow %s", query)))
	}
	now := time.Now()
	defer func() {
		stats.Timing("mysql:tx:queryrow", int64(time.Since(now)/time.Millisecond))
	}()
	r := tx.tx.QueryRowContext(tx.c, query, args...)
	return &Row{Row: r, db: tx.db, query: query, args: args}
}

// Stmt returns a transaction-specific prepared statement from an existing statement.
func (tx *Tx) Stmt(stmt *Stmt) *Stmt {
	as, ok := stmt.stmt.Load().(*sql.Stmt)
	if !ok {
		return nil
	}
	ts := tx.tx.StmtContext(tx.c, as)
	st := &Stmt{query: stmt.query, tx: true, t: tx.t, db: tx.db}
	st.stmt.Store(ts)
	return st
}

// Prepare creates a prepared statement for use within a transaction.
// The returned statement operates within the transaction and can no longer be
// used once the transaction has been committed or rolled back.
// To use an existing prepared statement on this transaction, see Tx.Stmt.
func (tx *Tx) Prepare(query string) (*Stmt, error) {
	if tx.t != nil {
		tx.t.SetTag(trace.String(trace.TagAnnotation, fmt.Sprintf("prepare %s", query)))
	}
	stmt, err := tx.tx.Prepare(query)
	if err != nil {
		err = errors.Wrapf(err, "prepare %s", query)
		return nil, err
	}
	st := &Stmt{query: query, tx: true, t: tx.t, db: tx.db}
	st.stmt.Store(stmt)
	return st, nil
}

// parseDSNAddr parse dsn name and return addr.
func parseDSNAddr(dsn string) (addr string) {
	if dsn == "" {
		return
	}
	part0 := strings.Split(dsn, "@")
	if len(part0) > 1 {
		part1 := strings.Split(part0[1], "?")
		if len(part1) > 0 {
			addr = part1[0]
		}
	}
	return
}

func ParseArgs(values interface{}) []interface{} {
	switch values.(type) {
	case []int8:
		vals := values.([]int8)
		var params = make([]interface{}, len(vals))
		for i, v := range vals {
			params[i] = v
		}
		return params
	case []int64:
		vals := values.([]int64)
		var params = make([]interface{}, len(vals))
		for i, v := range vals {
			params[i] = v
		}
		return params
	case []string:
		vals := values.([]string)
		var params = make([]interface{}, len(vals))
		for i, v := range vals {
			params[i] = v
		}
		return params
	default:
		vals := values.([]interface{})
		var params = make([]interface{}, len(vals))
		for i, v := range vals {
			params[i] = v
		}
		return params
	}
}

func BuildSqlIn(inValues interface{}) (sql string, params []interface{}) {
	var ph []string
	switch inValues.(type) {
	case []int8, []int, []int32, []int64:
		vals := inValues.([]int64)
		vlen := len(vals)
		if vlen == 0 {
			return
		}
		params = make([]interface{}, vlen)
		for i, v := range vals {
			ph = append(ph, constants.PlaceHolder)
			params[i] = v
		}
	case []string:
		vals := inValues.([]string)
		vlen := len(vals)
		if vlen == 0 {
			return
		}
		params = make([]interface{}, vlen)
		for i, v := range vals {
			params[i] = v
		}
	default:
		return
	}
	sql = fmt.Sprintf("%s (%s)", constants.In, strings.Join(ph, ","))
	return
}
