/**------------------------------------------------------------**
 * @filename redis/
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-08-02 17:03
 * @desc     go.jd100.com - redis -
 **------------------------------------------------------------**/
package redis

import (
	"context"
	"sync"

	"go.jd100.com/medusa/cache/redis/internal/pool"
)

type pipelineExecer func(context.Context, []Cmder) error

// Pipeliner is an mechanism to realise Redis Pipeline technique.
//
// Pipelining is a technique to extremely speed up processing by packing
// operations to batches, send them at once to Redis and read a replies in a
// singe step.
// See https://redis.io/topics/pipelining
//
// Pay attention, that Pipeline is not a transaction, so you can get unexpected
// results in case of big pipelines and small read/write timeouts.
// Redis client has retransmission logic in case of timeouts, pipeline
// can be retransmitted and commands can be executed more then once.
// To avoid this: it is good idea to use reasonable bigger read/write timeouts
// depends of your batch size and/or use TxPipeline.
type Pipeliner interface {
	StatefulCmdable
	Do(args ...interface{}) *Cmd
	Process(cmd Cmder) error
	Close() error
	Discard() error
	Exec() ([]Cmder, error)
	ExecContext(ctx context.Context) ([]Cmder, error)
}

var _ Pipeliner = (*Pipeline)(nil)

// Pipeline implements pipelining as described in
// http://redis.io/topics/pipelining. It's safe for concurrent use
// by multiple goroutines.
type Pipeline struct {
	cmdable
	statefulCmdable

	ctx  context.Context
	exec pipelineExecer

	mu     sync.Mutex
	cmds   []Cmder
	closed bool
}

func (p *Pipeline) init() {
	p.cmdable = p.Process
	p.statefulCmdable = p.Process
}

func (p *Pipeline) Do(args ...interface{}) *Cmd {
	cmd := NewCmd(args...)
	_ = p.Process(cmd)
	return cmd
}

// Process queues the cmd for later execution.
func (p *Pipeline) Process(cmd Cmder) error {
	p.mu.Lock()
	p.cmds = append(p.cmds, cmd)
	p.mu.Unlock()
	return nil
}

// Close closes the pipeline, releasing any open resources.
func (p *Pipeline) Close() error {
	p.mu.Lock()
	_ = p.discard()
	p.closed = true
	p.mu.Unlock()
	return nil
}

// Discard resets the pipeline and discards queued commands.
func (p *Pipeline) Discard() error {
	p.mu.Lock()
	err := p.discard()
	p.mu.Unlock()
	return err
}

func (p *Pipeline) discard() error {
	if p.closed {
		return pool.ErrClosed
	}
	p.cmds = p.cmds[:0]
	return nil
}

// Exec executes all previously queued commands using one
// client-server roundtrip.
//
// Exec always returns list of commands and error of the first failed
// command if any.
func (p *Pipeline) Exec() ([]Cmder, error) {
	return p.ExecContext(p.ctx)
}

func (p *Pipeline) ExecContext(ctx context.Context) ([]Cmder, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil, pool.ErrClosed
	}

	if len(p.cmds) == 0 {
		return nil, nil
	}

	cmds := p.cmds
	p.cmds = nil

	return cmds, p.exec(ctx, cmds)
}

func (p *Pipeline) Pipelined(fn func(Pipeliner) error) ([]Cmder, error) {
	if err := fn(p); err != nil {
		return nil, err
	}
	cmds, err := p.Exec()
	_ = p.Close()
	return cmds, err
}

func (p *Pipeline) Pipeline() Pipeliner {
	return p
}

