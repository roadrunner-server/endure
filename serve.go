package endure

import (
	"reflect"

	"github.com/spiral/errors"
	"go.uber.org/zap"
)

func (e *Endure) callServeFn(vertex *Vertex, in []reflect.Value) (*result, error) {
	const op = errors.Op("call_serve_fn")
	e.logger.Debug("preparing to serveInternal the vertex", zap.String("vertex id", vertex.ID))
	m, _ := reflect.TypeOf(vertex.Iface).MethodByName(ServeMethodName)
	ret := m.Func.Call(in)
	res := ret[0].Interface()
	if res != nil {
		e.logger.Debug("called serveInternal on the vertex", zap.String("vertex id", vertex.ID))
		if e, ok := res.(chan error); ok && e != nil {
			// error come right after we start serving the vertex
			if len(e) > 0 {
				return nil, errors.E(op, errors.FunctionCall, errors.Errorf("got first run error from vertex %s, stopping execution", vertex.ID))
			}
			return &result{
				errCh:    e,
				signal:   make(chan notify),
				vertexID: vertex.ID,
			}, nil
		}
	}
	// error, result should not be nil
	// the only one reason to be nil is to vertex return parameter (channel) is not initialized
	return nil, nil
}

// serveInternal run calls callServeFn for each node and put the results in the map
func (e *Endure) serveInternal(n *DllNode) error {
	const op = errors.Op("internal_serve")
	// check if type implements serveInternal, if implements, call serveInternal
	if reflect.TypeOf(n.Vertex.Iface).Implements(reflect.TypeOf((*Service)(nil)).Elem()) {
		in := make([]reflect.Value, 0, 1)
		// add service itself
		in = append(in, reflect.ValueOf(n.Vertex.Iface))

		res, err := e.callServeFn(n.Vertex, in)
		if err != nil {
			return errors.E(op, errors.FunctionCall, err)
		}
		if res != nil {
			e.results.Store(res.vertexID, res)
		} else {
			e.logger.Error("nil result returned from the vertex", zap.String("vertex id", n.Vertex.ID), zap.String("tip:", "serveInternal function should return initialized channel with errors"))
			return errors.E(op, errors.FunctionCall, errors.Errorf("nil result returned from the vertex, vertex id: %s", n.Vertex.ID))
		}

		// start polling the vertex
		e.poll(res)
	}

	return nil
}
