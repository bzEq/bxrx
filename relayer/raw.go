// Copyright (c) 2024 Kai Luo <gluokai@gmail.com>. All rights reserved.

package relayer

import (
	"log"
	"net"

	"github.com/bzEq/bxrx/core"
)

type Options struct {
	LocalAddr      string
	LocalHTTPProxy string
	NextHop        string
}

type TCPBE struct{}

func (self *TCPBE) Dial(addr string) (ch chan core.DialResult) {
	ch = make(chan core.DialResult)
	go func() {
		c, err := net.Dial("tcp", addr)
		if err != nil {
			log.Println(err)
			close(ch)
			return
		}
		log.Println("Relaying to", addr, "at", c.LocalAddr())
		ch <- core.DialResult{core.NewRawNetPort(c)}
	}()
	return
}
