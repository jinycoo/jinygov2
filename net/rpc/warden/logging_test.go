package warden

import (
	"context"
	"reflect"
	"testing"
	"time"

	"go.jd100.com/medusa/log"
)

func Test_logFn(t *testing.T) {
	type args struct {
		code int
		dt   time.Duration
	}
	tests := []struct {
		name string
		args args
		want func(context.Context, string, ...log.Field)
	}{
		{
			name: "ok",
			args: args{code: 0, dt: time.Millisecond},
			want: log.Infow,
		},
		{
			name: "slowlog",
			args: args{code: 0, dt: time.Second},
			want: log.Warnw,
		},
		{
			name: "business error",
			args: args{code: 2233, dt: time.Millisecond},
			want: log.Warnw,
		},
		{
			name: "system error",
			args: args{code: -1, dt: 0},
			want: log.Errorw,
		},
		{
			name: "system error and slowlog",
			args: args{code: -1, dt: time.Second},
			want: log.Errorw,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := logFn(tt.args.code, tt.args.dt); reflect.ValueOf(got).Pointer() != reflect.ValueOf(tt.want).Pointer() {
				t.Errorf("unexpect log function!")
			}
		})
	}
}
