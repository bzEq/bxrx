package core

import (
	"net"
	"testing"
)

func TestJsonRPC(t *testing.T) {
	type Req struct {
		Id uint64
	}
	type Resp struct {
		Rc int32
	}
	p0, p1 := net.Pipe()
	go func() {
		rpc := &JsonRPC{P: NewNetPort(p0, &HTTPProtocol{})}
		var req Req
		rpc.ReadRequest(&req)
		rpc.SendResponse(&Resp{Rc: 1024})
	}()
	rpc := &JsonRPC{P: NewNetPort(p1, &HTTPProtocol{})}
	var resp Resp
	rpc.Request(&Req{}, &resp)
	if resp.Rc != 1024 {
		t.Log(resp)
		t.Fail()
	}
}

func TestGobRPC(t *testing.T) {
	type Req struct {
		Id uint64
	}
	type Resp struct {
		Rc int32
	}
	p0, p1 := net.Pipe()
	go func() {
		rpc := &GobRPC{P: NewNetPort(p0, &HTTPProtocol{})}
		var req Req
		rpc.ReadRequest(&req)
		rpc.SendResponse(&Resp{Rc: 1024})
	}()
	rpc := &GobRPC{P: NewNetPort(p1, &HTTPProtocol{})}
	var resp Resp
	rpc.Request(&Req{}, &resp)
	if resp.Rc != 1024 {
		t.Log(resp)
		t.Fail()
	}
}

func TestPacketPipe(t *testing.T) {
	p := MakePipe()
	go func() {
		p[1].Write([]byte{0, 1, 2, 3})
	}()
	b := make([]byte, 8)
	n, err := p[0].Read(b)
	if err != nil || n != 4 {
		t.Fail()
	}
}
