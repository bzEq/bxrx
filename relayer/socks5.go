package relayer

import (
	"fmt"
	"log"
	"net"

	"github.com/bzEq/bxrx/core"
	"github.com/bzEq/bxrx/proxy/socks5"
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
		err = core.Tr(err)
		return
	}
	req := &socks5.Request{}
	err = socks5.ReceiveRequest(c, req)
	if err != nil {
		err = core.Tr(err)
		return
	}
	switch req.CMD {
	case socks5.CMD_CONNECT:
		reply := socks5.Reply{
			VER:      req.VER,
			REP:      socks5.REP_SUCC,
			ATYP:     socks5.ATYP_IPV4,
			BND_ADDR: make([]byte, net.IPv4len),
		}
		socks5.SendReply(c, reply)
		addr = socks5.GetDialAddress(req.ATYP, req.DST_ADDR, req.DST_PORT)
		p = core.NewRawNetPort(c)
		return
	default:
		reply := socks5.Reply{
			VER:      req.VER,
			REP:      socks5.REP_COMMAND_NOT_SUPPORTED,
			ATYP:     socks5.ATYP_IPV4,
			BND_ADDR: make([]byte, net.IPv4len),
		}
		socks5.SendReply(c, reply)
		err = core.Tr(fmt.Errorf("Unsupported CMD: %d", req.CMD))
		return
	}
}

func (self *Socks5FE) Accept() (ch chan core.AcceptResult) {
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
