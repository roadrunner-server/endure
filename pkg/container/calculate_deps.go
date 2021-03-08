package endure

import (
	"reflect"

	"github.com/spiral/endure/pkg/graph"
	"github.com/spiral/endure/pkg/vertex"
	"github.com/spiral/errors"
	"go.uber.org/zap"
)

/*
addProviders:
Adds a provided type via the Provider interface:
1. Key to the 'Vertex Provides' map with empty ProvidedEntry, because we use key at the Init stage and fill the map with
actual type after FnsProviderToInvoke will be invoked
2. FnsProviderToInvoke --> is the list of the functions to invoke via the reflection to get the actual provided type
*/
func (e *Endure) addProviders(vertexID string, vertex interface{}) error {
	// hot path
	if _, ok := vertex.(Provider); !ok {
		return nil
	}

	// vertex implements Provider interface
	return e.implProvidesPath(vertexID, vertex)
}

// vertex implements provider
func (e *Endure) implProvidesPath(vertexID string, vrtx interface{}) error {
	const op = errors.Op("endure_add_providers")
	provider := vrtx.(Provider)
	for _, fn := range provider.Provides() {
		v := e.graph.GetVertex(vertexID)

		// TODO merge function calls into one. Plugin1 -> fn's to invoke ProvideDB, ProvideDB2
		// Append functions which we will invoke when we start calling the structure functions after Init stage
		pe := vertex.ProviderEntry{
			/*
				For example:
				we need to invoke function ProvideDB - that will be FunctionName
				ReturnTypeId will be DB (in that case)
				We need return type to filter it in Init call, because in Init we may have one struct which returns
				two different types.
			*/

		}
		pe.FunctionName = getFunctionName(fn)

		// Check return types
		ret, err := providersReturnType(fn)
		if err != nil {
			return errors.E(op, err)
		}

		for i := 0; i < len(ret); i++ {
			// remove asterisk from the type string representation *Foo1 -> Foo1
			typeStr := removePointerAsterisk(ret[i].String())
			if typeStr == "error" {
				continue
			}
			// get the Vertex from the graph (v)

			pe.ReturnTypeIds = append(pe.ReturnTypeIds, typeStr)
			// function fn to invoke)
			/*
				   For the interface dependencies
					If Provided type is interface
					1. Check that type implement interface
					2. Write record, that this particular type also provides Interface dep
			*/
			if ret[i].Kind() == reflect.Interface {
				tmpValue := reflect.ValueOf(vrtx)
				tmpIsRef := isReference(ret[i])
				v.Provides[typeStr] = vertex.ProvidedEntry{
					IsReference: &tmpIsRef,
					Value:       tmpValue,
				}
				continue
			}

			// just init map value
			v.Provides[typeStr] = vertex.ProvidedEntry{
				IsReference: nil,
			}
		}

		v.Meta.FnsProviderToInvoke = append(v.Meta.FnsProviderToInvoke, pe)
	}
	return nil
}

func (e *Endure) addCollectorsDeps(vrtx *vertex.Vertex) error {
	// hot path
	if _, ok := vrtx.Iface.(Collector); !ok {
		return nil
	}

	// vertex implements Collector interface
	return e.implCollectorPath(vrtx)
}

func (e *Endure) walk(params []reflect.Type, v *vertex.Vertex) bool {
	onlyStructs := true
	for i := range params {
		if params[i].Kind() == reflect.Interface {
			onlyStructs = false
			if reflect.TypeOf(v.Iface).Implements(params[i]) {
				continue
			}
			return false
		}
		continue
	}

	if onlyStructs {
		return false
	}
	return true
}

/*
The logic here is the following:
As the first step, we add collector dependencies if a vertex implements Collector interface.
For every function in the Collector return args we check every vertex in the graph for satisfying to all interfaces (or
structs) in the function parameters list.

If we found such vertex we can go by the following paths:
1. Function parameters list contain interfaces
2. Function parameters list contain only structures (easy case)


*/
func (e *Endure) implCollectorPath(vrtx *vertex.Vertex) error {
	// vertexID string, vertex interface{} same vertex
	const op = errors.Op("endure_impl_collector_path")
	collector := vrtx.Iface.(Collector)
	// range Collectors functions
	for _, fn := range collector.Collects() {
		haveInterfaceDeps := false
		// what type it might depend on?
		params, err := fnIn(fn)
		if err != nil {
			return errors.E(op, err)
		}

		compatible := make([]*vertex.Vertex, 0, len(params))

		// check if we have Interface deps in the params
		// filter out interfaces, leave only structs
		for i := 0; i < len(e.graph.Vertices); i++ {
			// skip self
			if e.graph.Vertices[i].ID == vrtx.ID {
				continue
			}

			// false if params are structures
			if e.walk(params, e.graph.Vertices[i]) == true {
				compatible = append(compatible, e.graph.Vertices[i])
				// set, that we have interface deps
				haveInterfaceDeps = true
				e.logger.Info("vertex is compatible with Collects", zap.String("vertex id", e.graph.Vertices[i].ID), zap.String("collects from vertex", vrtx.ID))
				continue
			}
			e.logger.Info("vertex is not compatible with Collects", zap.String("vertex id", e.graph.Vertices[i].ID), zap.String("collects from vertex", vrtx.ID))
		}

		if len(compatible) == 0 && haveInterfaceDeps {
			e.logger.Info("no compatible vertices found", zap.String("collects from vertex", vrtx.ID))
			return nil
		}
		// process mixed deps (interfaces + structs)
		if haveInterfaceDeps {
			return e.processInterfaceDeps(compatible, getFunctionName(fn), vrtx, params)
		}
		// process only struct deps if not interfaces were found
		return e.processStructDeps(getFunctionName(fn), vrtx, params)
	}
	return nil
}

