/**------------------------------------------------------------**
 * @filename log/log.go
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-07-30 17:22
 * @desc     go.jd100.com - log - logger
 **------------------------------------------------------------**/
package log

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"go.jd100.com/medusa/log/core"
	"go.jd100.com/medusa/stat/prom"
)

const (
	_defaultLogAppName   = "app_log"
	_defaultFilterString = "***"
)

var (
	enc core.Encoder
	ent core.Entry
	// errProm prometheus error counter.
	errProm = prom.BusinessErrCount
)

func init() {
	cfg = InitConf(nil, _defaultLogAppName)
	enc = core.NewJSONEncoder(core.EncoderConfig{
		EncodeTime:     core.ISO8601TimeEncoder,
		EncodeDuration: core.StringDurationEncoder,
		EncodeCaller:   core.ShortCallerEncoder,
	})

}

func Init(c *Config, app string) {
	cfg = InitConf(c, app)
}

func Debug(msg string, fields ...Field) {
	buildOut(context.Background(), DebugLevel, msg, fields...)
}

func Info(msg string, fields ...Field) {
	buildOut(context.Background(), InfoLevel, msg, fields...)
}

func Infof(template string, args ...interface{}) {
	buildOut(context.Background(), InfoLevel, sprint(template, args...))
}

func Infow(ctx context.Context, msg string, fields ...Field) {
	buildOut(ctx, InfoLevel, msg, fields...)
}

func Warn(msg string, fields ...Field) {
	buildOut(context.Background(), WarnLevel, msg, fields...)
}

func Warnf(template string, args ...interface{}) {
	buildOut(context.Background(), WarnLevel, sprint(template, args...))
}

func Warnw(ctx context.Context, msg string, fields ...Field) {
	buildOut(ctx, WarnLevel, msg, fields...)
}

func Error(msg string, fields ...Field) {
	buildOut(context.Background(), ErrorLevel, msg, fields...)
}

func Errorf(template string, args ...interface{}) {
	buildOut(context.Background(), ErrorLevel, sprint(template, args...))
}

func Errorw(ctx context.Context, msg string, fields ...Field) {
	buildOut(ctx, ErrorLevel, msg, fields...)
}

func Fatalf(template string, args ...interface{}) {
	buildOut(context.Background(), FatalLevel, sprint(template, args...))
}

func Sync() {

}

func buildOut(ctx context.Context, level core.Level, msg string, fields ...Field) {
	if NewAtomicLevelAt(cfg.Level).Enabled(level) {
		ent.Level = level
		preFields := []Field{String(_level, level.String()), String(_log, msg), Time(_time, time.Now())}
		if ent.IsCaller {
			ent.Caller = core.NewEntryCaller(runtime.Caller(2))
			preFields = append(preFields, String("_caller", ent.Caller.TrimmedPath()))
		}
		preFields = addExtraFields(ctx, preFields)
		co := core.NewCore(enc, level, os.Stderr)
		ce := co.Check(ent, nil)
		if ce != nil {
			preFields = append(preFields, fields...)
			outFields := make([]Field, 0)
			for _, tf := range preFields {
				var field = tf
				for _, filter := range cfg.Filters {
					if field.Key == filter {
						field.Type = core.StringType
						field.String = _defaultFilterString
						break
					}
				}
				outFields = append(outFields, field)
			}
			ce.Write(outFields...)
		}

	}
}

func sprint(template string, args ...interface{}) (message string) {
	message = template
	argsLen := len(args)
	if message == "" && argsLen > 0 {
		message = fmt.Sprint(args...)
	} else if message != "" && argsLen > 0 {
		message = fmt.Sprintf(template, args...)
	}
	return
}

func errIncr(lv core.Level, source string) {
	if lv == ErrorLevel {
		errProm.Incr(source)
	}
}
