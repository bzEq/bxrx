// Copyright (c) 2020 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

import (
	"errors"
	"io"
	"log"
)

// SimpleSwitch is not responsible to close ports.
type SimpleSwitch struct {
	done [2]chan struct{}
	port [2]Port
}

func (self *SimpleSwitch) Run() {
	go func() {
		defer close(self.done[0])
		if err := self.switchTraffic(self.port[0], self.port[1]); err != nil {
			log.Println(err)
		}
	}()
	go func() {
		defer close(self.done[1])
		if err := self.switchTraffic(self.port[1], self.port[0]); err != nil {
			log.Println(err)
		}
	}()
	<-self.done[0]
	<-self.done[1]
}

func (self *SimpleSwitch) switchTraffic(in, out Port) error {
	for {
		var b IoVec
		if err := in.Unpack(&b); err != nil {
			// Don't treat io.EOF as error.
			if errors.Is(err, io.EOF) {
				err = nil
			}
			out.CloseWrite()
			return Tr(err)
		}
		if err := out.Pack(&b); err != nil {
			in.CloseRead()
			return Tr(err)
		}
	}
}

func RunSimpleSwitch(p0, p1 Port) {
	NewSimpleSwitch(p0, p1).Run()
}

func NewSimpleSwitch(p0, p1 Port) *SimpleSwitch {
	s := &SimpleSwitch{
		port: [2]Port{p0, p1},
		done: [2]chan struct{}{make(chan struct{}), make(chan struct{})},
	}
	return s
}
