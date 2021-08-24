package mux

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

const (
	OPTIONS = 1 << iota
	GET
	HEAD
	POST
	PUT
	DELETE
	TRACE
	CONNECT
	PATCH
)

// ErrGroupExisted is the error returned by Group when registers a existed group.
var ErrGroupExisted = errors.New("Group Existed")

// ErrParamsKeyEmpty is the error returned by HandleFunc when the params key is empty.
var ErrParamsKeyEmpty = errors.New("Params key must be not empty")

// contextKey is a key for use with context.WithValue. It's used as
// a pointer so it fits in an interface{} without allocation.
type contextKey struct {
	name string
}

// String returns a context key.
func (k *contextKey) String() string { return "context key " + k.name }

// RecoveryContextKey is a context key.
var RecoveryContextKey = &contextKey{"recovery"}

// Route is an HTTP request multiplexer.
type Route struct {
	mut         sync.RWMutex
	prefixes    map[string]*prefix
	middlewares []http.Handler
	recovery    http.Handler
	notFound    http.Handler
	group       string
	groups      map[string]*Route
}

type prefix struct {
	prefix string
	m      map[string]*Entry
}

// Entry represents an HTTP HandlerFunc entry.
type Entry struct {
	handler                                                      http.Handler
	key                                                          string
	match                                                        []string
	params                                                       map[string]string
	method                                                       int
	get, post, put, delete, patch, head, options, trace, connect http.Handler
}

// NewRoute returns a new NewRoute.
func NewRoute() *Route {
	return &Route{
		prefixes: make(map[string]*prefix),
		groups:   make(map[string]*Route),
	}
}

// newGroup create a route group,require prefix
func newGroup(group string) *Route {
	return &Route{
		prefixes: make(map[string]*prefix),
		groups:   make(map[string]*Route),
		group:    group,
	}
}

// ServeHTTP dispatches the request to the handler whose
// pattern most closely matches the request URL.
func (m *Route) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := m.replace(r.URL.Path)
	m.mut.RLock()
	entry := m.searchEntry(path, w, r)
	m.mut.RUnlock()
	if entry != nil {
		m.serveEntry(entry, w, r)
		return
	}
	if m.notFound != nil {
		m.notFound.ServeHTTP(w, r)
		return
	}
	http.Error(w, "404 Not Found : "+r.URL.String(), http.StatusNotFound)
}

// searchEntry search a route to the entry
func (m *Route) searchEntry(path string, w http.ResponseWriter, r *http.Request) *Entry {
	if entry := m.getHandlerFunc(path); entry != nil {
		return entry
	}
	for _, group := range m.groups {
		if entry := group.searchEntry(path, w, r); entry != nil {
			return entry
		}
	}
	return nil
}

// serveEntry find the corresponding service method
func (m *Route) serveEntry(entry *Entry, w http.ResponseWriter, r *http.Request) {
	if entry.method == 0 {
		m.serveHandler(entry.handler, w, r)
	} else if r.Method == "GET" && entry.method&GET > 0 {
		m.serveHandler(entry.get, w, r)
	} else if r.Method == "POST" && entry.method&POST > 0 {
		m.serveHandler(entry.post, w, r)
	} else if r.Method == "PUT" && entry.method&PUT > 0 {
		m.serveHandler(entry.put, w, r)
	} else if r.Method == "DELETE" && entry.method&DELETE > 0 {
		m.serveHandler(entry.delete, w, r)
	} else if r.Method == "PATCH" && entry.method&PATCH > 0 {
		m.serveHandler(entry.patch, w, r)
	} else if r.Method == "HEAD" && entry.method&HEAD > 0 {
		m.serveHandler(entry.head, w, r)
	} else if r.Method == "OPTIONS" && entry.method&OPTIONS > 0 {
		m.serveHandler(entry.options, w, r)
	} else if r.Method == "TRACE" && entry.method&TRACE > 0 {
		m.serveHandler(entry.trace, w, r)
	} else if r.Method == "CONNECT" && entry.method&CONNECT > 0 {
		m.serveHandler(entry.connect, w, r)
	} else {
		m.serveHandler(m.notFound, w, r)
	}
}

