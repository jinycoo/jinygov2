/**------------------------------------------------------------**
 * @filename log/agent.go
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-07-26 10:00
 * @desc     go.jd100.com - log - send log to container server
 **------------------------------------------------------------**/
package log

import (
	"sync"
	"time"

	"go.jd100.com/medusa/config/env"
	"go.jd100.com/medusa/ctime"
	"go.jd100.com/medusa/log/core"
)

const (
	_agentTimeout = ctime.Duration(20 * time.Millisecond)
	_mergeWait    = 1 * time.Second
	_maxBuffer    = 10 * 1024 * 1024 // 10mb
	_defaultChan  = 2048

	_defaultAgentConfig = "unixpacket:///var/run/lancer/collector_tcp.sock?timeout=100ms&chan=1024"
)

var (
	_logSeparator = []byte("\u0001")

	_defaultTaskIDs = map[string]string{
		env.DeployEnvUat:  "000069",
		env.DeployEnvPre:  "000161",
		env.DeployEnvProd: "000161",
	}
)

// AgentHandler agent struct.
type AgentHandler struct {
	c         *AgentConfig
	msgs      chan []core.Field
	waiter    sync.WaitGroup
	pool      sync.Pool
	enc       core.Encoder
	batchSend bool
	filters   map[string]struct{}
}

// AgentConfig agent config.
type AgentConfig struct {
	TaskID  string
	Buffer  int
	Proto   string         `dsn:"network"`
	Addr    string         `dsn:"address"`
	Chan    int            `dsn:"query.chan"`
	Timeout ctime.Duration `dsn:"query.timeout"`
}

// NewAgent a Agent.
func NewAgent(ac *AgentConfig) (a *AgentHandler) {
	if ac == nil {
		//ac = parseDSN(_agentDSN)
	}
	if len(ac.TaskID) == 0 {
		ac.TaskID = _defaultTaskIDs[env.DeployEnv]
	}
	a = &AgentHandler{
		c: ac,
		//enc: core.NewJSONEncoder(core.EncoderConfig{
		//	EncodeTime:     core.EpochTimeEncoder,
		//	EncodeDuration: core.SecondsDurationEncoder,
		//}, core.NewBuffer(0)),
	}
	a.pool.New = func() interface{} {
		return make([]core.Field, 0, 16)
	}
	if ac.Chan == 0 {
		ac.Chan = _defaultChan
	}
	a.msgs = make(chan []core.Field, ac.Chan)
	if ac.Timeout == 0 {
		ac.Timeout = _agentTimeout
	}
	if ac.Buffer == 0 {
		ac.Buffer = 100
	}
	a.waiter.Add(1)

	if a.c.Proto == "unixpacket" {
		a.batchSend = true
	}

	//go a.writeproc()
	return
}