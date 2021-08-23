package buffer

import (
	"testing"
)

func TestPageSize(t *testing.T) {
	b := NewBuffers(0)
	pageSizes := []int{0, 1, 512, 1023, 1024, 1024}
	sizes := []int{1024, 1024, 1024, 1024, 1024, 1025}
	results := []int{1024, 1024, 1024, 2046, 1024, 2048}
	if len(pageSizes) != len(sizes) && len(sizes) != len(results) {
		t.Error()
	}
	for i := 0; i < len(pageSizes); i++ {
		b.pageSize = pageSizes[i]
		size := sizes[i]
		if b.pageSize > 0 && size%b.pageSize > 0 {
			size = size/b.pageSize*b.pageSize + b.pageSize
		}
		if size != results[i] {
			t.Error(i, size, results[i])
		}
	}
}

func TestAssignPool(t *testing.T) {
	defaultBuffers = NewBuffers(1024)
	for i := 0; i < 4; i++ {
		size := 64*1024 + i
		p := AssignPool(size)
		if p.size < size {
			t.Error(p.size)
		}
		buf := GetBuffer(size)
		if len(buf) < size {
			t.Error(len(buf))
		}
		PutBuffer(buf)
	}
}

func BenchmarkAssignPool(b *testing.B) {
	bs := NewBuffers(pageSize)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		bs.AssignPool(size)
	}
}

func BenchmarkAssignPoolPageSize2(b *testing.B) {
	bs := NewBuffers(2)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		bs.AssignPool(size)
	}
}

func BenchmarkAssignPoolPageSize4(b *testing.B) {
	bs := NewBuffers(4)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		bs.AssignPool(size)
	}
}

func BenchmarkAssignPoolPageSize8(b *testing.B) {
	bs := NewBuffers(8)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		bs.AssignPool(size)
	}
}

func BenchmarkAssignPoolPageSize16(b *testing.B) {
	bs := NewBuffers(16)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		bs.AssignPool(size)
	}
}

func BenchmarkAssignPoolPageSize32(b *testing.B) {
	bs := NewBuffers(32)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		bs.AssignPool(size)
	}
}

func BenchmarkAssignPoolPageSize64(b *testing.B) {
	bs := NewBuffers(64)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		bs.AssignPool(size)
	}
}

func BenchmarkAssignPoolPageSize128(b *testing.B) {
	bs := NewBuffers(128)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		bs.AssignPool(size)
	}
}

func BenchmarkAssignPoolPageSize256(b *testing.B) {
	bs := NewBuffers(256)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		bs.AssignPool(size)
	}
}

func BenchmarkAssignPoolPageSize512(b *testing.B) {
	bs := NewBuffers(512)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		bs.AssignPool(size)
	}
}

func BenchmarkAssignPoolPageSize1024(b *testing.B) {
	bs := NewBuffers(1024)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		bs.AssignPool(size)
	}
}

func BenchmarkAssignPoolPageSize2048(b *testing.B) {
	bs := NewBuffers(2048)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		bs.AssignPool(size)
	}
}

func BenchmarkAssignPoolPageSize4096(b *testing.B) {
	bs := NewBuffers(4096)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		bs.AssignPool(size)
	}
}

func BenchmarkAssignPoolPageSize8192(b *testing.B) {
	bs := NewBuffers(8192)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		bs.AssignPool(size)
	}
}

func BenchmarkAssignPoolPageSize16384(b *testing.B) {
	bs := NewBuffers(16384)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		bs.AssignPool(size)
	}
}

func BenchmarkAssignSizedPool(b *testing.B) {
	bs := NewBuffers(pageSize)
	size := 64 * 1024
	bs.AssignPool(size)
	for i := 0; i < b.N; i++ {
		bs.AssignPool(size)
	}
}

func BenchmarkBuffers(b *testing.B) {
	bs := NewBuffers(0)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		buf := bs.AssignPool(size).GetBufferSize(size)
		bs.AssignPool(size).PutBuffer(buf)
	}
}

func BenchmarkBuffersPageSize2(b *testing.B) {
	bs := NewBuffers(2)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		buf := bs.AssignPool(size).GetBufferSize(size)
		bs.AssignPool(size).PutBuffer(buf)
	}
}

func BenchmarkBuffersPageSize4(b *testing.B) {
	bs := NewBuffers(4)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		buf := bs.AssignPool(size).GetBufferSize(size)
		bs.AssignPool(size).PutBuffer(buf)
	}
}

func BenchmarkBuffersPageSize8(b *testing.B) {
	bs := NewBuffers(8)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		buf := bs.AssignPool(size).GetBufferSize(size)
		bs.AssignPool(size).PutBuffer(buf)
	}
}

func BenchmarkBuffersPageSize16(b *testing.B) {
	bs := NewBuffers(16)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		buf := bs.AssignPool(size).GetBufferSize(size)
		bs.AssignPool(size).PutBuffer(buf)
	}
}

func BenchmarkBuffersPageSize32(b *testing.B) {
	bs := NewBuffers(32)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		buf := bs.AssignPool(size).GetBufferSize(size)
		bs.AssignPool(size).PutBuffer(buf)
	}
}

func BenchmarkBuffersPageSize64(b *testing.B) {
	bs := NewBuffers(64)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		buf := bs.AssignPool(size).GetBufferSize(size)
		bs.AssignPool(size).PutBuffer(buf)
	}
}

func BenchmarkBuffersPageSize128(b *testing.B) {
	bs := NewBuffers(128)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		buf := bs.AssignPool(size).GetBufferSize(size)
		bs.AssignPool(size).PutBuffer(buf)
	}
}

