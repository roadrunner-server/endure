package endure

import (
	"github.com/roadrunner-server/errors"
)

func (e *Endure) collects() error {
	vertices := e.graph.TopologicalOrder()

	for i := 0; i < len(vertices); i++ {
		if !vertices[i].IsActive() {
			continue
		}

		if _, ok := vertices[i].Plugin().(Collector); !ok {
			continue
		}

		// in deps
		collects := vertices[i].Plugin().(Collector).Collects()

		// get vals
		for j := 0; j < len(collects); j++ {
			impl := e.registar.Implements(collects[j].Type)
			if len(impl) == 0 {
				continue
			}

			for k := 0; k < len(impl); k++ {
				value, ok := e.registar.TypeValue(impl[k].Plugin(), collects[j].Type)
				if !ok {
					return errors.E("nil value from the implements. Value should be initialized due to the topological order")
				}

				// call user's callback
				collects[j].Callback(value.Interface())
			}
		}
	}

	return nil
}
