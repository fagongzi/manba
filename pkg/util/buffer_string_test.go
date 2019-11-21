package util

import (
	"bytes"
	"github.com/fagongzi/util/hack"
	"testing"
)

func BenchmarkHackToString(b *testing.B) {
	var buf bytes.Buffer
	buf.WriteString("[")
	buf.Write([]byte("GET"))
	buf.WriteString("]")
	buf.Write([]byte("api"))
	for n := 0; n < b.N; n++ {
		hack.SliceToString(buf.Bytes())
	}
}

func BenchmarkBufferToString(b *testing.B) {
	var buf bytes.Buffer
	buf.WriteString("[")
	buf.Write([]byte("GET"))
	buf.WriteString("]")
	buf.Write([]byte("api"))
	for n := 0; n < b.N; n++ {
		buf.String()
	}
}
