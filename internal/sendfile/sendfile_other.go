//go:build !darwin && !linux && !windows && !dragonfly && !freebsd && !netbsd && !openbsd
// +build !darwin,!linux,!windows,!dragonfly,!freebsd,!netbsd,!openbsd

package sendfile

import (
	"net"
)

// SendFile wraps the sendfile system call.
func SendFile(conn net.Conn, src int, pos, remain int64) (written int64, err error) {
	return sendFile(conn, src, pos, remain, maxBufferSize)
}
