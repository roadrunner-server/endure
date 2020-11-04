package endure

import (
	"bytes"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/spiral/errors"
)

type Output string

const (
	File   Output = "file"
	StdOut Output = "stdout"
	Empty  Output = "empty"
)

var _graphTmpl = template.Must(
	template.New("DotGraph").
		Funcs(template.FuncMap{
			"quote": strconv.Quote,
		}).
		Parse(`digraph endure {
	rankdir=TB;
	graph [compound=true];
	{{range $v := .}}
		{{range $g := $v.Dependencies}}
		{{quote $v.ID}} -> {{quote $g.ID}};
		{{end}}
	{{end}}
}`))

func (e *Endure) Visualize(vertices []*Vertex) error {
	const op = errors.Op("print_graph")
	f := new(bytes.Buffer)
	err := _graphTmpl.Execute(f, vertices)
	if err != nil {
		return errors.E(op, err)
	}
	// clear on exit
	defer f.Reset()

	// clean up
	strSl := strings.Split(f.String(), "\n")
	// remove old data from buffer
	f.Reset()
	// set last string, because vertices are not distinct
	var last string

	for i := 0; i < len(strSl); i++ {
		if strSl[i] == last {
			// skip
			continue
		}
		if strSl[i] == "\t" || strSl[i] == "\t\t" {
			// skip tabs
			continue
		}
		last = strSl[i]
		f.WriteString(strSl[i] + "\n")
	}

	switch e.output {
	case File:
		if e.path == "" {
			return errors.E(op, errors.Str("path not provided"))
		}
		file, err := os.OpenFile(e.path, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			return errors.E(op, err)
		}
		_, err = file.WriteString(f.String())
		if err != nil {
			return errors.E(op, err)
		}
		return nil
	case StdOut:
		_, err = os.Stdout.WriteString(f.String())
		if err != nil {
			return errors.E(op, err)
		}
		return nil
	default:
		return errors.E(op, errors.Str("unknown output"))
	}
}
