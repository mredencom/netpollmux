//go:build !linux && !darwin && !dragonfly && !freebsd && !netbsd && !openbsd
// +build !linux,!darwin,!dragonfly,!freebsd,!netbsd,!openbsd

package netpollmux

import (
	"errors"
	"time"
)

// description is the poll type.
const description = "none"

// Poll represents the poll that supports non-blocking I/O on file descriptors with polling.
type Poll struct {
}

// Create creates a new poll.
func Create() (*Poll, error) {
	return nil, errors.New("system not supported")
}

// SetTimeout sets the wait timeout.
func (p *Poll) SetTimeout(d time.Duration) (err error) {
	return nil
}

// Register registers a file descriptor.
func (p *Poll) Register(fd int) (err error) {
	return
}

// Write adds a write event.
func (p *Poll) Write(fd int) (err error) {
	return
}

// Unregister unregisters a file descriptor.
func (p *Poll) Unregister(fd int) (err error) {
	return
}

// Wait waits events.
func (p *Poll) Wait(events []Event) (n int, err error) {
	return
}

// Close closes the poll fd. The underlying file descriptor is closed by the
// destroy method when there are no remaining references.
func (p *Poll) Close() error {
	return nil
}
