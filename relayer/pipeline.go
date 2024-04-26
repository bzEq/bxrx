// Copyright (c) 2022 Kai Luo <gluokai@gmail.com>. All rights reserved.

package relayer

import (
	"github.com/bzEq/bxrx/core"
	"github.com/bzEq/bxrx/passes"
	"net"
)

func createRandomCodec() (*passes.RandomEncoder, *passes.RandomDecoder) {
	enc := &passes.RandomEncoder{}
	dec := &passes.RandomDecoder{}
	{
		pmb := &core.PackUnpackPassManagerBuilder{}
		pmb.AddPairedPasses(&passes.OBFSEncoder{}, &passes.OBFSDecoder{})
		pmb.AddPairedPasses(&passes.TailPaddingEncoder{}, &passes.TailPaddingDecoder{})
		pmb.AddPairedPasses(&passes.OBFSEncoder{}, &passes.OBFSDecoder{})
		enc.AddPM(pmb.BuildPackPassManager())
		dec.AddPM(pmb.BuildUnpackPassManager())
	}
	{
		pmb := &core.PackUnpackPassManagerBuilder{}
		pmb.AddPairedPasses(&passes.TailPaddingEncoder{}, &passes.TailPaddingDecoder{})
		pmb.AddPairedPasses(&passes.OBFSEncoder{}, &passes.OBFSDecoder{})
		pmb.AddPairedPasses(&passes.TailPaddingEncoder{}, &passes.TailPaddingDecoder{})
		enc.AddPM(pmb.BuildPackPassManager())
		dec.AddPM(pmb.BuildUnpackPassManager())
	}
	return enc, dec
}

func CreateProtocol(name string) core.Protocol {
	switch name {
	case "raw":
		return nil
	default:
		enc, dec := createRandomCodec()
		return &core.ProtocolWithPass{
			P:  &core.HTTPProtocol{},
			UP: dec,
			PP: enc,
		}
	}
}

type Pipeline struct{}

func (self *Pipeline) Create(c net.Conn) core.Port {
	return core.NewPort(c, CreateProtocol(""))
}
