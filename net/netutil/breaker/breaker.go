/**------------------------------------------------------------**
 * @filename breaker/breaker.go
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-07-31 14:47
 * @desc     go.jd100.com - breaker - 熔断器
 **------------------------------------------------------------**/
package breaker

import (
	"sync"
	"time"

	"go.jd100.com/medusa/ctime"
)

const (
	// StateOpen when circuit breaker open, request not allowed, after sleep
	// some duration, allow one single request for testing the health, if ok
	// then state reset to closed, if not continue the step.
	StateOpen int32 = iota
	// StateClosed when circuit breaker closed, request allowed, the breaker
	// calc the succeed ratio, if request num greater request setting and
	// ratio lower than the setting ratio, then reset state to open.
	StateClosed
	// StateHalfOpen when circuit breaker open, after slepp some duration, allow
	// one request, but not state closed.
	StateHalfOpen

	//_switchOn int32 = iota
	// _switchOff
)

// Breaker is a CircuitBreaker pattern.
type Breaker interface {
	Allow() error
	MarkSuccess()
	MarkFailed()
	ReportResult(error)
}

// Group represents a class of CircuitBreaker and forms a namespace in which
// units of CircuitBreaker.
type Group struct {
	mu   sync.RWMutex
	brks map[string]Breaker
	conf *Config
}

var (
	_mu   sync.RWMutex
	_conf = &Config{
		Window:  ctime.Duration(3 * time.Second),
		Bucket:  10,
		Request: 100,

		Sleep: ctime.Duration(500 * time.Millisecond),
		Ratio: 0.5,
		// Percentage of failures must be lower than 33.33%
		K: 1.5,

		// Pattern: "",
	}
	_group = NewGroup(_conf)
)

// Init init global breaker config, also can reload config after first time call.
func Init(conf *Config) {
	if conf == nil {
		return
	}
	_mu.Lock()
	_conf = conf
	_mu.Unlock()
}

// Go runs your function while tracking the breaker state of default group.
func Go(name string, run, fallback func() error) error {
	breaker := _group.Get(name)
	if err := breaker.Allow(); err != nil {
		return fallback()
	}
	return run()
}

// newBreaker new a breaker.
func newBreaker(c *Config) (b Breaker) {
	// factory
	return newSRE(c)
}

// NewGroup new a breaker group container, if conf nil use default conf.
func NewGroup(conf *Config) *Group {
	if conf == nil {
		_mu.RLock()
		conf = _conf
		_mu.RUnlock()
	} else {
		conf.fix()
	}
	return &Group{
		conf: conf,
		brks: make(map[string]Breaker),
	}
}

// Get get a breaker by a specified key, if breaker not exists then make a new one.
func (g *Group) Get(key string) Breaker {
	g.mu.RLock()
	brk, ok := g.brks[key]
	conf := g.conf
	g.mu.RUnlock()
	if ok {
		return brk
	}
	// NOTE here may new multi breaker for rarely case, let gc drop it.
	brk = newBreaker(conf)
	g.mu.Lock()
	if _, ok = g.brks[key]; !ok {
		g.brks[key] = brk
	}
	g.mu.Unlock()
	return brk
}

// Reload reload the group by specified config, this may let all inner breaker
// reset to a new one.
func (g *Group) Reload(conf *Config) {
	if conf == nil {
		return
	}
	conf.fix()
	g.mu.Lock()
	g.conf = conf
	g.brks = make(map[string]Breaker, len(g.brks))
	g.mu.Unlock()
}

// Go runs your function while tracking the breaker state of group.
func (g *Group) Go(name string, run, fallback func() error) error {
	breaker := g.Get(name)
	if err := breaker.Allow(); err != nil {
		return fallback()
	}
	return run()
}
