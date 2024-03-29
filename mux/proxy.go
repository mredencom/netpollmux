package mux

import (
	"github.com/php2go/netpollmux/internal/logger"
	"net/http"
	"net/http/httputil"
	"net/url"
)

const MaxConnPerHost = 16384

var transport *http.Transport

func init() {
	transport = &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		MaxIdleConnsPerHost: MaxConnPerHost,
	}
}

func Proxy(w http.ResponseWriter, r *http.Request, targetUrl string) {
	targetUrlParse, err := url.Parse(targetUrl)
	if err != nil {
		logger.Fatal(err)
	}
	target, err := url.Parse(targetUrlParse.Scheme + "://" + targetUrlParse.Host)
	if err != nil {
		logger.Fatal(err)
	}
	r.URL.Path = targetUrlParse.Path
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = transport
	proxy.ServeHTTP(w, r)
}
