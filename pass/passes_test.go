package pass

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"testing"

	"github.com/bzEq/bxrx/core"
)

func generateRandomSlice(l int) []byte {
	m := l / 8
	r := l % 8
	var buf bytes.Buffer
	for i := 0; i < m; i++ {
		binary.Write(&buf, binary.BigEndian, rand.Uint64())
	}
	for i := 0; i < r; i++ {
		buf.WriteByte(byte(rand.Uint64()))
	}
	return buf.Bytes()
}

func testCodec(t *testing.T, enc core.Pass, dec core.Pass) {
	for i := 24; i < 24+8; i++ {
		v := &core.IoVec{}
		s := generateRandomSlice(i)
		ss := string(s)
		v.Take(s)
		if err := enc.Run(v); err != nil {
			t.Fatal(err)
		}
		if err := dec.Run(v); err != nil {
			t.Fatal(err)
		}
		sss := string(v.Consume())
		if ss != sss {
			t.Log(ss)
			t.Log(sss)
			t.Fatal("Slice not equal after enc and dec")
		}
	}
}

func TestTailPadding(t *testing.T) {
	enc := &TailPaddingEncoder{}
	dec := &TailPaddingDecoder{}
	testCodec(t, enc, dec)
}

func TestOBFS(t *testing.T) {
	enc := &OBFSEncoder{}
	dec := &OBFSDecoder{}
	testCodec(t, enc, dec)
}

func TestHTTP(t *testing.T) {
	buf := &bytes.Buffer{}
	enc := NewHTTPEncoder(buf)
	dec := NewHTTPDecoder(buf, buf)
	testCodec(t, enc, dec)
}

func TestRandomCodec(t *testing.T) {
	enc := &RandomEncoder{}
	encPM0 := &core.PassManager{}
	encPM0.AddPass(&TailPaddingEncoder{})
	encPM0.AddPass(&OBFSEncoder{})
	enc.AddPM(encPM0)
	encPM1 := &core.PassManager{}
	encPM1.AddPass(&OBFSEncoder{})
	encPM1.AddPass(&TailPaddingEncoder{})
	enc.AddPM(encPM1)

	dec := &RandomDecoder{}
	decPM0 := &core.PassManager{}
	decPM0.AddPass(&OBFSDecoder{})
	decPM0.AddPass(&TailPaddingDecoder{})
	dec.AddPM(decPM0)
	decPM1 := &core.PassManager{}
	decPM1.AddPass(&TailPaddingDecoder{})
	decPM1.AddPass(&OBFSDecoder{})
	dec.AddPM(decPM1)
	testCodec(t, enc, dec)
}
