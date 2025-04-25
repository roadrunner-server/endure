package endure

import (
	"bytes"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/roadrunner-server/endure/v2/graph"
	"github.com/roadrunner-server/errors"
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

// Visualize visualizes the graph based on provided output value
func (e *Endure) Visualize(vertices []*graph.Vertex) error {
	const op = errors.Op("endure_visualize")
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

	for i := range strSl {
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

	_, err = os.Stdout.WriteString(f.String())
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}
