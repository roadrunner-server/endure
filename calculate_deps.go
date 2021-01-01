package endure

import (
	"reflect"

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
func (e *Endure) implProvidesPath(vertexID string, vertex interface{}) error {
	const op = errors.Op("add_providers")
	provider := vertex.(Provider)
	for _, fn := range provider.Provides() {
		v := e.graph.GetVertex(vertexID)
		if v.Provides == nil {
			v.Provides = make(map[string]ProvidedEntry)
		}

		// Make a slice
		if v.Meta.FnsProviderToInvoke == nil {
			v.Meta.FnsProviderToInvoke = make([]ProviderEntry, 0, 1)
		}

		// TODO merge function calls into one. Plugin1 -> fn's to invoke ProvideDB, ProvideDB2
		// Append functions which we will invoke when we start calling the structure functions after Init stage
		pe := ProviderEntry{
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
				tmpValue := reflect.ValueOf(vertex)
				tmpIsRef := isReference(ret[i])
				v.Provides[typeStr] = ProvidedEntry{
					IsReference: &tmpIsRef,
					Value:       tmpValue,
				}
				continue
			}

			// just init map value
			v.Provides[typeStr] = ProvidedEntry{
				IsReference: nil,
			}
		}

		v.Meta.FnsProviderToInvoke = append(v.Meta.FnsProviderToInvoke, pe)
	}
	return nil
}

func (e *Endure) addCollectorsDeps(vertex *Vertex) error {
	// hot path
	if _, ok := vertex.Iface.(Collector); !ok {
		return nil
	}

	// vertex implements Collector interface
	return e.implCollectorPath(vertex)
}

