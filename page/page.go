package page

import (
	"html/template"
	"io"
	"os"
)

type Tmpl struct {
	template *template.Template
}

func (t Tmpl) Execute(w io.Writer, name string, data interface{}) error {
	return t.template.ExecuteTemplate(w, name, data)
}

func (t Tmpl) Generate(fp string, name string, data interface{}) error {
	file, err := os.Create(fp)
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, name, data)
}

func New(files []string) *Tmpl {
	return &Tmpl{
		template: template.Must(template.ParseFiles(files...)),
	}
}

type RaceData struct {
	Name      string
	StartDate string
	Distance  string
	Pace      string
	Time      string
	MapboxUrl string
	StaticUrl string
}
