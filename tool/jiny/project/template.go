/**------------------------------------------------------------**
 * @filename project/xxx.go
 * @author   jinycoo - caojingyin@jiandan100.cn
 * @version  1.0.0
 * @date     2019/11/4 18:21
 * @desc     go.jd100.com - main - summary
 **------------------------------------------------------------**/
package project

const (
	_tplAppToml = `# Project base config setting.
name    = "{{.Name}}-{{.Module}}"
version = "1.0.0"
port    = ":80"
appID   = 1
# log setting default output stderr with json format.
[log]
    level = "info"
    filters = ["instance_id", "zone"]
# mysql database setting.
[mysql]
	addr = "127.0.0.1:3306"
	dsn = "{user}:{password}@tcp(127.0.0.1:3306)/{database}?timeout=1s&readTimeout=1s&writeTimeout=1s&parseTime=true&loc=Local&charset=utf8mb4,utf8"
	readDSN = ["{user}:{password}@tcp(127.0.0.2:3306)/{database}?timeout=1s&readTimeout=1s&writeTimeout=1s&parseTime=true&loc=Local&charset=utf8mb4,utf8","{user}:{password}@tcp(127.0.0.3:3306)/{database}?timeout=1s&readTimeout=1s&writeTimeout=1s&parseTime=true&loc=Local&charset=utf8,utf8mb4"]
	active = 20
	idle = 10
	idleTimeout ="4h"
	queryTimeout = "200ms"
	execTimeout = "300ms"
	tranTimeout = "400ms"
# cache - redis setting.
redisExpire = "24h"
[redis]
    name = "{{.Name}}-{{.Module}}"
    proto = "tcp"
    addr = "127.0.0.1:6379"
    password = ""
    db = 8
    idle = 100
    active = 100
    dialTimeout = "1s"
    readTimeout = "1s"
    writeTimeout = "1s"
    idleTimeout = "10s"
# mq - rabbit mq setting.
[mq]
    dsn = "amqp://{user}:{password}@{host}:5672/{vhost}"
    [mq.exchange]
        name = "{exchange_name}"
        type = "{type}"
        routingKey = "{routing_key}"
        declare = true
        durable = true
        autoDelete = false
        internal = false
        noWait = false
        [mq.exchange.queue]
             name = "{queue_name}"
# rpc - grpc setting.
[rpc.g]
    addr = "0.0.0.0:9000"
    timeout = "1s"
`

	_tplChangeLog = `## {{.Module}}/{{.Name}}

### v1.0.0
1. 上线功能xxx
`
	_tplContributors = `# Owner
{{.Owner}}

# Author

# Reviewer
`
	_tplReadme = `# {{.Module}}/{{.Name}}

## 项目简介
1.
`

	_tplMain = `/**------------------------------------------------------------**
 * @filename cmd/main.go
 * @author   {{.Owner}} - {{.Owner}}@jiandan100.cn
 * @version  1.0.0
 * @date     {{.Date}}
 * @desc     go.jd100.com - {{.Module}}-{{.Name}} - main
 **------------------------------------------------------------**/
package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.jd100.com/{{.Module}}/{{.Name}}/conf"
	"go.jd100.com/{{.Module}}/{{.Name}}/server/http"
	"go.jd100.com/{{.Module}}/{{.Name}}/service"

	"go.jd100.com/medusa/errors"
	"go.jd100.com/medusa/log"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	errors.Init(conf.Conf.Lang)
	log.Init(conf.Conf.Log, conf.Conf.Name)
	defer log.Sync()
	log.Info("{{.Name}}-{{.Module}} start")
	svc := service.New(conf.Conf)
	http.New(conf.Conf, svc)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Infof("get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			// ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
			// if err := httpSrv.Shutdown(ctx); err != nil {
			// 	log.Error("httpSrv.Shutdown error(%v)", err)
			// }
			log.Info("{{.Name}}-{{.Module}} exit")
			svc.Close()
			// cancel()
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
`
	_tplConf = `/**------------------------------------------------------------**
 * @filename conf/conf.go
 * @author   {{.Owner}} - {{.Owner}}@jiandan100.cn
 * @version  1.0.0
 * @date     {{.Date}}
 * @desc     go.jd100.com - {{.Module}}-{{.Name}} - conf
 **------------------------------------------------------------**/
package conf

import (
	"os"
	"path/filepath"

	"go.jd100.com/medusa/cache/redis"
	"go.jd100.com/medusa/config"
	"go.jd100.com/medusa/ctime"
	"go.jd100.com/medusa/database/sql"
	"go.jd100.com/medusa/errors"
	"go.jd100.com/medusa/log"
	"go.jd100.com/medusa/queue/rabbitmq"
	"go.jd100.com/medusa/utils"

	"github.com/BurntSushi/toml"
)

var (
	confPath string
	clt      *config.Client
	Conf   = &Config{}
)

type Config struct {
	Name          string
	Lang          string
	AppID         int
	Version       string
	Port          string

	Log           *log.Config
	Mysql         *sql.Config // *MysqlDB
	Mq            *rabbitmq.Config
	Redis         *redis.Config
	RedisExpire   ctime.Duration
}

// type MysqlDB struct {
// 	Db       *sql.Config
// 	Account  *sql.Config
// }

func Init() error {
	if confPath != "" {
		return local()
	} else {
		confPath = filepath.Join(utils.RootDir(), config.DefConfigFile)
		_, err := os.Stat(confPath)
		if err == nil {
			return local()
		}else {
			return remote()
		}
	}
}

func local() (err error) {
	_, err = toml.DecodeFile(confPath, &Conf)
	return
}

func remote() (err error) {
	if clt, err = config.New(); err != nil {
		return
	}
	if err = load(); err != nil {
		return
	}
	go func() {
		for range clt.Event() {
			if load() != nil {
				log.Errorf("config reload error (%v)", err)
			}
		}
	}()
	return err
}

func load() (err error) {
	var (
		s       string
		ok      bool
		tmpConf *Config
	)
	if s, ok = clt.Toml2(); !ok {
		return errors.New("load config center error")
	}
	if _, err = toml.Decode(s, &tmpConf); err != nil {
		return errors.New("could not decode config")
	}
	*Conf = *tmpConf
	return
}
`
	_tplGRPCMain = `/**------------------------------------------------------------**
 * @filename cmd/main.go
 * @author   {{.Owner}} - {{.Owner}}@jiandan100.cn
 * @version  1.0.0
 * @date     {{.Date}}
 * @desc     go.jd100.com - {{.Module}}-{{.Name}} - grpc main
 **------------------------------------------------------------**/
package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.jd100.com/{{.Module}}/{{.Name}}/conf"
	"go.jd100.com/{{.Module}}/{{.Name}}/server/grpc"
	"go.jd100.com/{{.Module}}/{{.Name}}/server/http"
	"go.jd100.com/{{.Module}}/{{.Name}}/service"

	"go.jd100.com/medusa/errors"
	"go.jd100.com/medusa/log"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	errors.Init(conf.Conf.Lang)
	log.Init(conf.Conf.Log, conf.Conf.Name)
	defer log.Sync()

	log.Info("{{.Name}}-{{.Module}} start")
	svc := service.New(conf.Conf)
	grpcSrv := grpc.New(svc)
	go httpSrv := http.New(svc)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
			if err := grpcSrv.Shutdown(ctx); err != nil {
				log.Error("grpcSrv.Shutdown error(%v)", err)
			}
			if err := httpSrv.Shutdown(ctx); err != nil {
				log.Error("httpSrv.Shutdown error(%v)", err)
			}
			log.Info("{{.Name}}-{{.Module}} exit")
			svc.Close()
			cancel()
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
`

	_tplDao = `/**------------------------------------------------------------**
 * @filename dao/dao.go
 * @author   {{.Owner}} - {{.Owner}}@jiandan100.cn
 * @version  1.0.0
 * @date     {{.Date}}
 * @desc     go.jd100.com - {{.Module}}-{{.Name}} - dao
 **------------------------------------------------------------**/
package dao

import (
	"context"
	"time"

	"go.jd100.com/{{.Module}}/{{.Name}}/conf"

	"go.jd100.com/medusa/cache/redis"
	"go.jd100.com/medusa/database/sql"
)

// Dao
type Dao struct {
	c           *conf.Config
	db          *sql.DB
	redis       *redis.Client
	redisExpire int32
}

// New new a dao and return.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
		// mysql
		db: sql.NewMySQL(c.Mysql),
		// redis
		redis:       redis.NewClient(c.Redis),
		redisExpire: int32(time.Duration(c.RedisExpire) / time.Second),
	}
	return
}

// Close close the resource.
func (d *Dao) Close() {
	d.redis.Close()
	d.db.Close()
}

// Ping ping the resource.
func (d *Dao) Ping(ctx context.Context) (err error) {
	if _, err = d.redis.Ping().Result(); err != nil {
		return
	}
	return d.db.Ping(ctx)
}

`
	_tplDaoMysql = `/**------------------------------------------------------------**
 * @filename dao/mysql.go
 * @author   {{.Owner}} - {{.Owner}}@jiandan100.cn
 * @version  1.0.0
 * @date     {{.Date}}
 * @desc     go.jd100.com - {{.Module}}-{{.Name}} - dao mysql scripts
 **------------------------------------------------------------**/
package dao

import (
	"context"

	"go.jd100.com/medusa/database/sql"
)

const (
	_shard = 100

	// db_name - table_name
	_findDemoSQL      = "SELECT mid, account FROM demo WHERE %s;"
	_countDemoSQL     = "SELECT COUNT(1) FROM demo WHERE %s;"
	_addDemoSQL       = "INSERT INTO demo (mid, account) VALUES (?, ?);"
	_batchAddDemoSQL  = "INSERT INTO demo(mid, account) VALUES "
	_editDemoSQL      = "UPDATE demo SET account = ? WHERE mid = ?;"
	_delDemoSQL       = "UPDATE demo SET deleted_at = ? WHERE mid = ?;"
)

func hit(id int64) int64 {
	return id % _shard
}

func (d *Dao) BeginTran(c context.Context) (tx *sql.Tx, err error) {
	return d.db.Begin(c)
}

`

	_tplService = `/**------------------------------------------------------------**
 * @filename service/service.go
 * @author   {{.Owner}} - {{.Owner}}@jiandan100.cn
 * @version  1.0.0
 * @date     {{.Date}}
 * @desc     go.jd100.com - {{.Module}}-{{.Name}} - service
 **------------------------------------------------------------**/
package service

import (
	"context"

	"go.jd100.com/{{.Module}}/{{.Name}}/conf"
	"go.jd100.com/{{.Module}}/{{.Name}}/dao"
)

// Service service.
type Service struct {
	conf  *conf.Config
	dao   *dao.Dao
}

// New new a service and return.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		conf:  c,
		dao: dao.New(c),
	}
	return
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context) (err error) {
	return s.dao.Ping(ctx)
}

// Close close the resource.
func (s *Service) Close() {
	s.dao.Close()
}

`
	_tplServiceTest = `/**------------------------------------------------------------**
 * @filename {{.Namespace}}/xxx_test.go
 * @author   {{.Owner}} - {{.Owner}}@jiandan100.cn
 * @version  1.0.0
 * @date     {{.Date}}
 * @desc     go.jd100.com - {{.Module}}-{{.Name}} - xxxx test
 **------------------------------------------------------------**/
package {{.Namespace}}

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMethodName(t *testing.T) {
	assert.Equal(t, "/", "/")
}

`

	_tplGPRCService = `/**------------------------------------------------------------**
 * @filename service/service.go
 * @author   {{.Owner}} - {{.Owner}}@jiandan100.cn
 * @version  1.0.0
 * @date     {{.Date}}
 * @desc     go.jd100.com - {{.Module}}-{{.Name}} - grpc service
 **------------------------------------------------------------**/
package service

import (
	"context"
	"fmt"

	pb "go.jd100.com/{{.Module}}/{{.Name}}/api"
	"go.jd100.com/{{.Module}}/{{.Name}}/conf"
	"go.jd100.com/{{.Module}}/{{.Name}}/dao"

	"github.com/golang/protobuf/ptypes/empty"
)

// Service service.
type Service struct {
	ac  *paladin.Map
	dao *dao.Dao
}

// New new a service and return.
func New() (s *Service) {
	s = &Service{
		conf:  c,
		dao: dao.New(c),
	}
	return
}

// SayHello grpc demo func.
func (s *Service) SayHello(ctx context.Context, req *pb.HelloReq) (reply *empty.Empty, err error) {
	reply = new(empty.Empty)
	fmt.Printf("hello %s", req.Name)
	return
}

// SayHelloURL bm demo func.
func (s *Service) SayHelloURL(ctx context.Context, req *pb.HelloReq) (reply *pb.HelloResp, err error) {
	reply = &pb.HelloResp{
		Content: "hello " + req.Name,
	}
	fmt.Printf("hello url %s", req.Name)
	return
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context) (err error) {
	return s.dao.Ping(ctx)
}

// Close close the resource.
func (s *Service) Close() {
	s.dao.Close()
}
`
	_tplHTTPServer = `/**------------------------------------------------------------**
 * @filename http/http.go
 * @author   {{.Owner}} - {{.Owner}}@jiandan100.cn
 * @version  1.0.0
 * @date     {{.Date}}
 * @desc     go.jd100.com - {{.Module}}-{{.Name}} - http router..
 **------------------------------------------------------------**/
package http

import (
	"go.jd100.com/{{.Module}}/{{.Name}}/conf"
	"go.jd100.com/{{.Module}}/{{.Name}}/service"

	"go.jd100.com/medusa/log"
	"go.jd100.com/medusa/net/http/jiny"
)

var sv *service.Service

// Init a api http server.
func New(c *conf.Config, service *service.Service) {
	sv = service
	
	jiny.Index(index) //  root /
	jiny.Ping(ping)   //  ping /ping

	v1 := jiny.Group("/v1")
	{
		initRouter(v1)
	}
	log.Infof("http listening and serving HTTP on %s", c.Port)
	if err := jiny.Run(c.Port); err != nil {
		log.Errorf("api.Start error(%v)", err)
		panic(err)
	}
}

func initRouter(g *jiny.RouterGroup) {
	g.GET("/", v1Index) // version(v1) /v1/
}

func ping(c *jiny.Context) {
	if err := sv.Ping(c); err != nil {
		log.Errorf("ping error(%v)", err)
		c.JSON("ping error", err)
		return
	}
	c.JSON("everything is good!", nil)
}

// index handler.
func index(c *jiny.Context) {
	c.JSON("{{.Name}}-{{.Module}} is running.", nil)
}

func v1Index(c *jiny.Context) {
	c.JSON("{{.Name}}-{{.Module}} version_v1 is running.", nil)
}
`
	_tplPBHTTPServer = `/**------------------------------------------------------------**
 * @filename http/http.go
 * @author   {{.Owner}} - {{.Owner}}@jiandan100.cn
 * @version  1.0.0
 * @date     {{.Date}}
 * @desc     go.jd100.com - {{.Module}}-{{.Name}} - grpc pb http
 **------------------------------------------------------------**/
package http

import (
	"net/http"

	pb "go.jd100.com/{{.Module}}/{{.Name}}/api"
	"go.jd100.com/{{.Module}}/{{.Name}}/conf"
	"go.jd100.com/{{.Module}}/{{.Name}}/service"

	"go.jd100.com/medusa/log"
	"go.jd100.com/medusa/net/http/jiny"
	"go.jd100.com/medusa/net/rpc/warden"
)

var svc *service.Service

// Init a api http server.
func Init(c *conf.Config, service *service.Service) {
	svc = service
	pb.RegisterMemberServer(c.ServerConfig, &server{as: s})
	jiny.Ping(ping)
	jiny.Index(index)
	log.Infof("grpc http listening and serving HTTP on %s", c.Port)
	if err := jiny.Run(c.Port); err != nil {
		log.Errorf("api.Start error(%v)", err)
		panic(err)
	}

}

func initRouter(g *jiny.RouterGroup) {
	g.GET("/", howGh)
}

func howGh(c *jiny.Context) {
	k := &model.Medusa{
		Hello: "Golang 大法好 !!!",
	}
	c.JSON(k, nil)
}

func ping(ctx *jiny.Context) {
	if err := svc.Ping(ctx); err != nil {
		log.Error("ping error(%v)", err)
		ctx.AbortWithStatus(http.StatusServiceUnavailable)
	}
}

// index handler.
func index(c *jiny.Context) {
	c.JSON("welcome to grpc http server index", nil)
}

`

	_tplAPIProto = `// 定义项目 API 的 proto 文件 可以同时描述 gRPC 和 HTTP API
// protobuf 文件参考:
//  - https://developers.google.com/protocol-buffers/
syntax = "proto3";

import "github.com/gogo/protobuf/gogoproto/gogo.proto";
import "google/protobuf/empty.proto";
import "google/api/annotations.proto";

// package 命名使用 {appid}.{version} 的方式, version 形如 v1, v2 ..
package member.service.v1;

option go_package = "api";
option (gogoproto.goproto_getters_all) = false;

service Member {
	rpc GetMInfoByMID(MIDReq) returns (MemberInfoReply);
    rpc GetMInfoByAccount(AccountReq) returns (MemberInfoReply);
}

message MIDReq {
	int64 mid = 1 [(gogoproto.moretags)='form:"mid" validate:"gt=0,required"'];
    string real_ip = 2;
}

message AccountReq {
    string account = 1 [(gogoproto.moretags) = '"validate:"required"'];
    string real_ip = 2;
}

message MemberInfoReply {
    MemberInfo info = 1 [(gogoproto.jsontag) = 'minfo'];
}

`
	_tplModel = `/**------------------------------------------------------------**
 * @filename model/model.go
 * @author   {{.Owner}} - {{.Owner}}@jiandan100.cn
 * @version  1.0.0
 * @date     {{.Date}}
 * @desc     go.jd100.com - {{.Module}}-{{.Name}} - models
 **------------------------------------------------------------**/
package model

// Medusa hello jiny.
type Medusa struct {
	Hello string
}`
	_tplGoMod = `module go.jd100.com/{{.Module}}/{{.Name}}

go 1.13.4

require (
	github.com/BurntSushi/toml v0.3.1
    github.com/stretchr/testify v1.4.0

	go.jd100.com/medusa v1.0.0-incompatible
)

replace go.jd100.com/medusa => ../medusa
`
	_tplGRPCServer = `/**------------------------------------------------------------**
 * @filename grpc/service.go
 * @author   {{.Owner}} - {{.Owner}}@jiandan100.cn
 * @version  1.0.0
 * @date     {{.Date}}
 * @desc     go.jd100.com - {{.Module}}-{{.Name}} - grpc server
 **------------------------------------------------------------**/
package grpc

import (
	pb "go.jd100.com/{{.Module}}/{{.Name}}/api"
	"go.jd100.com/{{.Module}}/{{.Name}}/conf"
	"go.jd100.com/{{.Module}}/{{.Name}}/service"

	"go.jd100.com/medusa/net/rpc/warden"
)

type server struct {
	as *service.Service
}

// var _ pb.MemberServer = &server{}

// New new a grpc server.
func New(c *warden.ServerConfig, s *service.Service) (svr *warden.Server) {
	svr = warden.NewServer(c)
	// pb.RegisterMemberServer(svr.Server(), &server{as: s})
	svr, err := svr.Start()
	if err != nil {
		panic(err)
	}
	return
}
`
	_tplGogen = `package api
// protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf/protobuf --gogo_out=plugins=grpc:. ./app/service/project_name/api/api.proto
//go:generate protoc --swagger --grpc --bm api.proto
`
)
