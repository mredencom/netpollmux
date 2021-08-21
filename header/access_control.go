package header

import (
	"net/http"
	"strconv"
	"strings"
)

func AllowOrigin(w http.ResponseWriter, value string) {
	SetHeader(w, AccessControlAllowOrigin, value)
}

func AllowOriginAll(w http.ResponseWriter) {
	SetHeader(w, AccessControlAllowOrigin, "*")
}

func AllowCredentials(w http.ResponseWriter) {
	SetHeader(w, AccessControlAllowCredentials, "true")
}

func MaxAge(w http.ResponseWriter, age int) {
	SetHeader(w, AccessControlMaxAge, strconv.Itoa(age))
}

func AllowMethods(w http.ResponseWriter, Methods ...string) {
	SetHeader(w, AccessControlAllowMethods, strings.Join(Methods, Comma))
}

func AllowHeaders(w http.ResponseWriter, Headers ...string) {
	SetHeader(w, AccessControlAllowHeaders, strings.Join(Headers, Comma))
}
