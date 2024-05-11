// Copyright (c) 2021 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
)

type Protocol interface {
	Pack(*IoVec, *bufio.Writer) error
	Unpack(*bufio.Reader, *IoVec) error
}

type ProtocolWithPass struct {
	P  Protocol
	PP Pass
	UP Pass
}

func (self *ProtocolWithPass) Pack(b *IoVec, out *bufio.Writer) error {
	if err := self.PP.Run(b); err != nil {
		return Tr(err)
	}
	return Tr(self.P.Pack(b, out))
}

func (self *ProtocolWithPass) Unpack(in *bufio.Reader, b *IoVec) error {
	if err := self.P.Unpack(in, b); err != nil {
		return Tr(err)
	}
	return Tr(self.UP.Run(b))
}

type HTTPProtocol struct{}

func (self *HTTPProtocol) Pack(b *IoVec, out *bufio.Writer) error {
	req, err := http.NewRequest("POST", "/", b)
	if err != nil {
		return Tr(err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.ContentLength = int64(b.Len())
	if req.ContentLength <= 0 {
		err = fmt.Errorf("Content length %d is abnormal", req.ContentLength)
		return Tr(err)
	}
	return Tr(req.Write(out))
}

func (self *HTTPProtocol) Unpack(in *bufio.Reader, b *IoVec) error {
	req, err := http.ReadRequest(in)
	if err != nil {
		return Tr(err)
	}
	defer req.Body.Close()
	if req.ContentLength <= 0 || req.ContentLength > DEFAULT_BUFFER_LIMIT {
		err = fmt.Errorf("Content length %d is abnormal", req.ContentLength)
		return Tr(err)
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return Tr(err)
	}
	if int64(len(body)) != req.ContentLength {
		err = fmt.Errorf("Content length %d, %d bytes read in the body", req.ContentLength, len(body))
		return Tr(err)
	}
	b.Take(body)
	return nil
}
