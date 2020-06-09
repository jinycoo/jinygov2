/**------------------------------------------------------------**
 * @filename core/entry.go
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-07-26 11:48
 * @desc     go.jd100.com - core - entry
 **------------------------------------------------------------**/
package core

import (
	"strings"
	"sync"

	"go.jd100.com/medusa/log/buffer"
)

var (
	_cePool = sync.Pool{New: func() interface{} {
		// Pre-allocate some space for cores.
		return &CheckedEntry{
			cores: make(map[string]*ioCore, 4),
		}
	}}
)

func getCheckedEntry() *CheckedEntry {
	ce := _cePool.Get().(*CheckedEntry)
	ce.reset()
	return ce
}

func putCheckedEntry(ce *CheckedEntry) {
	if ce == nil {
		return
	}
	_cePool.Put(ce)
}

type EntryCaller struct {
	Defined bool
	PC      uintptr
	File    string
	Line    int
}

func NewEntryCaller(pc uintptr, file string, line int, ok bool) EntryCaller {
	if !ok {
		return EntryCaller{}
	}
	return EntryCaller{
		PC:      pc,
		File:    file,
		Line:    line,
		Defined: true,
	}
}

func (ec EntryCaller) String() string {
	return ec.FullPath()
}

// FullPath returns a /full/path/to/package/file:line description of the
// caller.
func (ec EntryCaller) FullPath() string {
	if !ec.Defined {
		return "undefined"
	}
	buf := buffer.Get()
	buf.AppendString(ec.File)
	buf.AppendByte(':')
	buf.AppendInt(int64(ec.Line))
	caller := buf.String()
	buf.Free()
	return caller
}

// TrimmedPath returns a package/file:line description of the caller,
// preserving only the leaf directory name and file name.
func (ec EntryCaller) TrimmedPath() string {
	if !ec.Defined {
		return "undefined"
	}
	idx := strings.LastIndexByte(ec.File, '/')
	if idx == -1 {
		return ec.FullPath()
	}
	// Find the penultimate separator.
	idx = strings.LastIndexByte(ec.File[:idx], '/')
	if idx == -1 {
		return ec.FullPath()
	}
	buf := buffer.Get()
	// Keep everything after the penultimate separator.
	buf.AppendString(ec.File[idx+1:])
	buf.AppendByte(':')
	buf.AppendInt(int64(ec.Line))
	caller := buf.String()
	buf.Free()
	return caller
}

type Entry struct {
	Level      Level
	Caller     EntryCaller
	IsColor    bool
	IsCaller   bool
}

type CheckedEntry struct {
	Entry
	cores       map[string]*ioCore
}

func (ce *CheckedEntry) reset() {
	ce.Entry = Entry{}
	for i := range ce.cores {
		// don't keep references to cores
		ce.cores[i] = nil
	}
	ce.cores = make(map[string]*ioCore)
}

func (ce *CheckedEntry) Write(fields ...Field) {
	if ce == nil {
		return
	}
	for _, co := range ce.cores {
		co.Write(ce.Entry, fields)
	}
	putCheckedEntry(ce)
}

func (ce *CheckedEntry) AddCore(ent Entry, ck string, core *ioCore) *CheckedEntry {
	if ce == nil {
		ce = getCheckedEntry()
		ce.Entry = ent
	}
	ce.cores[ck] = core
	return ce
}
