package endure

import (
	"github.com/roadrunner-server/errors"
)

func (e *Endure) collects() error {
	vertices := e.graph.TopologicalOrder()

	for i := range vertices {
		if !vertices[i].IsActive() {
			continue
		}

		if _, ok := vertices[i].Plugin().(Collector); !ok {
			continue
		}

		// in deps
		collects := vertices[i].Plugin().(Collector).Collects()

		// get vals
		for j := range collects {
			impl := e.registar.ImplementsExcept(collects[j].Type, vertices[i].Plugin())
			if len(impl) == 0 {
				continue
			}

			for k := range impl {
				value, ok := e.registar.TypeValue(impl[k].Plugin(), collects[j].Type)
				if !ok {
					return errors.E("this is likely a bug, nil value from the implements. Value should be initialized due to the topological order")
				}

				// call user's callback
				collects[j].Callback(value.Interface())
			}
		}
	}

	return nil
}
