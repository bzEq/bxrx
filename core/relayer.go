// Copyright (c) 2024 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

type Relayer struct {
	fe Frontend
	be Backend
}

func NewRelayer(fe Frontend, be Backend) *Relayer {
	return &Relayer{fe, be}
}

type AcceptResult struct {
	Port
	Addr string
}

type Frontend interface {
	Accept() chan AcceptResult
}

type DialResult struct {
	Port
}

type Backend interface {
	Dial(addr string) chan DialResult
}

func (self *Relayer) Relay() error {
	for {
		c := self.fe.Accept()
		go func(chan AcceptResult) {
			ar, ok := <-c
			if !ok {
				return
			}
			defer ar.Port.Close()
			dr, ok := <-self.be.Dial(ar.Addr)
			if !ok {
				return
			}
			defer dr.Port.Close()
			RunSimpleSwitch(ar.Port, dr.Port)
		}(c)
	}
	return nil
}
