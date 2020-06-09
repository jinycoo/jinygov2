/**------------------------------------------------------------**
 * @filename log/level.go
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-07-29 14:04
 * @desc     go.jd100.com - log - level
 **------------------------------------------------------------**/
package log

import (
	"go.jd100.com/medusa/log/core"
)

const (
	DebugLevel = core.DebugLevel
	InfoLevel  = core.InfoLevel
	WarnLevel  = core.WarnLevel
	ErrorLevel = core.ErrorLevel
	FatalLevel = core.FatalLevel

	_defaultLevelS    = "info"
	_defaultUatLevelS = "warn"
	_defaultPreLevelS = "error"
)

var _levels = map[string]core.Level{
	"debug": DebugLevel,
	"info":  InfoLevel,
	"warn":  WarnLevel,
	"error": ErrorLevel,
	"fatal": FatalLevel,
}

// LevelEnablerFunc is a convenient way to implement core.LevelEnabler with
// an anonymous function.
//
// It's particularly useful when splitting log output between different
// outputs (e.g., standard error and standard out). For sample code, see the
// package-level AdvancedConfiguration example.
type LevelEnablerFunc func(core.Level) bool

// Enabled calls the wrapped function.
func (f LevelEnablerFunc) Enabled(lvl core.Level) bool { return f(lvl) }

// An AtomicLevel is an atomically changeable, dynamic logging level. It lets
// you safely change the log level of a tree of loggers (the root logger and
// any children created by adding context) at runtime.
//
// The AtomicLevel itself is an http.Handler that serves a JSON endpoint to
// alter its level.
//
// AtomicLevels must be created with the NewAtomicLevel constructor to allocate
// their internal atomic pointer.
type AtomicLevel struct {
	l *core.Int32
}

// NewAtomicLevel creates an AtomicLevel with InfoLevel and above logging
// enabled.
func NewAtomicLevel() AtomicLevel {
	return AtomicLevel{
		l: core.NewInt32(int32(InfoLevel)),
	}
}

// NewAtomicLevelAt is a convenience function that creates an AtomicLevel
// and then calls SetLevel with the given level.
func NewAtomicLevelAt(level string) AtomicLevel {
	a := NewAtomicLevel()
	if l, ok := _levels[level]; ok {
		a.SetLevel(l)
	}
	return a
}

// Enabled implements the core.LevelEnabler interface, which allows the
// AtomicLevel to be used in place of traditional static levels.
func (lvl AtomicLevel) Enabled(l core.Level) bool {
	return lvl.Level().Enabled(l)
}

// Level returns the minimum enabled log level.
func (lvl AtomicLevel) Level() core.Level {
	return core.Level(int8(lvl.l.Load()))
}

// SetLevel alters the logging level.
func (lvl AtomicLevel) SetLevel(l core.Level) {
	lvl.l.Store(int32(l))
}

// String returns the string representation of the underlying Level.
func (lvl AtomicLevel) String() string {
	return lvl.Level().String()
}
