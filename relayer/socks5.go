package relayer

import (
	"log"
	"net"

	"github.com/bzEq/bxrx/core"
	"github.com/bzEq/bxrx/proxy/socks5"
	"github.com/go-errors/errors"
)

type Socks5FE struct {
	ln *net.TCPListener
}

func NewSocks5FE(ln *net.TCPListener) *Socks5FE {
	return &Socks5FE{ln}
}

func (self *Socks5FE) handshake(c net.Conn) (p core.Port, addr string, err error) {
	err = socks5.ExchangeMetadata(c)
	if err != nil {
		log.Println("ExchangeMetadata")
		return
	}
	req, err := socks5.ReceiveRequest(c)
	if err != nil {
		log.Println("ReceiveRequest")
		return nil, "", err
	}
	switch req.CMD {
	case socks5.CMD_CONNECT:
		reply := socks5.Reply{
			VER:      req.VER,
			REP:      socks5.REP_SUCC,
			ATYP:     1,
			BND_ADDR: make([]byte, net.IPv4len),
		}
		socks5.SendReply(c, reply)
		addr = socks5.GetDialAddress(req)
	default:
		reply := socks5.Reply{
			VER:      req.VER,
			REP:      socks5.REP_COMMAND_NOT_SUPPORTED,
			ATYP:     1,
			BND_ADDR: make([]byte, net.IPv4len),
		}
		socks5.SendReply(c, reply)
		err = errors.Errorf("Unsupported CMD: %d", req.CMD)
	}
	p = core.NewRawPort(c)
	return
}

func (self *Socks5FE) Accept() (p core.Port, addr string, err error) {
	c, err := self.ln.Accept()
	if err != nil {
		return nil, "", err
	}
	p, addr, err = self.handshake(c)
	if err != nil {
		c.Close()
	}
	return
}
