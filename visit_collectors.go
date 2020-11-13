package endure

import (
	"reflect"

	"github.com/spiral/errors"
	"go.uber.org/zap"
)

func (e *Endure) fnCallCollectors(vertex *Vertex, in []reflect.Value, methodName string) error {
	const op = errors.Op("internal_call_collector_functions")
	// type implements Collector interface
	if reflect.TypeOf(vertex.Iface).Implements(reflect.TypeOf((*Collector)(nil)).Elem()) {
		// if type implements Collector() it should has FnsProviderToInvoke
		m, ok := reflect.TypeOf(vertex.Iface).MethodByName(methodName)
		if !ok {
			e.logger.Error("type has missing method in FnsCollectorToInvoke", zap.String("vertex id", vertex.ID), zap.String("method", methodName))
			return errors.E(op, errors.FunctionCall, errors.Str("type has missing method in FnsCollectorToInvoke"))
		}

		ret := m.Func.Call(in)
		// handle error
		if len(ret) > 0 {
			// error is the last return parameter in line
			rErr := ret[len(ret)-1].Interface()
			if rErr != nil {
				if err, ok := rErr.(error); ok && e != nil {
					e.logger.Error("error calling CollectorFns", zap.String("vertex id", vertex.ID), zap.Error(err))
					return errors.E(op, errors.FunctionCall, err)
				}
				return errors.E(op, errors.FunctionCall, errors.Str("unknown error occurred during the function call"))
			}
		} else {
			return errors.E(op, errors.FunctionCall, errors.Str("collector should return Value and error types"))
		}
	}
	return nil
}
