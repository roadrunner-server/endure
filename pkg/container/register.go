package endure

import (
	"github.com/roadrunner-server/errors"
)

func (e *Endure) register(name string, vrtx interface{}) error {
	// check the vertex
	const op = errors.Op("endure_register")
	if e.graph.HasVertex(name) {
		return errors.E(op, errors.Traverse, errors.Errorf("vertex `%s` already exists", name))
	}

	// just push the vertex
	// here we can append in future some meta information
	e.graph.AddVertex(name, vrtx)
	return nil
}
