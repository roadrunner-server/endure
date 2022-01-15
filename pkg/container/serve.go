package endure

import (
	"reflect"

	ll "github.com/roadrunner-server/endure/pkg/linked_list"
	"github.com/roadrunner-server/endure/pkg/vertex"
	"github.com/spiral/errors"
	"go.uber.org/zap"
)

func (e *Endure) callServeFn(vrtx *vertex.Vertex, in []reflect.Value) (*result, error) {
	const op = errors.Op("endure_call_serve_fn")
	e.logger.Debug("preparing to call Serve on the Vertex", zap.String("id", vrtx.ID))
	// find Serve method
	m, _ := reflect.TypeOf(vrtx.Iface).MethodByName(ServeMethodName)
	// call with needed number of `in` parameters
	ret := m.Func.Call(in)
	res := ret[0].Interface()
	e.logger.Debug("called Serve on the vertex", zap.String("id", vrtx.ID))
	if res != nil {
		if err, ok := res.(chan error); ok && err != nil {
			// error come right after we start serving the vrtx
			if len(err) > 0 {
				// read the error
				err := <-err
				return nil, errors.E(op, errors.FunctionCall, errors.Errorf("got initial serve error from the Vertex %s, stopping execution, error: %v", vrtx.ID, err))
			}
			return &result{
				errCh:    err,
				signal:   make(chan notify),
				vertexID: vrtx.ID,
			}, nil
		}
	}
	// error, result should not be nil
	// the only one reason to be nil is to vrtx return parameter (channel) is not initialized
	return nil, nil
}

// serveInternal run calls callServeFn for each node and put the results in the map
func (e *Endure) serveInternal(n *ll.DllNode) error {
	const op = errors.Op("endure_serve_internal")
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
			e.logger.Error("nil result returned from the vertex", zap.String("id", n.Vertex.ID), zap.String("tip:", "serveInternal function should return initialized channel with errors"))
			return errors.E(op, errors.FunctionCall, errors.Errorf("nil result returned from the vertex, id: %s", n.Vertex.ID))
		}

		// start polling the vertex
		e.poll(res)
	}

	return nil
}
