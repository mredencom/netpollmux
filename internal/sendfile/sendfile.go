package sendfile

import (
	"net"
	mmap2 "netpollmux/internal/mmap"
	"syscall"
)

const (
	// maxSendfileSize is the largest chunk size we ask the kernel to copy at a time.
	maxSendfileSize int = 4 << 20

	// maxBufferSize is the largest chunk size we ask the buffer to copy at a time.
	maxBufferSize int = 64 << 10
)

func sendFile(conn net.Conn, src int, pos, remain int64, maxSize int) (written int64, err error) {
	var b []byte
	for remain > 0 {
		n := maxSize
		if int(remain) < maxSize {
			n = int(remain)
		}
		offset := mmap2.Offset(pos)
		if offset < pos {
			pos = int64(pos - offset)
		}
		b, err = mmap2.Open(src, offset, int(pos)+n, mmap2.READ)
		if err != nil {
			return
		}
		n, errno := conn.Write(b[pos : pos+int64(n)])
		_ = mmap2.Munmap(b)
		if n > 0 {
			pos += int64(n)
			written += int64(n)
			remain -= int64(n)
		} else if (n == 0 && errno == nil) || (errno != nil && errno != syscall.EAGAIN) {
			err = errno
			break
		}
	}
	return written, err
}
