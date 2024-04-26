// Copyright (c) 2024 Kai Luo <gluokai@gmail.com>. All rights reserved.

package relayer

import (
	"net"

	"github.com/bzEq/bxrx/core"
)

type Options struct {
	LocalAddr string
	NextHop   string
}

type TCPBE struct{}

func (self *TCPBE) Dial(addr string) (core.Port, error) {
	c, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return core.NewRawPort(c), nil
}
