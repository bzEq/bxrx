// Copyright (c) 2020 Kai Luo <gluokai@gmail.com>. All rights reserved.

package pass

import (
	"bytes"
	"encoding/binary"
	"math/rand"

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
		return err
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
