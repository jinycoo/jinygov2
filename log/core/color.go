/**------------------------------------------------------------**
 * @filename core/color
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-07-15 10:36
 * @desc     go.jd100.com - core - log color
 **------------------------------------------------------------**/
package core

const (
	Black Color = iota + 30
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
)

var (
	_levelToColor = map[Level]Color{
		DebugLevel: Magenta,
		InfoLevel:  Blue,
		WarnLevel:  Yellow,
		ErrorLevel: Red,
	}
	_unknownLevelColor = Red
)

type Color uint8
