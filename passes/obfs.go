// Copyright (c) 2023 Kai Luo <gluokai@gmail.com>. All rights reserved.

package passes

// #cgo CXXFLAGS: -O3
// #include "bytes.h"
// #include <stdlib.h>
import "C"

import (
	"unsafe"
)

func byteSwapInPlace(b []byte) {
	l := len(b)
	if l == 0 {
		return
	}
	ptr := unsafe.Pointer(&b[0])
	C.ByteSwapInPlace(ptr, C.size_t(l))
}

type FastOBFS struct{}

func (self *FastOBFS) Encode(p []byte) ([]byte, error) {
	byteSwapInPlace(p)
	return p, nil
}

func (self *FastOBFS) Decode(p []byte) ([]byte, error) {
	byteSwapInPlace(p)
	return p, nil
}
