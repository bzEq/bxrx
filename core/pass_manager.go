// Copyright (c) 2023 Kai Luo <gluokai@gmail.com>. All rights reserved.

package core

import (
	"sync"
)

type Pass interface {
	Run(*IoVec) error
}

type PassManager struct {
	passes []Pass
}

func (self *PassManager) AddPass(p Pass) *PassManager {
	self.passes = append(self.passes, p)
	return self
}

func (self *PassManager) Run(b *IoVec) (err error) {
	for _, p := range self.passes {
		err = p.Run(b)
		if err != nil {
			return
		}
	}
	return
}

func NewPassManager(passes []Pass) *PassManager {
	return &PassManager{
		passes,
	}
}

type PackUnpackPassManagerBuilder struct {
	packPasses   []Pass
	unpackPasses []Pass
}

func (self *PackUnpackPassManagerBuilder) AddPairedPasses(pack Pass, unpack Pass) {
	self.packPasses = append(self.packPasses, pack)
	self.unpackPasses = append(self.unpackPasses, unpack)
}

func (self *PackUnpackPassManagerBuilder) BuildPackPassManager() *PassManager {
	return NewPassManager(self.packPasses)
}

func (self *PackUnpackPassManagerBuilder) BuildUnpackPassManager() *PassManager {
	n := len(self.unpackPasses)
	for i := 0; i < n/2; i++ {
		self.unpackPasses[i], self.unpackPasses[n-i-1] = self.unpackPasses[n-i-1], self.unpackPasses[i]
	}
	return NewPassManager(self.unpackPasses)
}

type SyncPass struct {
	Pass
	mu *sync.Mutex
}

func (self *SyncPass) Run(b *IoVec) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	return self.Pass.Run(b)
}

func AsSyncPass(p Pass, mu *sync.Mutex) *SyncPass {
	sp, succ := p.(*SyncPass)
	if !succ {
		return &SyncPass{p, mu}
	}
	return sp
}