func BenchmarkBuffersPageSize256(b *testing.B) {
	bs := NewBuffers(256)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		buf := bs.AssignPool(size).GetBufferSize(size)
		bs.AssignPool(size).PutBuffer(buf)
	}
}

func BenchmarkBuffersPageSize512(b *testing.B) {
	bs := NewBuffers(512)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		buf := bs.AssignPool(size).GetBufferSize(size)
		bs.AssignPool(size).PutBuffer(buf)
	}
}

func BenchmarkBuffersPageSize1024(b *testing.B) {
	bs := NewBuffers(1024)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		buf := bs.AssignPool(size).GetBufferSize(size)
		bs.AssignPool(size).PutBuffer(buf)
	}
}

func BenchmarkBuffersPageSize2048(b *testing.B) {
	bs := NewBuffers(2048)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		buf := bs.AssignPool(size).GetBufferSize(size)
		bs.AssignPool(size).PutBuffer(buf)
	}
}

func BenchmarkBuffersPageSize4096(b *testing.B) {
	bs := NewBuffers(4096)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		buf := bs.AssignPool(size).GetBufferSize(size)
		bs.AssignPool(size).PutBuffer(buf)
	}
}

func BenchmarkBuffersPageSize8192(b *testing.B) {
	bs := NewBuffers(8192)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		buf := bs.AssignPool(size).GetBufferSize(size)
		bs.AssignPool(size).PutBuffer(buf)
	}
}

func BenchmarkBuffersPageSize16384(b *testing.B) {
	bs := NewBuffers(16384)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		buf := bs.AssignPool(size).GetBufferSize(size)
		bs.AssignPool(size).PutBuffer(buf)
	}
}

func BenchmarkAssignPoolAndBuffers(b *testing.B) {
	bs := NewBuffers(0)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		p := bs.AssignPool(size)
		buf := p.GetBufferSize(size)
		p.PutBuffer(buf)
	}
}

func BenchmarkAssignPoolAndBuffersPageSize2(b *testing.B) {
	bs := NewBuffers(2)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		p := bs.AssignPool(size)
		buf := p.GetBufferSize(size)
		p.PutBuffer(buf)
	}
}

func BenchmarkAssignPoolAndBuffersPageSize4(b *testing.B) {
	bs := NewBuffers(4)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		p := bs.AssignPool(size)
		buf := p.GetBufferSize(size)
		p.PutBuffer(buf)
	}
}

func BenchmarkAssignPoolAndBuffersPageSize8(b *testing.B) {
	bs := NewBuffers(8)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		p := bs.AssignPool(size)
		buf := p.GetBufferSize(size)
		p.PutBuffer(buf)
	}
}

func BenchmarkAssignPoolAndBuffersPageSize16(b *testing.B) {
	bs := NewBuffers(16)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		p := bs.AssignPool(size)
		buf := p.GetBufferSize(size)
		p.PutBuffer(buf)
	}
}

func BenchmarkAssignPoolAndBuffersPageSize32(b *testing.B) {
	bs := NewBuffers(32)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		p := bs.AssignPool(size)
		buf := p.GetBufferSize(size)
		p.PutBuffer(buf)
	}
}

func BenchmarkAssignPoolAndBuffersPageSize64(b *testing.B) {
	bs := NewBuffers(64)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		p := bs.AssignPool(size)
		buf := p.GetBufferSize(size)
		p.PutBuffer(buf)
	}
}

func BenchmarkAssignPoolAndBuffersPageSize128(b *testing.B) {
	bs := NewBuffers(128)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		p := bs.AssignPool(size)
		buf := p.GetBufferSize(size)
		p.PutBuffer(buf)
	}
}

func BenchmarkAssignPoolAndBuffersPageSize256(b *testing.B) {
	bs := NewBuffers(256)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		p := bs.AssignPool(size)
		buf := p.GetBufferSize(size)
		p.PutBuffer(buf)
	}
}

func BenchmarkAssignPoolAndBuffersPageSize512(b *testing.B) {
	bs := NewBuffers(512)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		p := bs.AssignPool(size)
		buf := p.GetBufferSize(size)
		p.PutBuffer(buf)
	}
}

func BenchmarkAssignPoolAndBuffersPageSize1024(b *testing.B) {
	bs := NewBuffers(1024)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		p := bs.AssignPool(size)
		buf := p.GetBufferSize(size)
		p.PutBuffer(buf)
	}
}

func BenchmarkAssignPoolAndBuffersPageSize2048(b *testing.B) {
	bs := NewBuffers(2048)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		p := bs.AssignPool(size)
		buf := p.GetBufferSize(size)
		p.PutBuffer(buf)
	}
}

func BenchmarkAssignPoolAndBuffersPageSize4096(b *testing.B) {
	bs := NewBuffers(4096)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		p := bs.AssignPool(size)
		buf := p.GetBufferSize(size)
		p.PutBuffer(buf)
	}
}

func BenchmarkAssignPoolAndBuffersPageSize8192(b *testing.B) {
	bs := NewBuffers(8192)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		p := bs.AssignPool(size)
		buf := p.GetBufferSize(size)
		p.PutBuffer(buf)
	}
}

func BenchmarkAssignPoolAndBuffersPageSize16384(b *testing.B) {
	bs := NewBuffers(16384)
	for i := 0; i < b.N; i++ {
		size := i % (64 * 1024)
		p := bs.AssignPool(size)
		buf := p.GetBufferSize(size)
		p.PutBuffer(buf)
	}
}

func BenchmarkSizedBuffer(b *testing.B) {
	bs := NewBuffers(pageSize)
	size := 64 * 1024
	p := bs.AssignPool(size)
	for i := 0; i < b.N; i++ {
		buf := p.GetBufferSize(size)
		p.PutBuffer(buf)
	}
}
