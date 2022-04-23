package bytebuffer

import "github.com/valyala/bytebufferpool"

// ByteBuffer 是bytebufferpool.ByteBuffer的别名.
type ByteBuffer = bytebufferpool.ByteBuffer

// Get 从pool中取出一个空的byte buffer
var Get = bytebufferpool.Get

// Put 向pool中归还 byte buffer
var Put = func(b *ByteBuffer) {
	if b != nil {
		bytebufferpool.Put(b)
	}
}
