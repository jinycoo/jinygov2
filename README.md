### 简单科技 - Golang 项目代码库 Named：Medusa

#### 快速开始

###### 要求
Go version>=1.13

###### 安装

```sh
    比如  https://dl.google.com/go/go1.13.4.darwin-amd64.tar.gz  macOS
    wget https://dl.google.com/go/go$VERSION.$OS-$ARCH.tar.gz
    tar -C /usr/local -xzf go$VERSION.$OS-$ARCH.tar.gz

    export GOPATH=$HOME/go_path
    export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin

    或者 vim $HOME/.profile  
```

`jiny new project -o author -m module`

-o author   项目创建人

-m interface  模块名称  比如 后台=admin  接口=interface  服务=service

-p path   项目路经

```sh
  $ mkdir go.jd100.com   // 创建工作目录
  $ cd    go.jd100.com
  $ git clone http://gitlab.jiandan100.cn/webdev/medusa.git
  $ cd medusa/tool/jiny && go build -o $GOPATH/bin/jiny main.go
  $ jiny new project_demo -o author -m interface -p xxx/go.jd100.com 
  or $ cd ../../.. && jiny new project_demo -o author -m interface
  $ cd project_demo
  $ go mod tidy

  // 注 如果没翻墙，可以先执行下面的命令，来修改默认https://proxy.golang.org/
  // go 版本建议使用 最新 1.13.3
  // 默认代理有时可用，有时不可用，需注意
  go env -w GOPROXY=https://goproxy.cn,direct 

```

##### 项目规范

1 每个目录 需要有独立的 
  - README.md
  - CHANGELOG.md
  - CONTRIBUTORS.md
  
2 以后每个业务或者基础组件维护自己的版本号，在CHANGELOG.md中，构建以后的tag关联成自己的版本号

3 提供RPC内部服务 -m service，任务队列 -m job，对外网关服务 -m interface，管理后台服务 -m admin
  例如： jiny new member -m service   生成module = go.jd100.com/service/member 

4 每个项目的配置文件 app.toml 在bin文件夹中,测试运行时，最好build -o ../bin/demo main.go 或者 demo -conf xxx/bin/app.toml指定配置文件 

5 实例可参考 http://gitlab.jiandan100.cn/webdev/dean/tree/master/app/admin/dean

6 API 文档部分将使用生成swagger格式，可通过YAPI接口导入到公司YAPI平台上，减少程序员工作量

#### 代码结构

<details>
<summary>展开查看</summary>
<pre><code>.
   go.jd100.com       // 工作目录
   ├── project_name   // 项目目录
   │   ├── bin        // 配置目录 运行时请将build目录指定为该目录
   │   │   └── app.toml
   │   ├── conf  
   │   │   └── conf.go
   │   ├── cmd
   │   │   └── main.go // 程序入口
   │   ├── dao
   │   │   ├── mysql.go  // sql script 
   │   │   ├── redis.go  // cache redis key ...
   │   │   └── dao.go
   │   ├── install    // K8S  安装部署目录
   │   │   ├── build
   │   │   │   └── ....
   │   │   └── helm
   │   │       └── project_name
   │   ├── model
   │   │   └── model.go
   │   ├── server   // 服务 默认为http服务
   │   │   ├── grpc // GRPC 服务
   │   │   │   └── server.go
   │   │   └── http // http server
   │   │       └── server.go
   │   └── service
   │       └── service.go
   └── medusa
       ├── cache
       │   └── redis
       │       └── Owner: jinycoo
       ├── container
       │   └── pool
       │       └── Owner: 
       ├── database
       │   └── sql
       │       └── Owner: 
       ├── errors
       │   ├── Owner: jinycoo
       │   └── tip
       │       └── Owner: all
       ├── log
       │   └── Owner: jinycoo
       ├── naming
       │   └── discovery
       │       └── Owner: 
       ├── net
       │   ├── http
       │   │   ├── Owner: 
       │   │   └── jiny
       │   │       ├── Owner: 
       │   │       ├── middleware
       │   │       │   ├── Owner: 
       │   │       │   ├── antispam
       │   │       │   │   └── Owner: 
       │   │       │   ├── auth
       │   │       │   │   └── Owner: 
       │   │       │   ├── cache
       │   │       │   │   └── Owner: 
       │   │       │   ├── identify
       │   │       │   │   └── Owner: 
       │   │       │   ├── limit
       │   │       │   │   └── aqm
       │   │       │   │       └── Owner: 
       │   │       │   ├── proxy
       │   │       │   │   └── Owner: 
       │   │       │   ├── rate
       │   │       │   │   └── Owner: 
       │   │       │   ├── supervisor
       │   │       │   │   └── Owner: 
       │   │       │   ├── tag
       │   │       │   │   └── Owner: 
       │   │       │   └── verify
       │   │       │       └── Owner: 
       │   │       └── render
       │   │           └── Owner: 
       │   ├── metadata
       │   │   └── Owner: 
       │   ├── netutil
       │   │   └── breaker
       │   │       └── Owner: 
       │   ├── rpc
       │   │   └── warden
       │   │       ├── Owner: 
       │   │       ├── balancer
       │   │       │   └── wrr
       │   │       │       └── Owner: 
       │   │       └── resolver
       │   │           └── Owner: 
       │   └── trace
       │       └── Owner: 
       ├── rate
       │   └── limit
       │       └── bench
       │           └── stress
       │               └── Owner: 
       ├── stat
       │   └── sys
       │       └── cpu
       │           └── Owner: 
       │── tool
       │   └── jiny
       │       └── Owner:  
       └── sync
           └── errgroup
               └── Owner: 
</code></pre>
</details>

> Dev 发布部署相关  

- GRPC相关

[ProtoBuf](https://github.com/protocolbuffers/protobuf/releases)