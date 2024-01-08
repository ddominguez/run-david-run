package page

import (
	"html/template"
	"io"
)

type Page struct {
	files []string
}

func (p Page) Template() *template.Template {
	return template.Must(template.ParseFiles(p.files...))
}

func (p Page) Render(w io.Writer, name string, data interface{}) error {
	t := p.Template()
	return t.ExecuteTemplate(w, name, data)
}

func New(files []string) Page {
	return Page{files: files}
}
