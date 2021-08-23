package mmap

import (
	"os"
	"testing"
)

func TestMMap(t *testing.T) {
	ProtFlags(READ | WRITE | COPY | EXEC)
	name := "mmap"
	file, err := os.Create(name)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(name)
	defer file.Close()
	offset := Offset(int64(os.Getpagesize() * 64))
	size := 11
	file.Truncate(int64(size) + offset)
	file.Sync()
	if FSize(file) != size+int(offset) {
		t.Error(FSize(file), size+int(offset))
	}
	m, err := Open(Fd(file), offset, size, READ|WRITE)
	if err != nil {
		t.Error(err)
	}
	str := "Hello world"
	copy(m, []byte(str))
	if err := MSync(m); err != nil {
		t.Error(err)
	}
	buf := make([]byte, size)
	file.Seek(offset, os.SEEK_SET)
	file.Read(buf)
	if str != string(buf) {
		t.Errorf("%s!=%s", str, string(buf))
	}
	if err := MUnmap(m); err != nil {
		t.Error(err)
	}
	MSync(m)
	file.Sync()
}

func BenchmarkMMap(b *testing.B) {
	name := "mmap"
	file, err := os.Create(name)
	if err != nil {
		b.Error(err)
	}
	defer os.Remove(name)
	defer file.Close()
	size := 11
	file.Truncate(int64(size))
	file.Sync()
	m, err := Open(Fd(file), 0, FSize(file), READ|WRITE)
	if err != nil {
		b.Error(err)
	}
	str := "Hello world"
	for i := 0; i < b.N; i++ {
		copy(m, []byte(str))
		MSync(m)
		file.Sync()
	}
	if err := MUnmap(m); err != nil {
		b.Error(err)
	}
}