func (e *Endure) walk(params []reflect.Type, v *Vertex) bool {
	onlyStructs := true
	for _, param := range params {
		if param.Kind() == reflect.Interface {
			onlyStructs = false
			if reflect.TypeOf(v.Iface).Implements(param) {
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

func (e *Endure) implCollectorPath(vertex *Vertex) error {
	// vertexID string, vertex interface{} same vertex
	const op = errors.Op("add_collectors_deps")
	collector := vertex.Iface.(Collector)
	// range Collectors functions
	for _, fn := range collector.Collects() {
		haveInterfaceDeps := false
		// what type it might depend on?
		params, err := paramsList(fn)
		if err != nil {
			return errors.E(op, err)
		}

		compatible := make(Vertices, 0, len(params))

		// check if we have Interface deps in the params
		// filter out interfaces, leave only structs
		for i := 0; i < len(e.graph.Vertices); i++ {
			// skip self
			if e.graph.Vertices[i].ID == vertex.ID {
				continue
			}
			if e.walk(params, e.graph.Vertices[i]) == true {
				compatible = append(compatible, e.graph.Vertices[i])
				// set, that we have interface deps
				haveInterfaceDeps = true
				e.logger.Info("vertex is compatible with Collects", zap.String("vertex id", e.graph.Vertices[i].ID), zap.String("collects from vertex", vertex.ID))
				continue
			}
			e.logger.Info("vertex is not compatible with Collects", zap.String("vertex id", e.graph.Vertices[i].ID), zap.String("collects from vertex", vertex.ID))
		}

		if len(compatible) == 0 && haveInterfaceDeps {
			e.logger.Info("no compatible vertices found", zap.String("collects from vertex", vertex.ID))
			return nil
		}
		// process mixed deps (interfaces + structs)
		if haveInterfaceDeps {
			return e.processInterfaceDeps(compatible, fn, vertex, params)
		}
		// process only struct deps if not interfaces were found
		return e.processStructDeps(fn, vertex, params)
	}
	return nil
}

func (e *Endure) processInterfaceDeps(compatible Vertices, fn interface{}, vertex *Vertex, params []reflect.Type) error {
	const op = errors.Op("process interface deps")
	for _, compat := range compatible {
		// add vertex itself
		cp := CollectorEntry{
			in: make([]In, 0, 0),
			fn: getFunctionName(fn),
		}
		cp.in = append(cp.in, In{
			in:  reflect.ValueOf(vertex.Iface),
			dep: vertex.ID,
		})

		for _, param := range params {
			// check if type is primitive type
			if isPrimitive(param.String()) {
				e.logger.Panic("primitive type in the function parameters", zap.String("vertex id", vertex.ID), zap.String("type", param.String()))
			}

			paramStr := param.String()
			if vertex.ID == paramStr {
				continue
			}

			switch param.Kind() {
			case reflect.Ptr:
				if param.Elem().Kind() == reflect.Struct {
					dep := e.graph.VerticesMap[(removePointerAsterisk(param.String()))]
					if dep == nil {
						return errors.E(op, errors.Errorf("can't find provider for the struct parameter: %s", removePointerAsterisk(param.String())))
					}

					cp.in = append(cp.in, In{
						in:  reflect.ValueOf(dep.Iface),
						dep: dep.ID,
					})

					err := e.graph.AddDep(vertex, removePointerAsterisk(dep.ID), Collects, isReference(param), param.Kind())
					if err != nil {
						return errors.E(op, err)
					}
				}
			case reflect.Interface:
				cp.in = append(cp.in, In{
					in:  reflect.ValueOf(compat.Iface),
					dep: compat.ID,
				})

				err := e.graph.AddDep(vertex, removePointerAsterisk(compat.ID), Collects, isReference(param), param.Kind())
				if err != nil {
					return errors.E(op, err)
				}

			case reflect.Struct:
				dep := e.graph.VerticesMap[(removePointerAsterisk(param.String()))]
				if dep == nil {
					return errors.E(op, errors.Errorf("can't find provider for the struct parameter: %s", removePointerAsterisk(param.String())))
				}

				cp.in = append(cp.in, In{
					in:  reflect.ValueOf(dep.Iface),
					dep: dep.ID,
				})

				err := e.graph.AddDep(vertex, removePointerAsterisk(dep.ID), Collects, isReference(param), param.Kind())
				if err != nil {
					return errors.E(op, err)
				}
			}
		}
		vertex.Meta.CollectorEntries = append(vertex.Meta.CollectorEntries, cp)
	}

	return nil
}

func (e *Endure) processStructDeps(fn interface{}, vertex *Vertex, params []reflect.Type) error {
	const op = errors.Op("process struct deps")
	// process only struct deps
	cp := CollectorEntry{
		in: make([]In, 0, 0),
		fn: getFunctionName(fn),
	}
	cp.in = append(cp.in, In{
		in:  reflect.ValueOf(vertex.Iface),
		dep: vertex.ID,
	})

	for _, param := range params {
		// check if type is primitive type
		if isPrimitive(param.String()) {
			e.logger.Panic("primitive type in the function parameters", zap.String("vertex id", vertex.ID), zap.String("type", param.String()))
		}

		// skip self
		paramStr := param.String()
		if vertex.ID == paramStr {
			continue
		}

		dep := e.graph.VerticesMap[(removePointerAsterisk(param.String()))]
		if dep == nil {
			depVertex := e.graph.FindProviders(removePointerAsterisk(paramStr))
			if depVertex == nil {
				e.logger.Warn("can't find any provider for the dependency, collector function on the vertex will not be invoked", zap.String("dep id", removePointerAsterisk(param.String())), zap.String("vertex id", vertex.ID))
				return nil
			}
			dep = depVertex
			for k, v := range dep.Provides {
				if k == removePointerAsterisk(paramStr) {
					cp.in = append(cp.in, In{
						in:  reflect.Zero(reflect.TypeOf(v)),
						dep: k,
					})
				}
			}
		} else {
			cp.in = append(cp.in, In{
				in:  reflect.ValueOf(dep.Iface),
				dep: dep.ID,
			})
		}

		tmpIsRef := isReference(param)
		tmpValue := reflect.ValueOf(dep.Iface)
		e.graph.AddGlobalProvider(removePointerAsterisk(paramStr), tmpValue)
		e.graph.VerticesMap[dep.ID].AddProvider(removePointerAsterisk(paramStr), tmpValue, tmpIsRef, param.Kind())

		err := e.graph.AddDep(vertex, removePointerAsterisk(paramStr), Collects, isReference(param), param.Kind())
		if err != nil {
			return errors.E(op, err)
		}

		e.logger.Debug("adding dependency via Collects()", zap.String("vertex id", vertex.ID), zap.String("depends", paramStr))
	}

	vertex.Meta.CollectorEntries = append(vertex.Meta.CollectorEntries, cp)
	return nil
}

// addEdges calculates simple graph for the dependencies
func (e *Endure) addEdges() error {
	const Op = errors.Op("add_edges")
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

func (e *Endure) addInitDeps(vertex *Vertex, initMethod reflect.Method) error {
	const Op = errors.Op("add_init_deps")
	// Init function in arguments
	initArgs := functionParameters(initMethod)

	// iterate over all function parameters
	for _, initArg := range initArgs {
		if isPrimitive(initArg.String()) {
			e.logger.Panic("primitive type in the function parameters", zap.String("vertex id", vertex.ID), zap.String("type", initArg.String()))
			continue
		}
		// receiver
		if vertex.ID == removePointerAsterisk(initArg.String()) {
			continue
		}
		if initArg.Kind() == reflect.Interface {
			for i := 0; i < len(e.graph.Vertices); i++ {
				// if type implements interface we should add this struct as provider of the interface
				if reflect.TypeOf(e.graph.Vertices[i].Iface).Implements(initArg) {
					// skip double add
					if _, ok := e.graph.Vertices[i].Provides[removePointerAsterisk(initArg.String())]; ok {
						continue
					}
					tmpIsRef := isReference(initArg)
					tmpValue := reflect.ValueOf(e.graph.Vertices[i].Iface)
					e.graph.AddGlobalProvider(removePointerAsterisk(initArg.String()), tmpValue)
					e.graph.Vertices[i].AddProvider(removePointerAsterisk(initArg.String()), tmpValue, tmpIsRef, initArg.Kind())
				}
			}
		}

		err := e.graph.AddDep(vertex, removePointerAsterisk(initArg.String()), Init, isReference(initArg), initArg.Kind())
		if err != nil {
			return errors.E(Op, err)
		}
		e.logger.Debug("adding dependency via Init()", zap.String("vertex id", vertex.ID), zap.String("depends on", initArg.String()))
	}
	return nil
}
