// Copyright (c) 2020 Kai Luo <gluokai@gmail.com>. All rights reserved.

package pass

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"net/http"

	"github.com/bzEq/bxrx/core"
)

type TailPaddingEncoder struct{}

func (self *TailPaddingEncoder) Run(b *core.IoVec) error {
	l := (rand.Uint32() % 64) & (uint32(63) << 2)
	var padding bytes.Buffer
	for i := uint32(0); i < l/4; i++ {
		binary.Write(&padding, binary.BigEndian, rand.Uint32())
	}
	padding.WriteByte(byte(l))
	b.Take(padding.Bytes())
	return nil
}

type TailPaddingDecoder struct{}

func (self *TailPaddingDecoder) Run(b *core.IoVec) error {
	t, err := b.LastByte()
	if err != nil {
		return core.Tr(err)
	}
	return b.Drop(1 + int(t))
}

type OBFSEncoder struct {
	FastOBFS
}

func (self *OBFSEncoder) Run(b *core.IoVec) error {
	buf := b.Consume()
	buf, err := self.FastOBFS.Encode(buf)
	if err != nil {
		return core.Tr(err)
	}
	b.Take(buf)
	return nil
}

type OBFSDecoder struct {
	FastOBFS
}

func (self *OBFSDecoder) Run(b *core.IoVec) error {
	buf := b.Consume()
	buf, err := self.FastOBFS.Decode(buf)
	if err != nil {
		return core.Tr(err)
	}
	b.Take(buf)
	return nil
}

type RandomEncoder struct {
	pms []*core.PassManager
}

func (self *RandomEncoder) AddPM(p *core.PassManager) {
	self.pms = append(self.pms, p)
}

func (self *RandomEncoder) Run(b *core.IoVec) error {
	n := int(rand.Uint32())
	if err := self.pms[n%len(self.pms)].Run(b); err != nil {
		return core.Tr(err)
	}
	var padding bytes.Buffer
	padding.WriteByte(byte(n))
	b.Take(padding.Bytes())
	return nil
}

type RandomDecoder struct {
	pms []*core.PassManager
}

func (self *RandomDecoder) AddPM(p *core.PassManager) {
	self.pms = append(self.pms, p)
}

func (self *RandomDecoder) Run(b *core.IoVec) error {
	t, err := b.LastByte()
	if err != nil {
		return core.Tr(err)
	}
	n := int(t)
	b.Drop(1)
	return self.pms[n%len(self.pms)].Run(b)
}

type HTTPEncoder struct {
	wbuf *bufio.Writer
}

func NewHTTPEncoder(w io.Writer) *HTTPEncoder {
	return &HTTPEncoder{
		bufio.NewWriter(w),
	}
}

func (self *HTTPEncoder) Run(b *core.IoVec) error {
	req, err := http.NewRequest("POST", "/", b)
	if err != nil {
		return core.Tr(err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.ContentLength = int64(b.Len())
	if req.ContentLength <= 0 {
		err = fmt.Errorf("Content length %d is abnormal", req.ContentLength)
		return core.Tr(err)
	}
	err = req.Write(self.wbuf)
	if err != nil {
		return core.Tr(err)
	}
	return core.Tr(self.wbuf.Flush())
}

type HTTPDecoder struct {
	rbuf *bufio.Reader
}

func NewHTTPDecoder(r io.Reader) *HTTPDecoder {
	return &HTTPDecoder{
		bufio.NewReader(r),
	}
}

func (self *HTTPDecoder) Run(b *core.IoVec) error {
	req, err := http.ReadRequest(self.rbuf)
	if err != nil {
		return core.Tr(err)
	}
	defer req.Body.Close()
	if req.ContentLength <= 0 || req.ContentLength > core.DEFAULT_BUFFER_LIMIT {
		err = fmt.Errorf("Content length %d is abnormal", req.ContentLength)
		return core.Tr(err)
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return core.Tr(err)
	}
	if int64(len(body)) != req.ContentLength {
		err = fmt.Errorf("Content length %d, %d bytes read in the body", req.ContentLength, len(body))
		return core.Tr(err)
	}
	b.Take(body)
	return nil
}