func (e *Endure) processInterfaceDeps(compatible []*vertex.Vertex, fnName string, vrtx *vertex.Vertex, params []reflect.Type) error {
	const op = errors.Op("endure_process_interface_deps")
	for i := 0; i < len(compatible); i++ {
		// add vertex itself
		cp := vertex.CollectorEntry{
			In: make([]vertex.In, 0, 0),
			Fn: fnName,
		}
		cp.In = append(cp.In, vertex.In{
			In:  reflect.ValueOf(vrtx.Iface),
			Dep: vrtx.ID,
		})

		for j := 0; j < len(params); j++ {
			// check if type is primitive type
			if isPrimitive(params[j].String()) {
				e.logger.Panic("primitive type in the function parameters", zap.String("vertex id", vrtx.ID), zap.String("type", params[j].String()))
			}

			paramStr := params[j].String()
			if vrtx.ID == paramStr {
				continue
			}

			switch params[j].Kind() {
			case reflect.Ptr:
				if params[j].Elem().Kind() == reflect.Struct {
					dep := e.graph.VerticesMap[(removePointerAsterisk(params[j].String()))]
					if dep == nil {
						return errors.E(op, errors.Errorf("can't find provider for the struct parameter: %s", removePointerAsterisk(params[j].String())))
					}

					cp.In = append(cp.In, vertex.In{
						In:  reflect.ValueOf(dep.Iface),
						Dep: dep.ID,
					})

					err := e.graph.AddStructureDep(vrtx, removePointerAsterisk(dep.ID), graph.Collects, isReference(params[j]))
					if err != nil {
						return errors.E(op, err)
					}
				}
			case reflect.Interface:
				cp.In = append(cp.In, vertex.In{
					In:  reflect.ValueOf(compatible[i].Iface),
					Dep: compatible[i].ID,
				})

				err := e.graph.AddInterfaceDep(vrtx, removePointerAsterisk(compatible[i].ID), graph.Collects, isReference(params[j]))
				if err != nil {
					return errors.E(op, err)
				}

			case reflect.Struct:
				dep := e.graph.VerticesMap[(removePointerAsterisk(params[j].String()))]
				if dep == nil {
					return errors.E(op, errors.Errorf("can't find provider for the struct parameter: %s", removePointerAsterisk(params[j].String())))
				}

				cp.In = append(cp.In, vertex.In{
					In:  reflect.ValueOf(dep.Iface),
					Dep: dep.ID,
				})

				err := e.graph.AddStructureDep(vrtx, removePointerAsterisk(dep.ID), graph.Collects, isReference(params[j]))
				if err != nil {
					return errors.E(op, err)
				}
			}
		}
		vrtx.Meta.CollectorEntries = append(vrtx.Meta.CollectorEntries, cp)
	}

	return nil
}

func (e *Endure) processStructDeps(fnName string, vrtx *vertex.Vertex, params []reflect.Type) error {
	const op = errors.Op("endure_process_struct_deps")
	// process only struct deps
	cp := vertex.CollectorEntry{
		In: make([]vertex.In, 0, 0),
		Fn: fnName,
	}
	cp.In = append(cp.In, vertex.In{
		In:  reflect.ValueOf(vrtx.Iface),
		Dep: vrtx.ID,
	})

	for _, param := range params {
		// check if type is primitive type
		if isPrimitive(param.String()) {
			e.logger.Panic("primitive type in the function parameters", zap.String("vertex id", vrtx.ID), zap.String("type", param.String()))
		}

		// skip self
		paramStr := param.String()
		if vrtx.ID == paramStr {
			continue
		}

		dep := e.graph.VerticesMap[(removePointerAsterisk(param.String()))]
		if dep == nil {
			depVertex := e.graph.FindProviders(removePointerAsterisk(paramStr))
			if depVertex == nil {
				e.logger.Warn("can't find any provider for the dependency, collector function on the vertex will not be invoked", zap.String("dep id", removePointerAsterisk(param.String())), zap.String("vertex id", vrtx.ID))
				return nil
			}
			dep = depVertex
			for k, v := range dep.Provides {
				if k == removePointerAsterisk(paramStr) {
					cp.In = append(cp.In, vertex.In{
						In:  reflect.Zero(reflect.TypeOf(v)),
						Dep: k,
					})
				}
			}
		} else {
			cp.In = append(cp.In, vertex.In{
				In:  reflect.ValueOf(dep.Iface),
				Dep: dep.ID,
			})
		}

		tmpIsRef := isReference(param)
		tmpValue := reflect.ValueOf(dep.Iface)
		e.graph.AddGlobalProvider(removePointerAsterisk(paramStr), tmpValue)
		e.graph.VerticesMap[dep.ID].AddProvider(removePointerAsterisk(paramStr), tmpValue, tmpIsRef, param.Kind())

		err := e.graph.AddStructureDep(vrtx, removePointerAsterisk(paramStr), graph.Collects, isReference(param))
		if err != nil {
			return errors.E(op, err)
		}

		e.logger.Debug("adding dependency via Collects()", zap.String("vertex id", vrtx.ID), zap.String("depends", paramStr))
	}

	vrtx.Meta.CollectorEntries = append(vrtx.Meta.CollectorEntries, cp)
	return nil
}

