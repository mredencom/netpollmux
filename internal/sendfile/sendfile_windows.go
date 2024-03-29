//go:build windows
// +build windows

package sendfile

// SendFile wraps the sendfile system call.
func SendFile(conn net.Conn, src int, pos, remain int64) (written int64, err error) {
	return sendFile(conn, src, pos, remain, maxSendfileSize)
}
