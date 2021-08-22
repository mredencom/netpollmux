//go:build windows
// +build windows

package mmap

import (
	"os"
	"sync"
	"syscall"
	"unsafe"
)

const (
	PAGE_READONLY          = syscall.PAGE_READONLY
	PAGE_READWRITE         = syscall.PAGE_READWRITE
	PAGE_WRITECOPY         = syscall.PAGE_WRITECOPY
	PAGE_EXECUTE_READ      = syscall.PAGE_EXECUTE_READ
	PAGE_EXECUTE_READWRITE = syscall.PAGE_EXECUTE_READWRITE
	PAGE_EXECUTE_WRITECOPY = syscall.PAGE_EXECUTE_WRITECOPY

	FILE_MAP_COPY    = syscall.FILE_MAP_COPY
	FILE_MAP_WRITE   = syscall.FILE_MAP_WRITE
	FILE_MAP_READ    = syscall.FILE_MAP_READ
	FILE_MAP_EXECUTE = syscall.FILE_MAP_EXECUTE
)

// Offset returns the valid offset.
func Offset(offset int64) int64 {
	pageSize := int64(os.Getpagesize() * 16)
	return offset / pageSize * pageSize
}

func protFlags(p Prot) (prot int, flags int) {
	prot = PAGE_READONLY
	flags = FILE_MAP_READ
	if p&WRITE != 0 {
		prot = PAGE_READWRITE
		flags = FILE_MAP_WRITE
	}
	if p&COPY != 0 {
		prot = PAGE_WRITECOPY
		flags = FILE_MAP_COPY
	}
	if p&EXEC != 0 {
		prot <<= 4
		flags |= FILE_MAP_EXECUTE
	}
	return
}

type mmapper struct {
	sync.Mutex
	active map[*byte][]byte
}

func (m *mmapper) Mmap(fd int, offset int64, length int, prot int, flags int) (data []byte, err error) {
	if length <= 0 {
		return nil, syscall.EINVAL
	}
	handle, err := syscall.CreateFileMapping(syscall.Handle(fd), nil, uint32(prot), uint32((offset+int64(length))>>32), uint32((offset+int64(length))&0xFFFFFFFF), nil)
	if err != nil {
		return nil, err
	}

	addr, err := syscall.MapViewOfFile(handle, uint32(flags), uint32(offset>>32), uint32(offset&0xFFFFFFFF), uintptr(length))
	if err != nil {
		return nil, err
	}
	err = syscall.CloseHandle(syscall.Handle(handle))
	if err != nil {
		return nil, err
	}
	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{addr, length, length}
	b := *(*[]byte)(unsafe.Pointer(&sl))
	p := &b[cap(b)-1]
	m.Lock()
	defer m.Unlock()
	m.active[p] = b
	return b, nil
}

func (m *mmapper) Msync(b []byte) (err error) {
	slice := (*struct {
		addr uintptr
		len  int
		cap  int
	})(unsafe.Pointer(&b))
	p := &b[cap(b)-1]
	m.Lock()
	data := m.active[p]
	m.Unlock()
	if data == nil || &b[0] != &data[0] {
		return syscall.EINVAL
	}
	return syscall.FlushViewOfFile(slice.addr, uintptr(slice.len))
}

func (m *mmapper) Munmap(data []byte) (err error) {
	if len(data) == 0 || len(data) != cap(data) {
		return syscall.EINVAL
	}
	p := &data[cap(data)-1]
	m.Lock()
	b := m.active[p]
	m.Unlock()
	if b == nil || &b[0] != &data[0] {
		return syscall.EINVAL
	}
	err = syscall.UnmapViewOfFile(uintptr(unsafe.Pointer(&b[0])))
	if err != nil {
		return err
	}
	m.Lock()
	delete(m.active, p)
	m.Unlock()
	return nil
}

var mapper = &mmapper{
	active: make(map[*byte][]byte),
}

func mmap(fd int, offset int64, length int, prot int, flags int) (data []byte, err error) {
	return mapper.Mmap(fd, offset, length, prot, flags)
}

func msync(b []byte) (err error) {
	return mapper.Msync(b)
}

func munmap(b []byte) (err error) {
	return mapper.Munmap(b)
}
