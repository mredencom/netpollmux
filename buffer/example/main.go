package main

import "netpollmux/buffer"

func main() {
	buffers := buffer.NewBuffers(1024)
	size := 65536

	buf := buffers.GetBuffer(size)
	buffers.PutBuffer(buf)

	p := buffers.AssignPool(size)
	buf = p.GetBuffer(size)
	p.PutBuffer(buf)
}
