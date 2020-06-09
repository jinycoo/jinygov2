/**------------------------------------------------------------**
 * @filename core/
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-07-30 17:58
 * @desc     go.jd100.com - core -
 **------------------------------------------------------------**/
package core

import (
	"io"
	"sync"
)

type WriteSyncer interface {
	io.Writer
	Sync() error
}

type ioCore struct {
	LevelEnabler

	enc Encoder
	out WriteSyncer
}

func NewCore(enc Encoder, lvl Level, out WriteSyncer) *ioCore {
	return &ioCore{
		lvl,
		enc,
		out,
	}
}

func (c *ioCore) Check(ent Entry, ce *CheckedEntry) *CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, ent.Level.String(), c)
	}
	return ce
}

func (c *ioCore) Write(ent Entry, fields []Field) error {
	buf, err := c.enc.EncodeEntry(ent, fields)
	if err != nil {
		return err
	}
	if ent.IsColor {
		_, err = c.out.Write(ent.Level.AddColor(buf.Bytes()))
	} else {
		_, err = c.out.Write(buf.Bytes())
	}

	buf.Free()
	if err != nil {
		return err
	}
	if ent.Level > ErrorLevel {
		c.Sync()
	}
	return nil
}

func (c *ioCore) Sync() error {
	return c.out.Sync()
}

func (c *ioCore) clone() *ioCore {
	return &ioCore{
		LevelEnabler: c.LevelEnabler,
		enc:          c.enc.Clone(),
		out:          c.out,
	}
}

type lockedWriteSyncer struct {
	sync.Mutex
	ws WriteSyncer
}

func Lock(ws WriteSyncer) WriteSyncer {
	if _, ok := ws.(*lockedWriteSyncer); ok {
		return ws
	}
	return &lockedWriteSyncer{ws: ws}
}

func (s *lockedWriteSyncer) Write(bs []byte) (int, error) {
	s.Lock()
	n, err := s.ws.Write(bs)
	s.Unlock()
	return n, err
}

func (s *lockedWriteSyncer) Sync() error {
	s.Lock()
	err := s.ws.Sync()
	s.Unlock()
	return err
}
