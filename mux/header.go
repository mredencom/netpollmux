package mux

import "net/http"

const (
	ContentEncoding = "Content-Encoding"
	ContentLanguage = "Content-Language"
	ContentLength   = "Content-Length"
	ContentLocation = "Content-Location"
	ContentMD5      = "Content-MD5"
	ContentRange    = "Content-Range"
	ContentType     = "Content-Type"

	AcceptRanges = "Accept-Ranges"
	Bytes        = "bytes"

	AcceptEncoding = "Accept-Encoding"
	GZIP           = "gzip"
	DEFLATE        = "deflate"

	Vary          = "Vary"
	Charset       = "charset"
	CharsetPrefix = "charset="
	UTF8          = "UTF-8"
	GB18030       = "GB18030"
	GBK           = "GBK"

	Semicolon = ";"
	Comma     = ","

	ContentTypeJSON = "application/json"
	ContentTypeXML  = "text/xml"
	ContentTypeHTML = "text/html"
	ContentTypeText = "text/plain"

	AccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	AccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	AccessControlMaxAge           = "Access-Control-Max-Age"
	AccessControlAllowMethods     = "Access-Control-Allow-Methods"
	AccessControlAllowHeaders     = "Access-Control-Allow-Headers"
)

func SetHeader(w http.ResponseWriter, key, value string) {
	if _, ok := w.Header()[key]; ok {
		w.Header().Set(key, value)
	} else {
		w.Header().Add(key, value)
	}
}

func DelHeader(w http.ResponseWriter, key string) {
	w.Header().Del(key)
}

func WriteHeader(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
}

func GetRequestHeader(r *http.Request, key string) (value string) {
	return r.Header.Get(key)
}

func GetResponseHeader(w http.ResponseWriter, key string) (value string) {
	return w.Header().Get(key)
}
