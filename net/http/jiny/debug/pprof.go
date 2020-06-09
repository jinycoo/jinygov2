/**------------------------------------------------------------**
 * @filename debug/pprof.go
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-07-25 13:10
 * @desc     go.jd100.com - debug - pprof 性能优化
 **------------------------------------------------------------**/
package debug

import (
	"flag"
	"net/http"
	"net/http/pprof"
	"os"
	"sync"

	"go.jd100.com/medusa/config/dsn"
	"go.jd100.com/medusa/errors"
)

var (
	pprofOnce sync.Once
	debugDSN  string
)

func init() {
	v := os.Getenv("HTTP_PPROF")
	if v == "" {
		v = "tcp://0.0.0.0:6066"
	}
	flag.StringVar(&debugDSN, "http.pprof", v, "listen http perf dsn, or use HTTP_PPROF env variable.")
}

func StartPprof() {
	pprofOnce.Do(func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)

		go func() {
			d, err := dsn.Parse(debugDSN)
			if err != nil {
				panic(errors.Errorf("jiny: http perf dsn must be tcp://$host:port, %s:error(%v)", debugDSN, err))
			}
			if err := http.ListenAndServe(d.Host, mux); err != nil {
				panic(errors.Errorf("jiny: listen %s: error(%v)", d.Host, err))
			}
		}()
	})
}