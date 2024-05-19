package relayer

import (
	"bytes"
	"strings"
	"sync"
	"testing"
)

func TestHTTPInternalError(t *testing.T) {
	buf := &bytes.Buffer{}
	dec := &HTTPInternalErrorPass{
		Pass:   nil,
		Writer: buf,
		mu:     &sync.Mutex{},
	}
	if err := dec.InternalError(); err != nil {
		t.Fatal(err)
	}
	s := string(buf.Bytes())
	if strings.HasPrefix("HTTP/1.1 500", s) {
		t.Log(s)
		t.Fail()
	}
}
