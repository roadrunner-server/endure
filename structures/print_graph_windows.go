// +build windows

package structures

import (
	"github.com/spiral/endure/errors"
)

func PrintGraph(vertices []*Vertex) error {
	const op = errors.Op("print_graph")
	return errors.E(op, errors.Unsupported, errors.Str("windows currently not supported for this feature"))
}
