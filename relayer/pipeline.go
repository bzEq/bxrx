// Copyright (c) 2022 Kai Luo <gluokai@gmail.com>. All rights reserved.

package relayer

import (
	"net"

	"github.com/bzEq/bxrx/core"
	"github.com/bzEq/bxrx/pass"
)

func createRandomCodec() (*pass.RandomEncoder, *pass.RandomDecoder) {
	enc := &pass.RandomEncoder{}
	dec := &pass.RandomDecoder{}
	{
		pmb := &core.PackUnpackPassManagerBuilder{}
		pmb.AddPairedPasses(&pass.OBFSEncoder{}, &pass.OBFSDecoder{})
		pmb.AddPairedPasses(&pass.TailPaddingEncoder{}, &pass.TailPaddingDecoder{})
		pmb.AddPairedPasses(&pass.OBFSEncoder{}, &pass.OBFSDecoder{})
		enc.AddPM(pmb.BuildPackPassManager())
		dec.AddPM(pmb.BuildUnpackPassManager())
	}
	{
		pmb := &core.PackUnpackPassManagerBuilder{}
		pmb.AddPairedPasses(&pass.TailPaddingEncoder{}, &pass.TailPaddingDecoder{})
		pmb.AddPairedPasses(&pass.OBFSEncoder{}, &pass.OBFSDecoder{})
		pmb.AddPairedPasses(&pass.TailPaddingEncoder{}, &pass.TailPaddingDecoder{})
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
