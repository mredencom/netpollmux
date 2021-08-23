package mmap

import (
	"os"
)

// Prot is the protection flag.
type Prot int

const (
	// READ represents the read prot
	READ Prot = 1 << iota
	// WRITE represents the write prot
	WRITE
	// COPY represents the copy prot
	COPY
	// EXEC represents the exec prot
	EXEC
)

// Fd returns the integer file descriptor referencing the open file.
// The file descriptor is valid only until f.Close is called or f is garbage collected.
func Fd(f *os.File) int {
	return int(f.Fd())
}

// FSize returns the file size.
func FSize(f *os.File) int {
	cursor, _ := f.Seek(0, os.SEEK_CUR)
	ret, _ := f.Seek(0, os.SEEK_END)
	f.Seek(cursor, os.SEEK_SET)
	return int(ret)
}

// ProtFlags returns prot and flags by Prot p.
func ProtFlags(p Prot) (prot int, flags int) {
	return protFlags(p)
}

// Open opens a MMap
func Open(fd int, offset int64, length int, p Prot) (data []byte, err error) {
	prot, flags := ProtFlags(p)
	return MMap(fd, offset, length, prot, flags)
}

//MMap calls the MMap system call.
func MMap(fd int, offset int64, length int, prot int, flags int) (data []byte, err error) {
	return mMap(fd, offset, length, prot, flags)
}

// MSync calls the msync system call.
func MSync(b []byte) (err error) {
	return mSync(b)
}

// MUnmap calls the MUnmap system call.
func MUnmap(b []byte) (err error) {
	return mUnmap(b)
}
