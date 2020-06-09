/**------------------------------------------------------------**
 * @filename jiny/middleware.go
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-07-23 18:01
 * @desc     go.jd100.com - jiny - middleware
 **------------------------------------------------------------**/
package jiny

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime"
	"strconv"
	"time"

	"go.jd100.com/medusa/ctime"
	"go.jd100.com/medusa/errors"
	"go.jd100.com/medusa/log"
	"go.jd100.com/medusa/net/metadata"
	"go.jd100.com/medusa/net/trace"
	//"go.jd100.com/medusa/utils/json"
)

const _defaultComponentName = "net/http"

type LogFormatterParams struct {
	// StatusCode is HTTP response code.
	StatusCode int `json:"http_code"`
	// ClientIP equals Context's ClientIP method.
	IP string `json:"ip"`
	// Method is the HTTP method given to the request.
	Method string `json:"method"`
	// Path is a path the client requests.
	Path string `json:"path"`
	// ErrorMessage is set if error has occurred in processing the request.
	ErrorMessage string `json:"error_msg"`
	// Keys are the keys set on the request's context.
	Keys map[string]interface{} `json:"keys"`
	Params string `json:"params"`
	TimeoutQuota float64 `json:"timeout_quota"`
}


func Logger() HandlerFn {
	return func(c *Context) {
		// Start timer
		start := time.Now()
		req := c.Request
		ip := metadata.String(c, metadata.RemoteIP)
		if ip == "" {
			ip = c.ClientIP()
		}
		path := req.URL.Path
		raw := req.URL.RawQuery
		params := req.Form
		var quota float64
		if deadline, ok := c.Context.Deadline(); ok {
			quota = time.Until(deadline).Seconds()
		}

		c.Next()

		logf := log.Info
		delay := time.Now().Sub(start)
		latency := ctime.DiffMilli(delay)
		if latency >= 500 {
			logf = log.Warn
		}

		param := LogFormatterParams {
			Keys:     c.Keys,
			IP:       ip,
			Method:   c.Request.Method,
			StatusCode: c.Writer.Status(),
			ErrorMessage: errors.ECause(c.Error).Message(), // c.Errors.ByType(ErrorTypePrivate).String(),
			Params: params.Encode(),
			TimeoutQuota: quota,
		}
		if raw != "" {
			path = path + "?" + raw
		}
		param.Path = path
		msg, _ := json.Marshal(&param)
		logf(string(msg), log.String("mod", "jiny"), log.Duration("latency", delay))
	}
}

func Recovery() HandlerFn {
	return func(c *Context) {
		defer func() {
			var rawReq []byte
			if err := recover(); err != nil {
				const size = 64 << 10
				buf := make([]byte, size)
				buf = buf[:runtime.Stack(buf, false)]
				if c.Request != nil {
					rawReq, _ = httputil.DumpRequest(c.Request, false)
				}
				pl := fmt.Sprintf("[Recovery] http call panic: %s\n%v\n%s\n", string(rawReq), err, buf)
				_, err = fmt.Fprintf(os.Stderr, pl)
				log.Error(pl)
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}

func Trace() HandlerFn {
	return func(c *Context) {
		t, err := trace.Extract(trace.HTTPFormat, c.Request.Header)
		if err != nil {
			var opts []trace.Option
			if ok, _ := strconv.ParseBool(trace.ETTraceDebug); ok {
				opts = append(opts, trace.EnableDebug())
			}
			t = trace.New(c.Request.URL.Path, opts...)
		}
		t.SetTitle(c.Request.URL.Path)
		t.SetTag(trace.String(trace.TagComponent, _defaultComponentName))
		t.SetTag(trace.String(trace.TagHTTPMethod, c.Request.Method))
		t.SetTag(trace.String(trace.TagHTTPURL, c.Request.URL.String()))
		t.SetTag(trace.String(trace.TagSpanKind, "server"))
		t.SetTag(trace.String("caller", metadata.String(c.Context, metadata.Caller)))
		c.Context = trace.NewContext(c.Context, t)
		c.Next()
		t.Finish(&c.Error)
	}
}

func JsonHandler404() HandlerFn {
	return func(c *Context) {
		err := errors.NothingFound
		c.JSON(http.StatusNotFound, err)
	}
}