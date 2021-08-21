package remote

import (
	"encoding/binary"
	"net"
	"net/http"
)

const (
	XRealIP       = "X-Real-IP"
	XForwardedFor = "X-Forwarded-For"
)

func RemoteAddr(req *http.Request) (addr string) {
	addr = req.RemoteAddr
	if ip := req.Header.Get(XRealIP); ip != "" {
		addr = ip
	} else if ip = req.Header.Get(XForwardedFor); ip != "" {
		addr = ip
	} else {
		var err error
		addr, _, err = net.SplitHostPort(addr)
		if err != nil {
			return ""
		}
	}
	if addr == "::1" {
		addr = "127.0.0.1"
	}
	return
}

func IpTo4(addr string) uint32 {
	ip := net.ParseIP(addr)
	if ip == nil {
		return 0
	}
	ip = ip.To4()
	return binary.BigEndian.Uint32(ip)
}
