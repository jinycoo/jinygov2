/**------------------------------------------------------------**
 * @filename pool/pool_single.go
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-08-02 17:20
 * @desc     go.jd100.com - pool - single pool
 **------------------------------------------------------------**/
package pool

import "context"

type SingleConnPool struct {
	cn *Conn
}

var _ Pooler = (*SingleConnPool)(nil)

func NewSingleConnPool(cn *Conn) *SingleConnPool {
	return &SingleConnPool{
		cn: cn,
	}
}

func (p *SingleConnPool) NewConn(context.Context) (*Conn, error) {
	panic("not implemented")
}

func (p *SingleConnPool) CloseConn(*Conn) error {
	panic("not implemented")
}

func (p *SingleConnPool) Get(ctx context.Context) (*Conn, error) {
	return p.cn, nil
}

func (p *SingleConnPool) Put(cn *Conn) {
	if p.cn != cn {
		panic("p.cn != cn")
	}
}

func (p *SingleConnPool) Remove(cn *Conn) {
	if p.cn != cn {
		panic("p.cn != cn")
	}
}

func (p *SingleConnPool) Len() int {
	return 1
}

func (p *SingleConnPool) IdleLen() int {
	return 0
}

func (p *SingleConnPool) Stats() *Stats {
	return nil
}

func (p *SingleConnPool) Close() error {
	return nil
}
