package mux

import (
	"bytes"
	"errors"
	"html/template"
	"net/http"
	"sync"
)

const defaultTemplateName = "Html"

var bufRenderPool *sync.Pool

func init() {
	bufRenderPool = &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(nil)
		},
	}
}
func bufPoolPut(buf *bytes.Buffer) {
	buf.Reset()
	bufRenderPool.Put(buf)
}

type Tmpl struct {
	mut             sync.RWMutex
	defaultTemplate *template.Template
	templates       *template.Template
	render          *Render
}

func NewTmpl() *Tmpl {
	return &Tmpl{render: DefaultRender}
}

func NewTmplWithRender(render *Render) *Tmpl {
	return &Tmpl{render: render}
}

func (t *Tmpl) Parse(text string) error {
	t.mut.Lock()
	defer t.mut.Unlock()
	tmpl, err := template.New(defaultTemplateName).Parse(text)
	if err != nil {
		return err
	}
	t.defaultTemplate = tmpl
	return nil
}

func (t *Tmpl) Execute(w http.ResponseWriter, r *http.Request, data interface{}, code int) (int, error) {
	t.mut.RLock()
	defer t.mut.RUnlock()
	if t.defaultTemplate == nil {
		return 0, errors.New("Must ParseTemplate")
	}
	html, err := t.execute(data)
	if err != nil {
		return 0, err
	}
	SetContentTypeWithCharset(w, ContentTypeHTML, t.render.charset)
	return t.render.Body(w, r, []byte(html), code)
}

func (t *Tmpl) execute(data interface{}) (string, error) {
	buf := bufRenderPool.Get().(*bytes.Buffer)
	defer bufPoolPut(buf)
	if err := t.defaultTemplate.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (t *Tmpl) ParseTemplate(name, text string) error {
	t.mut.Lock()
	defer t.mut.Unlock()
	var tmpl *template.Template
	var err error
	if t.templates != nil {
		tmpl, err = t.templates.New(name).Parse(text)
	} else {
		tmpl, err = template.New(name).Parse(text)
	}
	if err != nil {
		return err
	}
	t.templates = tmpl
	return nil
}

func (t *Tmpl) ParseFiles(filenames ...string) error {
	t.mut.Lock()
	defer t.mut.Unlock()
	tmpl, err := template.ParseFiles(filenames...)
	if err != nil {
		return err
	}
	t.templates = tmpl
	return nil
}

func (t *Tmpl) ExecuteTemplate(w http.ResponseWriter, r *http.Request, name string, data interface{}, code int) (int, error) {
	t.mut.RLock()
	defer t.mut.RUnlock()
	if t.templates == nil {
		return 0, errors.New("Must ParseFiles")
	}
	html, err := t.executeTemplate(name, data)
	if err != nil {
		return 0, err
	}
	SetContentTypeWithCharset(w, ContentTypeHTML, t.render.charset)
	return t.render.Body(w, r, []byte(html), code)
}

func (t *Tmpl) executeTemplate(name string, data interface{}) (string, error) {
	buf := bufRenderPool.Get().(*bytes.Buffer)
	defer bufPoolPut(buf)
	if err := t.templates.ExecuteTemplate(buf, name, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func Execute(w http.ResponseWriter, r *http.Request, text string, data interface{}, code int) (int, error) {
	t := NewTmpl()
	err := t.Parse(text)
	if err != nil {
		return 0, err
	}
	return t.Execute(w, r, data, code)
}
