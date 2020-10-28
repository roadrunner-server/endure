// +build windows

package structures

import (
	"github.com/spiral/errors"
)

func Visualize(vertices []*Vertex) error {
	const op = errors.Op("print_graph")
	return errors.E(op, errors.Unsupported, errors.Str("windows currently not supported for this feature"))
}
