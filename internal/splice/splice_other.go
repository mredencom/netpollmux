//go:build !linux
// +build !linux

package splice

import (
	"github.com/php2go/netpollmux/internal/buffer"
	"net"
)

const (
	// maxSpliceSize is the maximum amount of data Splice asks
	// the kernel to move in a single call to splice(2).
	maxSpliceSize = 64 << 10
)

// newContext returns a new context.
func newContext(b *bucket) (*context, error) {
	pool := buffer.AssignPool(maxSpliceSize)
	buf := pool.GetBuffer()
	return &context{buffer: buf, pool: pool, bucket: b}, nil
}

// Close closes the context.
func (ctx *context) Close() {
	ctx.pool.PutBuffer(ctx.buffer[:cap(ctx.buffer)])
}

// Splice wraps the splice system call.
//
// splice() moves data between two file descriptors without copying between
// kernel address space and user address space. It transfers up to len bytes
// of data from the file descriptor rfd to the file descriptor wfd,
// where one of the descriptors must refer to a pipe.
func Splice(dst, src net.Conn, len int64) (n int64, err error) {
	return spliceBuffer(dst, src, len)
}
