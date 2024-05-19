package relayer

import (
	"encoding/gob"
	"log"
	"net"

	"github.com/bzEq/bxrx/core"
	"github.com/bzEq/bxrx/proxy/wrap"
)

type WrapFE struct {
	ln *net.TCPListener
	pb core.PortBuilder
}

func NewWrapFE(ln *net.TCPListener, pb core.PortBuilder) *WrapFE {
	return &WrapFE{ln, pb}
}

func (self *WrapFE) handshake(c net.Conn) (p core.Port, addr string, err error) {
	p = self.pb.FromConn(c)
	var b core.IoVec
	err = p.Unpack(&b)
	if err != nil {
		err = core.Tr(err)
		return
	}
	dec := gob.NewDecoder(&b)
	var req wrap.TCPRequest
	err = dec.Decode(&req)
	if err != nil {
		err = core.Tr(err)
		return
	}
	addr = req.Addr
	return
}

func (self *WrapFE) Accept() (ch chan core.AcceptResult) {
	ch = make(chan core.AcceptResult)
	c, err := self.ln.Accept()
	if err != nil {
		log.Println(err)
		close(ch)
		return
	}
	go func() {
		p, addr, err := self.handshake(c)
		if err != nil {
			log.Println(err)
			close(ch)
			c.Close()
			return
		}
		ch <- core.AcceptResult{p, addr}
	}()
	return
}

func NewWrapBE(raddr string, pb core.PortBuilder) *WrapBE {
	return &WrapBE{raddr, pb}
}

type WrapBE struct {
	raddr string
	pb    core.PortBuilder
}

func (self *WrapBE) handshake(c net.Conn, addr string) (p core.Port, err error) {
	var b core.IoVec
	enc := gob.NewEncoder(&b)
	req := wrap.TCPRequest{addr}
	err = enc.Encode(&req)
	if err != nil {
		err = core.Tr(err)
		return
	}
	p = self.pb.FromConn(c)
	err = p.Pack(&b)
	if err != nil {
		err = core.Tr(err)
		return
	}
	return
}

func (self *WrapBE) Dial(addr string) (ch chan core.DialResult) {
	ch = make(chan core.DialResult)
	go func() {
		c, err := net.Dial("tcp", self.raddr)
		if err != nil {
			log.Println(err)
			close(ch)
			return
		}
		log.Println("Relaying to", addr, "at", c.LocalAddr())
		p, err := self.handshake(c, addr)
		if err != nil {
			log.Println(err)
			close(ch)
			c.Close()
			return
		}
		ch <- core.DialResult{p}
	}()
	return
}
