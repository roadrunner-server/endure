package endure

import (
	"fmt"
	"reflect"

	"github.com/spiral/endure/structures"
	"github.com/spiral/errors"
	"go.uber.org/zap"
)

/*
addProviders:
Adds a provided type via the Provider interface. And adding:
1. Key to the 'Vertex Provides' map with empty ProvidedEntry, because we use key at the Init stage and fill the map with
actual type after FnsProviderToInvoke will be invoked
2. FnsProviderToInvoke --> is the list of the Provided function to invoke via the reflection
*/
func (e *Endure) addProviders(vertexID string, vertex interface{}) error {
	op := errors.Op("add_providers")
	// if vertex provides some deps via Provides, calculate it
	if provider, ok := vertex.(Provider); ok {
		for _, fn := range provider.Provides() {
			// Check return types
			ret, err := providersReturnType(fn)
			if err != nil {
				return errors.E(op, err)
			}

			// remove asterisk from the type string representation *Foo1 -> Foo1
			typeStr := removePointerAsterisk(ret.String())
			// get the Vertex from the graph (gVertex)
			gVertex := e.graph.GetVertex(vertexID)
			if gVertex.Provides == nil {
				gVertex.Provides = make(map[string]structures.ProvidedEntry)
			}

			// Make a slice
			if gVertex.Meta.FnsProviderToInvoke == nil {
				gVertex.Meta.FnsProviderToInvoke = make([]structures.ProviderEntry, 0, 1)
			}

			// TODO merge function calls into one. Plugin1 -> fn's to invoke ProvideDB, ProvideDB2
			// Append functions which we will invoke when we start calling the structure functions after Init stage
			gVertex.Meta.FnsProviderToInvoke = append(gVertex.Meta.FnsProviderToInvoke, structures.ProviderEntry{
				/*
					For example:
					we need to invoke function ProvideDB - that will be FunctionName
					ReturnTypeId will be DB (in that case)
					We need return type to filter it in Init call, because in Init we may have one struct which returns
					two different types.
				*/
				FunctionName: getFunctionName(fn), // function name to invoke
				ReturnTypeId: typeStr,             // return type ID
			})

			/*
				   For the interface dependencies
					If Provided type is interface
					1. Check that type implement interface
					2. Write record, that this particular type also provides Interface
			*/
			if ret.Kind() == reflect.Interface {
				if reflect.TypeOf(vertex).Implements(ret) {
					tmpValue := reflect.ValueOf(vertex)
					tmpIsRef := isReference(ret)
					gVertex.Provides[typeStr] = structures.ProvidedEntry{
						IsReference: &tmpIsRef,
						Value:       &tmpValue,
					}
				}
			} else {
				gVertex.Provides[typeStr] = structures.ProvidedEntry{
					IsReference: nil,
					Value:       nil,
				}
			}
		}
	}
	return nil
}

// addEdges calculates simple graph for the dependencies
func (e *Endure) addEdges() error {
	const Op = "add_edges"
	// vertexID for example S2
	for vertexID, vrtx := range e.graph.VerticesMap {
		// we already checked the interface satisfaction
		// and we can safely skip the OK parameter here
		init, _ := reflect.TypeOf(vrtx.Iface).MethodByName(InitMethodName)

		if init.Type == nil {
			e.logger.Fatal("init method is absent in struct", zap.String("vertex id", vertexID))
			return errors.E(Op, fmt.Errorf("init method is absent in struct"))
		}

		/* Add the dependencies (if) which this vertex needs to init
		Information we know at this step is:
		1. vertexID
		2. Vertex structure value (interface)
		3. Provided type
		4. Provided type String name
		5. FunctionName of the dependencies which we should found
		We add 3 and 4 points to the Vertex
		*/
		err := e.addCollectorsDeps(vertexID, vrtx.Iface)
		if err != nil {
			return errors.E(Op, err)
		}

		/*
			At this step we know (and build) all dependencies via the Collects interface and connected all providers
			to it's dependencies.
			The next step is to calculate dependencies provided by the Init() method
			for example S1.Init(foo2.DB) S1 --> foo2.S2 (not foo2.DB, because vertex which provides foo2.DB is foo2.S2)
		*/
		err = e.addInitDeps(vertexID, init)
		if err != nil {
			return errors.E(Op, err)
		}
	}

	return nil
}

