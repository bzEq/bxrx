// Copyright (c) 2024 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

import (
	"log"
	"net"
)

type Relayer struct {
	fe Frontend
	be Backend
}

func NewRelayer(fe Frontend, be Backend) *Relayer {
	return &Relayer{fe, be}
}

type Frontend interface {
	Accept() (Port, net.Addr, error)
}

type Backend interface {
	Dial(addr net.Addr) (Port, error)
}

func (self *Relayer) Relay() error {
	for {
		fp, addr, err := self.fe.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go func(fp Port, addr net.Addr) {
			defer fp.Close()
			log.Println("Dialing ", addr)
			bp, err := self.be.Dial(addr)
			if err != nil {
				log.Println(err)
				return
			}
			defer bp.Close()
			RunSimpleSwitch(fp, bp)
		}(fp, addr)
	}
	return nil
}
