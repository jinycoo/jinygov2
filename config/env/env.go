/**------------------------------------------------------------**
 * @filename env/env.go
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-07-15 13:14
 * @desc     go.jd100.com - env - environment params
 **------------------------------------------------------------**/
package env

import (
	"flag"
	"os"
)

const (
	// deploy env.
	DeployEnvDev  = "dev"  // 开发
	DeployEnvUat  = "uat"  // 测试
	DeployEnvPre  = "pre"  // 预发布
	DeployEnvProd = "pro"  // 生产

	// env default setting
	_region    = "bj"
	_zone      = "jd100"
	_deployEnv = "dev"

	// app default port
	_httpPort   = "8088"
	_gRpcPort   = "9000"
	_thriftPort = "9200"
)

var (
	// Region avaliable region where app at.
	Region string
	// Zone avaliable zone where app at.
	Zone string
	// Hostname machine hostname.
	Hostname string
	// DeployEnv deploy env where app at.
	DeployEnv string
	// IP
	IP = os.Getenv("POD_IP")
	// AppID is global unique application id, register by service tree.
	// such as main.arch.disocvery.
	AppID string
	// Color is the identification of different experimental group in one caster cluster.
	Color string


	// HTTPPort app listen http port.
	HTTPPort string
	// GRPCPort app listen gRpc port.
	GRPCPort string
	// Thrift app listen thrift port.
	ThriftPort string
)

func init() {
	var err error
	if Hostname, err = os.Hostname(); err != nil || Hostname == "" {
		Hostname = os.Getenv("HOSTNAME")
	}

	addFlag(flag.CommandLine)
}

func addFlag(fs *flag.FlagSet) {
	// env
	fs.StringVar(&Region, "region", defaultString("REGION", _region), "available region. or use REGION env variable, value: sh etc.")
	fs.StringVar(&Zone, "zone", defaultString("ZONE", _zone), "available zone. or use ZONE env variable, value: sh001/sh002 etc.")
	fs.StringVar(&DeployEnv, "deploy.env", defaultString("DEPLOY_ENV", _deployEnv), "deploy env. or use DEPLOY_ENV env variable, value: dev/fat1/uat/pre/prod etc.")
	fs.StringVar(&AppID, "appid", os.Getenv("APP_ID"), "appId is global unique application id, register by service tree. or use APP_ID env variable.")
	fs.StringVar(&Color, "deploy.color", os.Getenv("DEPLOY_COLOR"), "deploy.color is the identification of different experimental group.")

	// app
	fs.StringVar(&HTTPPort, "http.port", defaultString("DISCOVERY_HTTP_PORT", _httpPort), "app listen http port, default: 8088")
	fs.StringVar(&GRPCPort, "grpc.port", defaultString("DISCOVERY_GRPC_PORT", _gRpcPort), "app listen gRpc port, default: 9000")
	fs.StringVar(&ThriftPort, "thrift.port", defaultString("DISCOVERY_THRIFT_PORT", _thriftPort), "app listen thrift port, default: 9200")
}

func defaultString(env, value string) string {
	v := os.Getenv(env)
	if v == "" {
		return value
	}
	return v
}