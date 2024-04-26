package relayer

import (
	"encoding/gob"
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

func (self *WrapFE) handshake(c net.Conn) (p core.Port, addr net.Addr, err error) {
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
		return
	}
	addr = &req.Addr
	return
}

func (self *WrapFE) Accept() (p core.Port, addr net.Addr, err error) {
	c, err := self.ln.Accept()
	if err != nil {
		return nil, nil, err
	}
	p, addr, err = self.handshake(c)
	if err != nil {
		c.Close()
	}
	return
}

func NewWrapBE(raddr *net.TCPAddr, pc core.PortCreator) *WrapBE {
	return &WrapBE{raddr, pc}
}

type WrapBE struct {
	raddr *net.TCPAddr
	pc    core.PortCreator
}

func (self *WrapBE) handshake(c net.Conn, addr net.Addr) (p core.Port, err error) {
	var b core.IoVec
	enc := gob.NewEncoder(&b)
	req := wrap.TCPRequest{*addr.(*net.TCPAddr)}
	err = enc.Encode(&req)
	if err != nil {
		return
	}
	p = self.pc.Create(c)
	err = p.Pack(&b)
	return
}

func (self *WrapBE) Dial(addr net.Addr) (p core.Port, err error) {
	c, err := net.DialTCP("tcp", nil, self.raddr)
	if err != nil {
		return nil, err
	}
	p, err = self.handshake(c, addr)
	if err != nil {
		c.Close()
	}
	return
}
