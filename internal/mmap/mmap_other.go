//go:build !darwin && !linux && !windows && !dragonfly && !freebsd && !netbsd && !openbsd
// +build !darwin,!linux,!windows,!dragonfly,!freebsd,!netbsd,!openbsd

package mmap

import (
	"errors"
	"os"
	"sync"
	"syscall"
	"unsafe"

	"github.com/php2go/netpollmux/internal/buffer"
)

var (
	buffers = sync.Map{}
	assign  int32
)

// Offset returns the valid offset.
func Offset(offset int64) int64 {
	pageSize := int64(os.Getpagesize())
	return offset / pageSize * pageSize
}

func protFlags(p Prot) (prot int, flags int) {
	return 0, 0
}

type mMapper struct {
	sync.Mutex
	active map[*byte]*f
}

type f struct {
	fd     int
	offset int64
	buf    []byte
}

func (m *mMapper) MMap(fd int, offset int64, length int, prot int, flags int) (data []byte, err error) {
	if length <= 0 {
		return nil, syscall.EINVAL
	}
	pool := buffer.AssignPool(length)
	buf := pool.GetBuffer()
	cursor, _ := syscall.Seek(fd, 0, os.SEEK_CUR)
	syscall.Seek(fd, offset, os.SEEK_SET)
	n, err := syscall.Read(fd, buf)
	syscall.Seek(fd, cursor, os.SEEK_SET)
	if err != nil {
		return nil, err
	}
	if n < length {
		return nil, errors.New("length > file size")
	}
	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{uintptr(unsafe.Pointer(&buf[0])), length, length}
	b := *(*[]byte)(unsafe.Pointer(&sl))
	p := &b[cap(b)-1]
	m.Lock()
	defer m.Unlock()
	m.active[p] = &f{fd, offset, b}
	return b, nil
}

func (m *mMapper) MSync(b []byte) (err error) {
	if len(b) == 0 || len(b) != cap(b) {
		return syscall.EINVAL
	}
	p := &b[cap(b)-1]
	m.Lock()
	f := m.active[p]
	m.Unlock()
	if f == nil || f.buf == nil || &f.buf[0] != &b[0] {
		return syscall.EINVAL
	}
	cursor, _ := syscall.Seek(f.fd, 0, os.SEEK_CUR)
	syscall.Seek(f.fd, f.offset, os.SEEK_SET)
	_, err = syscall.Write(f.fd, b)
	syscall.Seek(f.fd, cursor, os.SEEK_SET)
	return err
}

func (m *mMapper) MUnmap(data []byte) (err error) {
	if len(data) == 0 || len(data) != cap(data) {
		return syscall.EINVAL
	}
	p := &data[cap(data)-1]
	m.Lock()
	f := m.active[p]
	m.Unlock()
	if f == nil || f.buf == nil || &f.buf[0] != &data[0] {
		return syscall.EINVAL
	}
	cursor, _ := syscall.Seek(f.fd, 0, os.SEEK_CUR)
	syscall.Seek(f.fd, 0, os.SEEK_SET)
	_, err = syscall.Write(f.fd, data)
	syscall.Seek(f.fd, cursor, os.SEEK_SET)
	m.Lock()
	delete(m.active, p)
	m.Unlock()
	pool := buffer.AssignPool(cap(f.buf))
	pool.PutBuffer(f.buf)
	f = nil
	return err
}

var mapper = &mMapper{
	active: make(map[*byte]*f),
}

func mMap(fd int, offset int64, length int, prot int, flags int) (data []byte, err error) {
	return mapper.MMap(fd, offset, length, prot, flags)
}

func mSync(b []byte) (err error) {
	return mapper.Msync(b)
}

func mUnmap(b []byte) (err error) {
	return mapper.Munmap(b)
}
