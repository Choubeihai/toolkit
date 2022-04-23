package ringbuffer

import (
	"errors"
	"github.com/Choubeihai/toolkit/byteconv"
	"github.com/Choubeihai/toolkit/math"
	"github.com/Choubeihai/toolkit/pool/bytebuffer"
	"github.com/Choubeihai/toolkit/pool/byteslice"
)

const (
	// DefaultBufferSize ring-buffer初始化时的大小
	DefaultBufferSize = 1024 // 1KB

	// bufferGrowThreshold ring-buffer的最大容量
	bufferGrowThreshold = 4 * 1024 // 4KB
)

var ErrIsEmpty = errors.New("ring-buffer is empty")

// RingBuffer 用户态环形缓冲区
type RingBuffer struct {
	buf     []byte
	size    int
	r       int // 下一次要读取的位置
	w       int // 下一次要写入的位置
	isEmpty bool
}

// EmptyRingBuffer 空ring-buffer
var EmptyRingBuffer = New(0)

// New returns a new RingBuffer whose buffer has the given size.
func New(size int) *RingBuffer {
	if size == 0 {
		return &RingBuffer{isEmpty: true}
	}
	size = math.CeilToPowerOfTwo(size)
	return &RingBuffer{
		buf:     make([]byte, size),
		size:    size,
		isEmpty: true,
	}
}

// Peek 返回环形缓冲区中的n个byte，但是不移动RingBuffer的r指针
func (rb *RingBuffer) Peek(n int) (head []byte, tail []byte) {
	if rb.isEmpty {
		return
	}

	if n <= 0 {
		return
	}

	if rb.w > rb.r {
		m := rb.w - rb.r
		if m > n {
			m = n
		}
		head = rb.buf[rb.r : rb.r+m]
		return
	}

	m := rb.size - rb.r + rb.w
	if m > n {
		m = n
	}

	if rb.r+m <= rb.size {
		head = rb.buf[rb.r : rb.r+m]
	} else {
		c1 := rb.size - rb.r
		head = rb.buf[rb.r:]
		c2 := m - c1
		tail = rb.buf[:c2]
	}

	return
}

// PeekAll 返回环形缓冲区中的所有字节，但是不移动RingBuffer的r指针
func (rb *RingBuffer) PeekAll() (head []byte, tail []byte) {
	if rb.isEmpty {
		return
	}

	if rb.w > rb.r {
		head = rb.buf[rb.r:rb.w]
		return
	}

	head = rb.buf[rb.r:]
	if rb.w != 0 {
		tail = rb.buf[:rb.w]
	}

	return
}

// Discard 通过移动r指针跳过环形缓冲区中的n个字节
func (rb *RingBuffer) Discard(n int) {
	if n <= 0 {
		return
	}

	if n < rb.Length() {
		rb.r = (rb.r + n) % rb.size
	} else {
		rb.Reset()
	}
}

// Read 从环形缓冲区中读取len(p)个字节到p中，同时移动r指针
func (rb *RingBuffer) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	if rb.isEmpty {
		return 0, ErrIsEmpty
	}

	if rb.w > rb.r {
		n = rb.w - rb.r
		if n > len(p) {
			n = len(p)
		}
		copy(p, rb.buf[rb.r:rb.r+n])
		rb.r += n
		if rb.r == rb.w {
			rb.Reset()
		}
		return
	}

	n = rb.size - rb.r + rb.w
	if n > len(p) {
		n = len(p)
	}

	if rb.r+n <= rb.size {
		copy(p, rb.buf[rb.r:rb.r+n])
	} else {
		c1 := rb.size - rb.r
		copy(p, rb.buf[rb.r:])
		c2 := n - c1
		copy(p[c1:], rb.buf[:c2])
	}
	rb.r = (rb.r + n) % rb.size
	if rb.r == rb.w {
		rb.Reset()
	}

	return
}

// ReadByte 读取环形缓冲区中一个字节，同时移动r指针
func (rb *RingBuffer) ReadByte() (b byte, err error) {
	if rb.isEmpty {
		return 0, ErrIsEmpty
	}
	b = rb.buf[rb.r]
	rb.r++
	if rb.r == rb.size {
		rb.r = 0
	}
	if rb.r == rb.w {
		rb.Reset()
	}

	return
}

// Write 向环形缓冲区中写入len(p)个字节，同时移动w指针
func (rb *RingBuffer) Write(p []byte) (n int, err error) {
	n = len(p)
	if n == 0 {
		return
	}

	free := rb.Free()
	if n > free {
		rb.grow(rb.size + n - free)
	}

	if rb.w >= rb.r {
		c1 := rb.size - rb.w
		if c1 >= n {
			copy(rb.buf[rb.w:], p)
			rb.w += n
		} else {
			copy(rb.buf[rb.w:], p[:c1])
			c2 := n - c1
			copy(rb.buf, p[c1:])
			rb.w = c2
		}
	} else {
		copy(rb.buf[rb.w:], p)
		rb.w += n
	}

	if rb.w == rb.size {
		rb.w = 0
	}

	rb.isEmpty = false

	return
}