// addEdges calculates simple graph for the dependencies
func (e *Endure) addEdges() error {
	const Op = errors.Op("endure_add_edges")
	// vertexID for example S2
	for vertexID, vrtx := range e.graph.VerticesMap {
		// we already checked the interface satisfaction
		// and we can safely skip the OK parameter here
		initMethod, _ := reflect.TypeOf(vrtx.Iface).MethodByName(InitMethodName)

		if initMethod.Type == nil {
			e.logger.Fatal("Init method is absent in struct", zap.String("vertex id", vertexID))
			return errors.E(Op, errors.Errorf("init method is absent in struct"))
		}

		/* Add the dependencies (if) which this vertex needs to internal_init
		Information we know at this step is:
		1. vertexID
		2. Vertex structure value (interface)
		3. Provided type
		4. Provided type String fn
		5. FunctionName of the dependencies which we should found
		We add 3 and 4 points to the Vertex
		*/
		err := e.addCollectorsDeps(vrtx)
		if err != nil {
			return errors.E(Op, err)
		}

		/*
			At this step we know (and build) all dependencies via the Collects interface and connected all providers
			to it's dependencies.
			The next step is to calculate dependencies provided by the Init() method
			for example S1.Init(foo2.DB) S1 --> foo2.S2 (not foo2.DB, because vertex which provides foo2.DB is foo2.S2)
		*/
		err = e.addInitDeps(vrtx, initMethod)
		if err != nil {
			return errors.E(Op, err)
		}
	}

	return nil
}

func (e *Endure) addInitDeps(vrtx *vertex.Vertex, initMethod reflect.Method) error {
	const Op = errors.Op("endure_add_init_deps")
	// Init function in arguments
	initArgs := functionParameters(initMethod)

	// iterate over all function parameters
	for _, initArg := range initArgs {
		if isPrimitive(initArg.String()) {
			e.logger.Panic("primitive type in the function parameters", zap.String("vertex id", vrtx.ID), zap.String("type", initArg.String()))
		}
		initArgStr := removePointerAsterisk(initArg.String())
		// receiver
		if vrtx.ID == initArgStr {
			continue
		}

		// if init arg disabled, remove_vertex the whole vertex
		if _, ok := e.disabled[initArgStr]; ok {
			e.logger.Info("vertex receives disabled init vertex", zap.String("vertex id", vrtx.ID), zap.String("disabled init arg", initArgStr))
			e.disabled[vrtx.ID] = true
			continue
		}

		if initArg.Kind() == reflect.Interface {
			for i := 0; i < len(e.graph.Vertices); i++ {
				// if type implements interface we should add this struct as provider of the interface
				if reflect.TypeOf(e.graph.Vertices[i].Iface).Implements(initArg) {
					// skip double add
					if _, ok := e.graph.Vertices[i].Provides[initArgStr]; ok {
						continue
					}
					tmpIsRef := isReference(initArg)
					tmpValue := reflect.ValueOf(e.graph.Vertices[i].Iface)
					e.graph.AddGlobalProvider(initArgStr, tmpValue)
					e.graph.Vertices[i].AddProvider(initArgStr, tmpValue, tmpIsRef, initArg.Kind())
				}
			}

			err := e.graph.AddInterfaceDep(vrtx, initArgStr, graph.Init, isReference(initArg))
			if err != nil {
				return errors.E(Op, err)
			}
			e.logger.Debug("adding dependency via Init()", zap.String("vertex id", vrtx.ID), zap.String("depends on", initArg.String()))
			continue
		}
		err := e.graph.AddStructureDep(vrtx, initArgStr, graph.Init, isReference(initArg))
		if err != nil {
			return errors.E(Op, err)
		}
		e.logger.Debug("adding dependency via Init()", zap.String("vertex id", vrtx.ID), zap.String("depends on", initArg.String()))
	}
	return nil
}