// Recovery returns a recovery handler function that recovers from any panics and writes a 500 status code.
func Recovery(w http.ResponseWriter, r *http.Request) {
	err := r.Context().Value(RecoveryContextKey)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = fmt.Fprintf(w, "500 Internal Server Error : %v\n", err)
}

func (m *Route) serveHandler(handler http.Handler, w http.ResponseWriter, r *http.Request) {
	if m.recovery != nil {
		defer func() {
			if err := recover(); err != nil {
				ctx := context.WithValue(r.Context(), RecoveryContextKey, err)
				m.recovery.ServeHTTP(w, r.WithContext(ctx))
			}
		}()
	}
	m.middleware(w, r)
	if handler != nil {
		handler.ServeHTTP(w, r)
	}
}

// getHandlerFunc return according to the path  Entry
func (m *Route) getHandlerFunc(path string) *Entry {
	if prefix, key, ok := m.matchParams(path); ok {
		if entry, ok := m.prefixes[prefix].m[key]; ok {
			return entry
		}
	}
	return nil
}

// HandleFunc registers a handler function with the given pattern to the Mux.
func (m *Route) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) *Entry {
	return m.Handle(pattern, http.HandlerFunc(handler))
}

// Handle registers a handler with the given pattern to the Mux.
func (m *Route) Handle(pattern string, handler http.Handler) *Entry {
	m.mut.Lock()
	defer m.mut.Unlock()
	pattern = m.replace(pattern)
	pre, key, match, params := m.parseParams(m.group + pattern)
	if v, ok := m.prefixes[pre]; ok {
		if entry, ok := v.m[key]; ok {
			entry.handler = handler
			entry.key = key
			entry.match = match
			entry.params = params
			m.prefixes[pre].m[key] = entry
			return entry
		}
		entry := &Entry{}
		entry.handler = handler
		entry.key = key
		entry.match = match
		entry.params = params
		m.prefixes[pre].m[key] = entry
		return entry
	}
	m.prefixes[pre] = &prefix{m: make(map[string]*Entry), prefix: pre}
	entry := &Entry{}
	entry.handler = handler
	entry.key = key
	entry.match = match
	entry.params = params
	m.prefixes[pre].m[key] = entry
	return entry
}

// Group registers a group with the given pattern to the Mux.
func (m *Route) Group(group string, f func(m *Route)) {
	m.mut.Lock()
	defer m.mut.Unlock()
	group = m.replace(group)
	groupMux := newGroup(group)
	f(groupMux)
	if _, ok := m.groups[group]; ok {
		panic(ErrGroupExisted)
	}
	groupMux.middlewares = m.middlewares
	m.groups[group] = groupMux
}

// NotFound registers a not found handler function to the Mux.
func (m *Route) NotFound(handler http.HandlerFunc) {
	m.mut.Lock()
	defer m.mut.Unlock()
	m.notFound = handler
}

// Recovery registers a recovery handler function to the Mux.
func (m *Route) Recovery(handler http.HandlerFunc) {
	m.mut.Lock()
	defer m.mut.Unlock()
	m.recovery = handler
}

// Use uses middleware.
func (m *Route) Use(handlers ...http.HandlerFunc) {
	m.mut.Lock()
	defer m.mut.Unlock()
	for _, handler := range handlers {
		m.middlewares = append(m.middlewares, handler)
	}
}

func (m *Route) middleware(w http.ResponseWriter, r *http.Request) {
	for _, handler := range m.middlewares {
		handler.ServeHTTP(w, r)
	}
}

// Params returns http request params.
func (m *Route) Params(r *http.Request) map[string]string {
	params := make(map[string]string)
	path := m.replace(r.URL.Path)
	m.mut.RLock()
	defer m.mut.RUnlock()
	if prefix, key, ok := m.matchParams(path); ok {
		if entry, ok := m.prefixes[prefix].m[key]; ok &&
			len(entry.match) > 0 && len(path) > len(prefix) {
			strs := strings.Split(path[len(prefix):], "/")
			if len(strs) == len(entry.match) {
				for i := 0; i < len(strs); i++ {
					if entry.match[i] != "" {
						params[entry.match[i]] = strs[i]
					}
				}
			}
		}
	}
	return params
}