// WriteByte 向环形缓冲区中写入一个字节，同时移动w指针
func (rb *RingBuffer) WriteByte(c byte) error {
	if rb.Free() < 1 {
		rb.grow(1)
	}
	rb.buf[rb.w] = c
	rb.w++

	if rb.w == rb.size {
		rb.w = 0
	}
	rb.isEmpty = false

	return nil
}

// Length 返回环形缓冲区中可读取的字节长度
func (rb *RingBuffer) Length() int {
	if rb.r == rb.w {
		if rb.isEmpty {
			return 0
		}
		return rb.size
	}

	if rb.w > rb.r {
		return rb.w - rb.r
	}

	return rb.size - rb.r + rb.w
}

// Len 返回环形缓冲区底层buf的长度len()
func (rb *RingBuffer) Len() int {
	return len(rb.buf)
}

// Cap 返回环形缓冲区底层buf的容量cap()
func (rb *RingBuffer) Cap() int {
	return rb.size
}

// Free 返回环形缓冲区剩余可写入的长度
func (rb *RingBuffer) Free() int {
	if rb.r == rb.w {
		if rb.isEmpty {
			return rb.size
		}
		return 0
	}

	if rb.w < rb.r {
		return rb.r - rb.w
	}

	return rb.size - rb.w + rb.r
}

// WriteString 是 Write 的封装，以字符串的方式写入
func (rb *RingBuffer) WriteString(s string) (int, error) {
	return rb.Write(byteconv.StringToBytes(s))
}

// ByteBuffer 以ByteBuffer的形式返回环形缓冲区中所有可读字节，不移动r指针
func (rb *RingBuffer) ByteBuffer() *bytebuffer.ByteBuffer {
	if rb.isEmpty {
		return nil
	} else if rb.w == rb.r {
		bb := bytebuffer.Get()
		_, _ = bb.Write(rb.buf[rb.r:])
		_, _ = bb.Write(rb.buf[:rb.w])
		return bb
	}

	bb := bytebuffer.Get()
	if rb.w > rb.r {
		_, _ = bb.Write(rb.buf[rb.r:rb.w])
		return bb
	}

	_, _ = bb.Write(rb.buf[rb.r:])

	if rb.w != 0 {
		_, _ = bb.Write(rb.buf[:rb.w])
	}

	return bb
}

// WithByteBuffer 返回一个ByteBuffer，内部字节为：环形缓冲区中所有可读数据+给定字节切片b
func (rb *RingBuffer) WithByteBuffer(b []byte) *bytebuffer.ByteBuffer {
	if rb.isEmpty {
		return &bytebuffer.ByteBuffer{B: b}
	} else if rb.w == rb.r {
		bb := bytebuffer.Get()
		_, _ = bb.Write(rb.buf[rb.r:])
		_, _ = bb.Write(rb.buf[:rb.w])
		_, _ = bb.Write(b)
		return bb
	}

	bb := bytebuffer.Get()
	if rb.w > rb.r {
		_, _ = bb.Write(rb.buf[rb.r:rb.w])
		_, _ = bb.Write(b)
		return bb
	}

	_, _ = bb.Write(rb.buf[rb.r:])

	if rb.w != 0 {
		_, _ = bb.Write(rb.buf[:rb.w])
	}
	_, _ = bb.Write(b)

	return bb
}

// IsFull 返回环形缓冲区是否已满
func (rb *RingBuffer) IsFull() bool {
	return rb.r == rb.w && !rb.isEmpty
}

// IsEmpty 返回环形缓冲区是否为空
func (rb *RingBuffer) IsEmpty() bool {
	return rb.isEmpty
}

// Reset 重置环形缓冲区
func (rb *RingBuffer) Reset() {
	rb.isEmpty = true
	rb.r, rb.w = 0, 0
}

// grow 环形缓冲区底层buf的扩容策略
func (rb *RingBuffer) grow(newCap int) {
	if n := rb.size; n == 0 {
		if newCap <= DefaultBufferSize {
			newCap = DefaultBufferSize
		} else {
			newCap = math.CeilToPowerOfTwo(newCap)
		}
	} else {
		doubleCap := n + n
		if newCap <= doubleCap {
			if n < bufferGrowThreshold {
				newCap = doubleCap
			} else {
				for 0 < n && n < newCap {
					n += n / 4
				}
				if n > 0 {
					newCap = n
				}
			}
		}
	}
	newBuf := byteslice.Get(newCap)
	oldLen := rb.Length()
	_, _ = rb.Read(newBuf)
	byteslice.Put(rb.buf)
	rb.buf = newBuf
	rb.r = 0
	rb.w = oldLen
	rb.size = newCap
	if rb.w > 0 {
		rb.isEmpty = false
	}
}