func (e *Endure) addCollectorsDeps(vertexID string, vertex interface{}) error {
	const Op = "add_collectors_deps"
	if register, ok := vertex.(Collector); ok {
		for _, fn := range register.Collects() {
			// what type it might depend on?
			argsTypes, err := argType(fn)
			if err != nil {
				return errors.E(Op, err)
			}

			// at is like foo2.S2
			// we already checked argsTypes len
			for _, at := range argsTypes {
				// check if type is primitive type
				if isPrimitive(at.String()) {
					e.logger.Panic("primitive type in the function parameters", zap.String("vertex id", vertexID), zap.String("type", at.String()))
				}
				atStr := at.String()
				if vertexID == atStr {
					continue
				}
				// depends at interface via Collectors
				/*
					In this case we should do the following:
					1. Find all types, which implement this interface
					2. Make this type depend from all these types
					3.
				*/
				if at.Kind() == reflect.Interface {
					// go over all dependencies
					for i := 0; i < len(e.graph.Vertices); i++ {
						if reflect.TypeOf(e.graph.Vertices[i].Iface).Implements(at) {
							tmpIsRef := isReference(at)
							tmpValue := reflect.ValueOf(e.graph.Vertices[i].Iface)
							e.graph.Vertices[i].AddProvider(removePointerAsterisk(atStr), tmpValue, tmpIsRef, at.Kind())
						}
					}
				}
				// if we found, that some structure depends on some type
				// we also save it in the `depends` section
				// name s1 (for example)
				// vertex - S4 func

				// we store pointer in the Deps structure in the isRef field
				err = e.graph.AddDep(vertexID, removePointerAsterisk(atStr), structures.Collects, isReference(at), at.Kind())
				if err != nil {
					return errors.E(Op, err)
				}
				e.logger.Debug("adding dependency via Collects()", zap.String("vertex id", vertexID), zap.String("depends", atStr))
			}

			// get the Vertex from the graph (gVertex)
			gVertex := e.graph.GetVertex(vertexID)
			if gVertex.Provides == nil {
				gVertex.Provides = make(map[string]structures.ProvidedEntry)
			}

			if gVertex.Meta.FnsCollectorToInvoke == nil {
				gVertex.Meta.FnsCollectorToInvoke = make([]string, 0, 5)
			}

			e.logger.Debug("appending collector function to invoke later", zap.String("vertex id", vertexID), zap.String("function name", getFunctionName(fn)))

			gVertex.Meta.FnsCollectorToInvoke = append(gVertex.Meta.FnsCollectorToInvoke, getFunctionName(fn))
		}
	}

	return nil
}

func (e *Endure) addInitDeps(vertexID string, initMethod reflect.Method) error {
	const Op = "add_init_deps"
	// Init function in arguments
	initArgs := functionParameters(initMethod)

	// iterate over all function parameters
	for _, initArg := range initArgs {
		if isPrimitive(initArg.String()) {
			e.logger.Panic("primitive type in the function parameters", zap.String("vertex id", vertexID), zap.String("type", initArg.String()))
			continue
		}
		// receiver
		if vertexID == removePointerAsterisk(initArg.String()) {
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
					e.graph.Vertices[i].AddProvider(removePointerAsterisk(initArg.String()), tmpValue, tmpIsRef, initArg.Kind())
				}
			}
		}

		err := e.graph.AddDep(vertexID, removePointerAsterisk(initArg.String()), structures.Init, isReference(initArg), initArg.Kind())
		if err != nil {
			return errors.E(Op, err)
		}
		e.logger.Debug("adding dependency via Init()", zap.String("vertex id", vertexID), zap.String("depends", initArg.String()))
	}
	return nil
}
