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
	pc core.PortCreator
}

func NewWrapFE(ln *net.TCPListener, pc core.PortCreator) *WrapFE {
	return &WrapFE{ln, pc}
}

func (self *WrapFE) handshake(c net.Conn) (p core.Port, addr string, err error) {
	p = self.pc.Create(c)
	var b core.IoVec
	err = p.Unpack(&b)
	if err != nil {
		return
	}
	dec := gob.NewDecoder(&b)
	var req wrap.TCPRequest
	err = dec.Decode(&req)
	if err != nil {
		log.Println(err)
		return
	}
	addr = req.Addr
	return
}

func (self *WrapFE) Accept() (p core.Port, addr string, err error) {
	c, err := self.ln.Accept()
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	p, addr, err = self.handshake(c)
	if err != nil {
		c.Close()
	}
	return
}

func NewWrapBE(raddr string, pc core.PortCreator) *WrapBE {
	return &WrapBE{raddr, pc}
}

type WrapBE struct {
	raddr string
	pc    core.PortCreator
}

func (self *WrapBE) handshake(c net.Conn, addr string) (p core.Port, err error) {
	var b core.IoVec
	enc := gob.NewEncoder(&b)
	req := wrap.TCPRequest{addr}
	err = enc.Encode(&req)
	if err != nil {
		log.Println(err)
		return
	}
	p = self.pc.Create(c)
	err = p.Pack(&b)
	return
}

func (self *WrapBE) Dial(addr string) (p core.Port, err error) {
	c, err := net.Dial("tcp", self.raddr)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	p, err = self.handshake(c, addr)
	if err != nil {
		c.Close()
	}
	return
}
