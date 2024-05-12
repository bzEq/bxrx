package relayer

import (
	"net"

	"github.com/bzEq/bxrx/core"
)

type HTTPProxyFE struct {
	ch chan core.AcceptResult
}

func NewHTTPProxyFE() *HTTPProxyFE {
	return &HTTPProxyFE{
		make(chan core.AcceptResult),
	}
}

func (self *HTTPProxyFE) Capture(c net.Conn, raddr string) {
	self.ch <- core.AcceptResult{core.NewRawNetPort(c), raddr}
}

func (self *HTTPProxyFE) Accept() (ch chan core.AcceptResult) {
	c := <-self.ch
	ch = make(chan core.AcceptResult)
	go func() {
		ch <- c
	}()
	return
}
