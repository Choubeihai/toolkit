package byteslice

import (
	"math"
	"math/bits"
	"sync"
)

var builtinPool Pool

// Pool 有32个sync.Pool，代表byte slice的长度(2的幂次方)
type Pool struct {
	pools [32]sync.Pool
}

// Get 从builtinPool中返回一个长度为size的byte slice
func Get(size int) []byte {
	return builtinPool.Get(size)
}

// Put 向builtinPool中归还byte slice
func Put(buf []byte) {
	builtinPool.Put(buf)
}

func (p *Pool) Get(size int) []byte {
	if size <= 0 {
		return nil
	}
	if size > math.MaxInt32 {
		return make([]byte, size)
	}
	idx := index(uint32(size))
	if v := p.pools[idx].Get(); v != nil {
		bp := v.(*[]byte)
		return (*bp)[:size]
	}
	return make([]byte, 1<<idx)[:size]
}

func (p *Pool) Put(buf []byte) {
	size := cap(buf)
	if size == 0 || size > math.MaxInt32 {
		return
	}
	idx := index(uint32(size))
	if size != 1<<idx { // this byte slice is not from Pool.Get(), put it into the previous interval of idx
		idx--
	}
	p.pools[idx].Put(&buf)
}

// index 返回长度为n字节的byte slice的Pool.pools下标
func index(n uint32) uint32 {
	return uint32(bits.Len32(n - 1))
}
