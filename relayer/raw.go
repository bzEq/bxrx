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

func (self *TCPBE) Dial(addr net.Addr) (core.Port, error) {
	c, err := net.Dial(addr.Network(), addr.String())
	if err != nil {
		return nil, err
	}
	return core.NewRawPort(c), nil
}
