// Copyright (c) 2021 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

import (
	"bufio"
	"fmt"
	"io"
	"log"
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
	err := self.PP.Run(b)
	if err != nil {
		return err
	}
	return self.P.Pack(b, out)
}

func (self *ProtocolWithPass) Unpack(in *bufio.Reader, b *IoVec) error {
	err := self.P.Unpack(in, b)
	if err != nil {
		return err
	}
	return self.UP.Run(b)
}

type HTTPProtocol struct{}

func (self *HTTPProtocol) Pack(b *IoVec, out *bufio.Writer) error {
	req, err := http.NewRequest("POST", "/", b)
	if err != nil {
		log.Println(err)
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	if req.GetBody == nil {
		req.ContentLength = int64(b.Len())
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(b), nil
		}
	}
	log.Println("Packing %d bytes", req.ContentLength)
	return req.Write(out)
}

func (self *HTTPProtocol) Unpack(in *bufio.Reader, b *IoVec) error {
	req, err := http.ReadRequest(in)
	if err != nil {
		log.Println(err)
		return err
	}
	defer req.Body.Close()
	if req.ContentLength < 0 || req.ContentLength > DEFAULT_BUFFER_LIMIT {
		err = fmt.Errorf("Content length %d is abnormal", req.ContentLength)
		log.Println(err)
		return err
	}
	log.Println("Unpacking %d bytes", req.ContentLength)
	body := make([]byte, req.ContentLength)
	if _, err = io.ReadFull(req.Body, body); err != nil {
		log.Println(err)
		return err
	}
	b.Take(body)
	return nil
}
