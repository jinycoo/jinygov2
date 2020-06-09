/**------------------------------------------------------------**
 * @filename redis/commands.go
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-08-02 16:33
 * @desc     go.jd100.com - redis - commands
 **------------------------------------------------------------**/
package redis

import "time"

type cmdable func(cmd Cmder) error
type statefulCmdable func(cmd Cmder) error

// Z represents sorted set member.
type Z struct {
	Score  float64
	Member interface{}
}

// ZWithKey represents sorted set member including the name of the key where it was popped.
type ZWithKey struct {
	Z
	Key string
}

type StatefulCmdable interface {
	Auth(password string) *StatusCmd
	Select(index int) *StatusCmd
	SwapDB(index1, index2 int) *StatusCmd
	ClientSetName(name string) *BoolCmd
}

func (c statefulCmdable) Auth(password string) *StatusCmd {
	cmd := NewStatusCmd("auth", password)
	_ = c(cmd)
	return cmd
}

func (c cmdable) Echo(message interface{}) *StringCmd {
	cmd := NewStringCmd("echo", message)
	_ = c(cmd)
	return cmd
}

func (c cmdable) Ping() *StatusCmd {
	cmd := NewStatusCmd("ping")
	_ = c(cmd)
	return cmd
}

func (c cmdable) Wait(numSlaves int, timeout time.Duration) *IntCmd {
	cmd := NewIntCmd("wait", numSlaves, int(timeout/time.Millisecond))
	_ = c(cmd)
	return cmd
}

func (c cmdable) Quit() *StatusCmd {
	panic("not implemented")
}

func (c statefulCmdable) Select(index int) *StatusCmd {
	cmd := NewStatusCmd("select", index)
	_ = c(cmd)
	return cmd
}

func (c statefulCmdable) SwapDB(index1, index2 int) *StatusCmd {
	cmd := NewStatusCmd("swapdb", index1, index2)
	_ = c(cmd)
	return cmd
}

// ClientSetName assigns a name to the connection.
func (c statefulCmdable) ClientSetName(name string) *BoolCmd {
	cmd := NewBoolCmd("client", "setname", name)
	_ = c(cmd)
	return cmd
}
