package endure

import (
	"github.com/spiral/endure/pkg/vertex"
	"github.com/spiral/errors"
)

func (e *Endure) register(name string, vrtx interface{}, order int) error {
	// check the vertex
	const op = errors.Op("internal_register")
	if e.graph.HasVertex(name) {
		return errors.E(op, errors.Traverse, errors.Errorf("vertex `%s` already exists", name))
	}

	meta := vertex.Meta{
		Order: order,
	}

	// just push the vertex
	// here we can append in future some meta information
	e.graph.AddVertex(name, vrtx, meta)
	return nil
}
