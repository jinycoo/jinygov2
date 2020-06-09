/**------------------------------------------------------------**
 * @filename core/level.go
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-07-26 11:33
 * @desc     go.jd100.com - core - level
 **------------------------------------------------------------**/
package core

import (
	"fmt"

	"go.jd100.com/medusa/errors"
)

var errUnmarshalNilLevel = errors.New("can't unmarshal a nil *Level")

type Level int8

const (
	DebugLevel Level = iota - 1
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel

	_minLevel = DebugLevel
	_maxLevel = FatalLevel
)

func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	default:
		return fmt.Sprintf("Level(%d)", l)
	}
}

// Enabled returns true if the given level is at or above this level.
func (l Level) Enabled(lvl Level) bool {
	return lvl >= l
}

func (l Level) AddColor(body []byte) []byte {
	c := _levelToColor[l]
	result := []byte(fmt.Sprintf("\x1b[%dm", uint8(c)))
	result = append(result, body...)
	result = append(result, []byte("\x1b[0m")...)
	return result
}

// LevelEnabler decides whether a given logging level is enabled when logging a
// message.
type LevelEnabler interface {
	Enabled(Level) bool
}