// matchParams match route params
func (m *Route) matchParams(path string) (string, string, bool) {
	var key string
	for _, p := range m.prefixes {
		if strings.HasPrefix(path, p.prefix) {
			r := path[len(p.prefix):]
			if r == "" {
				return p.prefix, "", true
			}
			for _, v := range p.m {
				count := strings.Count(r, "/")
				if count+1 == len(v.match) {
					form := strings.Split(r, "/")
					for i := 0; i < len(form); i++ {
						if v.match[i] != "" {
							if i > 0 {
								key += "/:"
							} else {
								key += ":"
							}
						} else {
							key += "/" + form[i]
						}
					}
					if key == v.key {
						return p.prefix, v.key, true
					}
				}
			}
		}
	}
	return "", "", false
}

// parseParams parse params
func (m *Route) parseParams(pattern string) (string, string, []string, map[string]string) {
	var match []string
	var key, prefix string
	params := make(map[string]string)
	if strings.Contains(pattern, ":") {
		idx := strings.Index(pattern, ":")
		prefix = pattern[:idx]
		if idx+1 == len(pattern) || strings.Contains(pattern, ":/") {
			panic(ErrParamsKeyEmpty)
		}
		match = strings.Split(pattern[idx:], "/")
		for i := 0; i < len(match); i++ {
			if strings.Contains(match[i], ":") {
				match[i] = strings.Trim(match[i], ":")
				params[match[i]] = ""
				if i > 0 {
					key += "/:"
				} else {
					key += ":"
				}
			} else {
				key += "/" + match[i]
				match[i] = ""
			}
		}
	} else {
		prefix = pattern
	}
	return prefix, key, match, params
}

// replace replace a string
func (m *Route) replace(s string) string {
	for strings.Contains(s, "//") {
		s = strings.ReplaceAll(s, "//", "/")
	}
	return s
}

// GET adds a GET HTTP method to the entry.
func (entry *Entry) GET() *Entry {
	entry.method |= GET
	entry.get = entry.handler
	return entry
}

// POST adds a POST HTTP method to the entry.
func (entry *Entry) POST() *Entry {
	entry.method |= POST
	entry.post = entry.handler
	return entry
}

// PUT adds a PUT HTTP method to the entry.
func (entry *Entry) PUT() *Entry {
	entry.method |= PUT
	entry.put = entry.handler
	return entry
}

// DELETE adds a DELETE HTTP method to the entry.
func (entry *Entry) DELETE() *Entry {
	entry.method |= DELETE
	entry.delete = entry.handler
	return entry
}

// PATCH adds a PATCH HTTP method to the entry.
func (entry *Entry) PATCH() *Entry {
	entry.method |= PATCH
	entry.patch = entry.handler
	return entry
}

// HEAD adds a HEAD HTTP method to the entry.
func (entry *Entry) HEAD() *Entry {
	entry.method |= HEAD
	entry.head = entry.handler
	return entry
}

// OPTIONS adds a OPTIONS HTTP method to the entry.
func (entry *Entry) OPTIONS() *Entry {
	entry.method |= OPTIONS
	entry.options = entry.handler
	return entry
}

// TRACE adds a TRACE HTTP method to the entry.
func (entry *Entry) TRACE() *Entry {
	entry.method |= TRACE
	entry.trace = entry.handler
	return entry
}

// CONNECT adds a CONNECT HTTP method to the entry.
func (entry *Entry) CONNECT() *Entry {
	entry.method |= CONNECT
	entry.connect = entry.handler
	return entry
}

// All adds all HTTP method to the entry.
func (entry *Entry) All() {
	entry.GET()
	entry.POST()
	entry.HEAD()
	entry.OPTIONS()
	entry.PUT()
	entry.PATCH()
	entry.DELETE()
	entry.TRACE()
	entry.CONNECT()
}
