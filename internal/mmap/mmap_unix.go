//go:build darwin || linux || dragonfly || freebsd || netbsd || openbsd
// +build darwin linux dragonfly freebsd netbsd openbsd

package mmap

import (
	"os"
	"syscall"
	"unsafe"
)

const (
	ProtRead  = syscall.PROT_READ
	ProtWrite = syscall.PROT_WRITE
	ProtExec  = syscall.PROT_EXEC

	MapShared  = syscall.MAP_SHARED
	MapPrivate = syscall.MAP_PRIVATE
	MapCopy    = MapPrivate
)

// Offset returns the valid offset.
func Offset(offset int64) int64 {
	pageSize := int64(os.Getpagesize())
	return offset / pageSize * pageSize
}

func protFlags(p Prot) (prot int, flags int) {
	prot = ProtRead
	flags = MapShared
	if p&WRITE != 0 {
		prot |= ProtWrite
	}
	if p&COPY != 0 {
		flags = MapCopy
	}
	if p&EXEC != 0 {
		prot |= ProtExec
	}
	return
}

func mMap(fd int, offset int64, length int, prot int, flags int) (data []byte, err error) {
	return syscall.Mmap(fd, offset, length, prot, flags)
}

func mSync(b []byte) (err error) {
	_, _, errno := syscall.Syscall(syscall.SYS_MSYNC, uintptr(unsafe.Pointer(&b[0])), uintptr(len(b)), syscall.MS_SYNC)
	if errno != 0 {
		err = syscall.Errno(errno)
	}
	return
}

func mUnmap(b []byte) (err error) {
	return syscall.Munmap(b)
}
