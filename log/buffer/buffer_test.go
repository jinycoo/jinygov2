package buffer

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBufferWrites(t *testing.T) {
	buf := Get()

	tests := []struct {
		desc string
		f    func()
		want string
	}{
		{"AppendByte", func() { buf.AppendByte('v') }, "v"},
		{"AppendString", func() { buf.AppendString("foo") }, "foo"},
		{"AppendIntPositive", func() { buf.AppendInt(42) }, "42"},
		{"AppendIntNegative", func() { buf.AppendInt(-42) }, "-42"},
		{"AppendUint", func() { buf.AppendUint(42) }, "42"},
		{"AppendBool", func() { buf.AppendBool(true) }, "true"},
		{"AppendFloat64", func() { buf.AppendFloat(3.14, 64) }, "3.14"},
		// Intenationally introduce some floating-point error.
		{"AppendFloat32", func() { buf.AppendFloat(float64(float32(3.14)), 32) }, "3.14"},
		{"AppendWrite", func() { buf.Write([]byte("foo")) }, "foo"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			buf.Reset()
			tt.f()
			assert.Equal(t, tt.want, buf.String(), "Unexpected String().")
			assert.Equal(t, tt.want, string(buf.Bytes()), "Unexpected string(Bytes()).")
			assert.Equal(t, len(tt.want), buf.Len(), "Unexpected buffer length.")
			// We're not writing more than a kibibyte in tests.
			assert.Equal(t, _size, buf.Cap(), "Expected buffer capacity to remain constant.")
		})
	}
}

func BenchmarkBuffers(b *testing.B) {
	str := strings.Repeat("a", 1024)
	slice := make([]byte, 1024)
	buf := bytes.NewBuffer(slice)
	custom := Get()
	b.Run("ByteSlice", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			slice = append(slice, str...)
			slice = slice[:0]
		}
	})
	b.Run("BytesBuffer", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf.WriteString(str)
			buf.Reset()
		}
	})
	b.Run("CustomBuffer", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			custom.AppendString(str)
			custom.Reset()
		}
	})
}
