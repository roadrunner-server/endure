package endure

import (
	"github.com/spiral/errors"
)

func (e *Endure) register(name string, vertex interface{}, order int) error {
	// check the vertex
	const op = errors.Op("internal_register")
	if e.graph.HasVertex(name) {
		return errors.E(op, errors.Traverse, errors.Errorf("vertex `%s` already exists", name))
	}

	meta := Meta{
		Order: order,
	}

	// just push the vertex
	// here we can append in future some meta information
	e.graph.AddVertex(name, vertex, meta)
	return nil
}
