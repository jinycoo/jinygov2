/**------------------------------------------------------------**
 * @filename ctime/
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-07-15 14:06
 * @desc     go.jd100.com - ctime -
 **------------------------------------------------------------**/
package ctime

import (
	"context"
	"fmt"
	"time"
)

const (
	PRCDiff        = 28800
	LayoutDate     = "2006-01-02"
	LayoutDatetime = "2006-01-02 15:04:05"
)

type Time     int64
type JsonTime time.Time
type Duration time.Duration

func (j JsonTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(j).In(PRCLocal()).Format(LayoutDatetime)+`"`), nil
}

func (j JsonTime) In() JsonTime {
	return JsonTime(time.Time(j).In(PRCLocal()))
}

func (j JsonTime) Unix() int64 {
	return time.Time(j).In(PRCLocal()).Unix()
}

func (j JsonTime) String(layout string) string {
	return time.Time(j).Format(layout)
}

func (d *Duration) UnmarshalText(text []byte) error {
	tmp, err := time.ParseDuration(string(text))
	if err == nil {
		*d = Duration(tmp)
	}
	return err
}

// Shrink will decrease the duration by comparing with context's timeout duration
// and return new timeout\context\CancelFunc.
func (d Duration) Shrink(c context.Context) (Duration, context.Context, context.CancelFunc) {
	if deadline, ok := c.Deadline(); ok {
		if ctimeout := time.Until(deadline); ctimeout < time.Duration(d) {
			// deliver small timeout
			return Duration(ctimeout), c, func() {}
		}
	}
	ctx, cancel := context.WithTimeout(c, time.Duration(d))
	return d, ctx, cancel
}

func PRCLocal() *time.Location {
	return time.FixedZone("CST", PRCDiff)
}

func TimeToMillis(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

func DiffMilli(d time.Duration) int64 {
	return d.Nanoseconds() / int64(time.Millisecond)
}

func Now() int64 {
	return time.Now().In(PRCLocal()).Unix()
}

func Today() time.Time {
	td, _ := time.ParseInLocation(LayoutDate, time.Now().Format(LayoutDate), PRCLocal())
	return td
}

type ClassTime string

func (ct ClassTime) SubTime(index int8) string {
	var ctime = []byte(ct)
	if len(ctime) != 8 {
		return ""
	}
	return string(ctime[:index])
}

func (ct ClassTime) Abbr() string {
	return ct.SubTime(5)
}

func (ct ClassTime) ParseDateFormat(date JsonTime) string {
	return fmt.Sprintf("%s %s", date.String(LayoutDate), ct)
}

func (ct ClassTime) ParseDateUnix(date JsonTime) int64 {
	dt, err := time.ParseInLocation(LayoutDatetime, ct.ParseDateFormat(date), PRCLocal())
	if err != nil {
		return 0
	}
	return dt.Unix()
}
