/**------------------------------------------------------------**
 * @filename mgdb/mg.go
 * @author   jinycoo - caojingyin@jiandan100.cn
 * @version  1.0.0
 * @date     2019/11/13 14:46
 * @desc     go.jd100.com - mgdb - mongodb conn
 **------------------------------------------------------------**/
package mgdb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo/readpref"

	"go.jd100.com/medusa/log"
	"go.jd100.com/medusa/net/trace"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	_family = "mg_client"
)

type DB struct {
	t           trace.Trace
	conf        *Config
	collections []*mongo.Collection
	ClientDB    *mongo.Database
}

func NewMgDB(c *Config) (db *DB) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.Timeout))
	defer cancel()
	clt, err := mongo.Connect(ctx, options.Client().ApplyURI(c.DSN))
	if err != nil {
		log.Errorf("mongo connect err (%v)", err)
	}
	db = &DB{
		ClientDB: clt.Database(c.Database),
	}
	return
}

func (mg *DB) Close() error {
	return mg.ClientDB.Client().Disconnect(context.TODO())
}

func (mg *DB) Ping(ctx context.Context) (err error) {
	if t, ok := trace.FromContext(ctx); ok {
		t = t.Fork(_family, "ping")
		t.SetTag(trace.String(trace.TagAddress, mg.conf.Addr), trace.String(trace.TagComment, ""))
		defer t.Finish(&err)
	}
	return mg.ClientDB.Client().Ping(ctx, readpref.Primary())
}
