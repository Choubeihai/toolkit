package listbuffer

import "github.com/Choubeihai/toolkit/pool/bytebuffer"

type ByteBuffer struct {
	Buf  *bytebuffer.ByteBuffer
	next *ByteBuffer
}

// Len 返回 ByteBuffer 的长度
func (b *ByteBuffer) Len() int {
	if b.Buf == nil {
		return -1
	}
	return b.Buf.Len()
}

// IsEmpty ByteBuffer是否为空
func (b *ByteBuffer) IsEmpty() bool {
	if b.Buf == nil {
		return true
	}
	return b.Buf.Len() == 0
}

// ListBuffer is a linked list of ByteBuffer.
type ListBuffer struct {
	bs    [][]byte
	head  *ByteBuffer
	tail  *ByteBuffer
	size  int
	bytes int64 // the total size of the link list of ByteBuffer
}

// Pop returns and removes the head of l. If l is empty, it returns nil.
func (l *ListBuffer) Pop() *ByteBuffer {
	if l.head == nil {
		return nil
	}
	b := l.head
	l.head = b.next
	if l.head == nil {
		l.tail = nil
	}
	b.next = nil
	l.size--
	l.bytes -= int64(b.Buf.Len())
	return b
}

// PushFront adds the new node to the head of l.
func (l *ListBuffer) PushFront(b *ByteBuffer) {
	if b == nil {
		return
	}
	if l.head == nil {
		b.next = nil
		l.tail = b
	} else {
		b.next = l.head
	}
	l.head = b
	l.size++
	l.bytes += int64(b.Buf.Len())
}

// PushBack adds a new node to the tail of l.
func (l *ListBuffer) PushBack(b *ByteBuffer) {
	if b == nil {
		return
	}
	if l.tail == nil {
		l.head = b
	} else {
		l.tail.next = b
	}
	b.next = nil
	l.tail = b
	l.size++
	l.bytes += int64(b.Buf.Len())
}

// PushBytesFront is a wrapper of PushFront, which accepts []byte as its argument.
func (l *ListBuffer) PushBytesFront(p []byte) {
	if len(p) == 0 {
		return
	}
	bb := bytebuffer.Get()
	_, _ = bb.Write(p)
	l.PushFront(&ByteBuffer{Buf: bb})
}

// PushBytesBack is a wrapper of PushBack, which accepts []byte as its argument.
func (l *ListBuffer) PushBytesBack(p []byte) {
	if len(p) == 0 {
		return
	}
	bb := bytebuffer.Get()
	_, _ = bb.Write(p)
	l.PushBack(&ByteBuffer{Buf: bb})
}

// PeekBytesList assembles the [][]byte based on the list of ByteBuffer,
// it won't remove these nodes from l until DiscardBytes() is called.
func (l *ListBuffer) PeekBytesList() [][]byte {
	l.bs = l.bs[:0]
	for iter := l.head; iter != nil; iter = iter.next {
		l.bs = append(l.bs, iter.Buf.B)
	}
	return l.bs
}

// PeekBytesListWithBytes is like PeekBytesList but accepts [][]byte and puts them onto head.
func (l *ListBuffer) PeekBytesListWithBytes(bs ...[]byte) [][]byte {
	l.bs = l.bs[:0]
	for _, b := range bs {
		if len(b) > 0 {
			l.bs = append(l.bs, b)
		}
	}
	for iter := l.head; iter != nil; iter = iter.next {
		l.bs = append(l.bs, iter.Buf.B)
	}
	return l.bs
}

// DiscardBytes removes some nodes based on n.
func (l *ListBuffer) DiscardBytes(n int) {
	if n <= 0 {
		return
	}
	for n != 0 {
		b := l.Pop()
		if b == nil {
			break
		}
		if n < b.Len() {
			b.Buf.B = b.Buf.B[n:]
			l.PushFront(b)
			break
		}
		n -= b.Len()
		bytebuffer.Put(b.Buf)
	}
}

// Len returns the length of the list.
func (l *ListBuffer) Len() int {
	return l.size
}

// Bytes returns the amount of bytes in this list.
func (l *ListBuffer) Bytes() int64 {
	return l.bytes
}

// IsEmpty reports whether l is empty.
func (l *ListBuffer) IsEmpty() bool {
	return l.head == nil
}

// Reset removes all elements from this list.
func (l *ListBuffer) Reset() {
	for b := l.Pop(); b != nil; b = l.Pop() {
		bytebuffer.Put(b.Buf)
	}
	l.head = nil
	l.tail = nil
	l.size = 0
	l.bytes = 0
	l.bs = l.bs[:0]
}
