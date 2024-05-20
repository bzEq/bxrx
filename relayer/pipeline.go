// Copyright (c) 2022 Kai Luo <gluokai@gmail.com>. All rights reserved.

package relayer

import (
	"errors"
	"io"
	"net"
	"net/http"
	"sync"

	"github.com/bzEq/bxrx/core"
	"github.com/bzEq/bxrx/pass"
)

func createRandomCodec() (*pass.RandomEncoder, *pass.RandomDecoder) {
	enc := &pass.RandomEncoder{}
	dec := &pass.RandomDecoder{}
	{
		pmb := &core.PackUnpackPassManagerBuilder{}
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
	{
		pmb := &core.PackUnpackPassManagerBuilder{}
		pmb.AddPairedPasses(&pass.OBFSEncoder{}, &pass.OBFSDecoder{})
		enc.AddPM(pmb.BuildPackPassManager())
		dec.AddPM(pmb.BuildUnpackPassManager())
	}
	{
		pmb := &core.PackUnpackPassManagerBuilder{}
		pmb.AddPairedPasses(&pass.OBFSEncoder{}, &pass.OBFSDecoder{})
		pmb.AddPairedPasses(&pass.TailPaddingEncoder{}, &pass.TailPaddingDecoder{})
		enc.AddPM(pmb.BuildPackPassManager())
		dec.AddPM(pmb.BuildUnpackPassManager())
	}
	return enc, dec
}

type Pipeline struct{}

func (self *Pipeline) FromConn(c net.Conn) core.Port {
	enc, dec := createRandomCodec()
	pack := &core.PassManager{}
	unpack := &core.PassManager{}
	mu := &sync.Mutex{}
	pack.AddPass(enc).AddPass(core.NewSyncPass(pass.NewHTTPEncoder(c), mu))
	unpack.AddPass(pass.NewHTTPDecoder(c)).AddPass(dec)
	return core.NewNetPort(c, pack, &HTTPInternalErrorPass{unpack, c, mu})
}

type HTTPInternalErrorPass struct {
	core.Pass
	io.Writer
	mu *sync.Mutex
}

func (self *HTTPInternalErrorPass) InternalError() error {
	resp := http.Response{
		Status:     "500 Internal Server Error",
		StatusCode: 500,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}
	self.mu.Lock()
	defer self.mu.Unlock()
	return core.Tr(resp.Write(self.Writer))
}

func (self *HTTPInternalErrorPass) Run(b *core.IoVec) error {
	if err := self.Pass.Run(b); err != nil {
		if !errors.Is(err, io.EOF) {
			self.InternalError()
		}
		return err
	}
	return nil
}
