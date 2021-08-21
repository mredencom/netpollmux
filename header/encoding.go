package header

import (
	"net/http"
	"strings"
)

func CheckAcceptEncoding(r *http.Request, compressType string) bool {
	if strings.Contains(GetRequestHeader(r, AcceptEncoding), compressType) {
		return true
	}
	return false
}

func SetContentEncoding(w http.ResponseWriter, compressType string) {
	SetHeader(w, ContentEncoding, compressType)
	SetHeader(w, Vary, AcceptEncoding)
	DelHeader(w, ContentLength)
}
