/**------------------------------------------------------------**
 * @filename log/config.go
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-07-30 17:23
 * @desc     go.jd100.com - log - config
 **------------------------------------------------------------**/
package log

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"

	"go.jd100.com/medusa/config/env"
)

var cfg *Config

type (
	logFilter     []string
	verboseModule map[string]int32
)

type Config struct {
	Family string
	Host   string

	Level  string

	Std    bool
	File   bool
	Dir    string

	Agent interface{}
	Modules map[string]string
	Filters []string
}

func InitConf(conf *Config, app string) *Config {
	if conf == nil {
		conf = new(Config)
		conf.Std = true
		conf.Level = _defaultLevelS
		conf.Family = app
	}
	_, exist := _levels[conf.Level]
	switch env.DeployEnv {
	case "", env.DeployEnvDev:
		if !exist {
			conf.Level = _defaultLevelS
		}
		if conf.Std {
			ent.IsColor = true
		}
		ent.IsCaller = true
	case env.DeployEnvUat:
		if !exist {
			conf.Level = _defaultUatLevelS
		}
		ent.IsColor = false
		ent.IsCaller = true
	default:
		if !exist {
			conf.Level = _defaultPreLevelS
		}
	}

	if len(conf.Family) == 0 {
		if len(app) > 0 {
			conf.Family = app
		}
	}

	if len(env.AppID) != 0 {
		conf.Family = env.AppID // for caster
	}

	if len(conf.Host) == 0 {
		conf.Host = env.Hostname
		if len(conf.Host) == 0 {
			host, _ := os.Hostname()
			conf.Host = host
		}
	}
	return conf
}

func (f *logFilter) String() string {
	return fmt.Sprint(*f)
}

// Set sets the value of the named command-line flag.
// format: -log.filter key1,key2
func (f *logFilter) Set(value string) error {
	for _, i := range strings.Split(value, ",") {
		*f = append(*f, strings.TrimSpace(i))
	}
	return nil
}

func (m verboseModule) String() string {
	// FIXME strings.Builder
	var buf bytes.Buffer
	for k, v := range m {
		buf.WriteString(k)
		buf.WriteString(strconv.FormatInt(int64(v), 10))
		buf.WriteString(",")
	}
	return buf.String()
}

// Set sets the value of the named command-line flag.
// format: -log.module file=1,file2=2
func (m verboseModule) Set(value string) error {
	for _, i := range strings.Split(value, ",") {
		kv := strings.Split(i, "=")
		if len(kv) == 2 {
			if v, err := strconv.ParseInt(kv[1], 10, 64); err == nil {
				m[strings.TrimSpace(kv[0])] = int32(v)
			}
		}
	}
	return nil
}
