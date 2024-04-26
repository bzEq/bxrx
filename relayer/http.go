package relayer

import (
	"net"

	"github.com/bzEq/bxrx/core"
)

type HTTPProxyFE struct {
	q chan struct {
		core.Port
		string
	}
}

func NewHTTPProxyFE() *HTTPProxyFE {
	return &HTTPProxyFE{
		make(chan struct {
			core.Port
			string
		}),
	}
}

func (self *HTTPProxyFE) Capture(c net.Conn, raddr string) {
	self.q <- struct {
		core.Port
		string
	}{core.NewRawPort(c), raddr}
}

func (self *HTTPProxyFE) Accept() (p core.Port, addr string, err error) {
	t := <-self.q
	p = t.Port
	addr = t.string
	return
}
