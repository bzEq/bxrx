// Copyright (c) 2022 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

const DEFAULT_TIMEOUT = 60 * 60
const DEFAULT_BUFFER_SIZE = 256 << 10
const DEFAULT_BUFFER_LIMIT = 64 << 20
const DEFAULT_UDP_TIMEOUT = 60
const DEFAULT_UDP_BUFFER_SIZE = 2 << 10

type PortCreator interface {
	Create(net.Conn) Port
}

func CloseRead(c net.Conn) error {
	if c, ok := c.(*net.TCPConn); ok {
		return c.CloseRead()
	}
	if c, ok := c.(*net.UnixConn); ok {
		return c.CloseRead()
	}
	return nil
}

func CloseWrite(c net.Conn) error {
	if c, ok := c.(*net.TCPConn); ok {
		return c.CloseWrite()
	}
	if c, ok := c.(*net.UnixConn); ok {
		return c.CloseWrite()
	}
	return nil
}

type Port interface {
	Pack(*IoVec) error
	Unpack(*IoVec) error
	CloseRead() error
	CloseWrite() error
	Close() error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
}

type NetPort struct {
	conn    net.Conn
	proto   Protocol
	rbuf    *bufio.Reader
	wbuf    *bufio.Writer
	timeout time.Duration
}

func NewNetPortWithTimeout(c net.Conn, p Protocol, timeout int) *NetPort {
	return &NetPort{
		conn:    c,
		proto:   p,
		rbuf:    bufio.NewReader(c),
		wbuf:    bufio.NewWriter(c),
		timeout: time.Duration(timeout) * time.Second,
	}
}

func NewNetPort(c net.Conn, p Protocol) *NetPort {
	return NewNetPortWithTimeout(c, p, DEFAULT_TIMEOUT)
}

func (self *NetPort) Unpack(b *IoVec) error {
	if err := self.conn.SetReadDeadline(time.Now().Add(self.timeout)); err != nil {
		return Tr(err)
	}
	if err := self.proto.Unpack(self.rbuf, b); err != nil {
		if errors.Is(err, io.EOF) {
			log.Println(self.conn.RemoteAddr(), "->", self.conn.LocalAddr(), "is closed")
		}
		return Tr(err)
	}
	return nil
}

func (self *NetPort) Pack(b *IoVec) error {
	if err := self.conn.SetWriteDeadline(time.Now().Add(self.timeout)); err != nil {
		return Tr(err)
	}
	if err := self.proto.Pack(b, self.wbuf); err != nil {
		return Tr(err)
	}
	return Tr(self.wbuf.Flush())
}

func (self *NetPort) CloseRead() error {
	return CloseRead(self.conn)
}

func (self *NetPort) CloseWrite() error {
	return CloseWrite(self.conn)
}

func (self *NetPort) Close() error {
	return self.conn.Close()
}

func (self *NetPort) LocalAddr() net.Addr {
	return self.conn.LocalAddr()
}

func (self *NetPort) RemoteAddr() net.Addr {
	return self.conn.RemoteAddr()
}

type RawNetPort struct {
	conn    net.Conn
	timeout time.Duration
	buf     []byte
	nr      int
}

func NewRawNetPortWithTimeout(c net.Conn, timeout int) *RawNetPort {
	return &RawNetPort{
		conn:    c,
		timeout: time.Duration(timeout) * time.Second,
	}
}

func NewRawNetPort(c net.Conn) *RawNetPort {
	return NewRawNetPortWithTimeout(c, DEFAULT_TIMEOUT)
}

func (self *RawNetPort) Pack(b *IoVec) error {
	if err := self.conn.SetWriteDeadline(time.Now().Add(self.timeout)); err != nil {
		return Tr(err)
	}
	_, err := b.WriteTo(self.conn)
	return Tr(err)
}

func (self *RawNetPort) growBuffer() {
	l := len(self.buf)
	nl := l
	if nl <= self.nr {
		if self.nr == 0 {
			// If DEFAULT_BUFFER_SIZE is too small, times of buffer allocation will increase and thus hurt performance.
			nl = DEFAULT_BUFFER_SIZE
		} else {
			nl = self.nr * 2
		}
	}
	// Ensure we have sufficient buffer for UDP transfer.
	if nl < DEFAULT_UDP_BUFFER_SIZE {
		nl = DEFAULT_UDP_BUFFER_SIZE
	}
	if nl > DEFAULT_BUFFER_LIMIT {
		nl = DEFAULT_BUFFER_LIMIT
	}
	if nl <= l {
		return
	}
	self.buf = make([]byte, nl)
}

func (self *RawNetPort) Unpack(b *IoVec) (err error) {
	self.growBuffer()
	err = self.conn.SetReadDeadline(time.Now().Add(self.timeout))
	if err != nil {
		return Tr(err)
	}
	self.nr, err = self.conn.Read(self.buf)
	if err != nil {
		if errors.Is(err, io.EOF) {
			log.Println(self.conn.RemoteAddr(), "->", self.conn.LocalAddr(), "is closed")
		}
		self.nr = 0
		return Tr(err)
	}
	b.Take(self.buf[:self.nr])
	self.buf = self.buf[self.nr:]
	return nil
}

func (self *RawNetPort) CloseRead() error {
	return CloseRead(self.conn)
}

func (self *RawNetPort) CloseWrite() error {
	return CloseWrite(self.conn)
}

func (self *RawNetPort) Close() error {
	return self.conn.Close()
}

func (self *RawNetPort) LocalAddr() net.Addr {
	return self.conn.LocalAddr()
}

func (self *RawNetPort) RemoteAddr() net.Addr {
	return self.conn.RemoteAddr()
}

type SyncPort struct {
	Port
	umu, pmu sync.Mutex
}

func (self *SyncPort) Unpack(b *IoVec) error {
	self.umu.Lock()
	defer self.umu.Unlock()
	return self.Port.Unpack(b)
}

func (self *SyncPort) Pack(b *IoVec) error {
	self.pmu.Lock()
	defer self.pmu.Unlock()
	return self.Port.Pack(b)
}

func NewSyncPortWithTimeout(c net.Conn, p Protocol, timeout int) *SyncPort {
	return &SyncPort{
		Port: NewNetPortWithTimeout(c, p, timeout),
	}
}

func AsSyncPort(p Port) Port {
	if _, succ := p.(*SyncPort); !succ {
		return &SyncPort{Port: p}
	}
	return p
}
