package endure

import (
	"fmt"
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
		// Check return types
		ret, err := providersReturnType(fn)
		if err != nil {
			return errors.E(op, err)
		}

		// remove asterisk from the type string representation *Foo1 -> Foo1
		typeStr := removePointerAsterisk(ret.String())
		// get the Vertex from the graph (v)
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
		v.Meta.FnsProviderToInvoke = append(v.Meta.FnsProviderToInvoke, ProviderEntry{
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
				2. Write record, that this particular type also provides Interface dep
		*/
		if ret.Kind() == reflect.Interface {
			tmpValue := reflect.ValueOf(vertex)
			tmpIsRef := isReference(ret)
			v.Provides[typeStr] = ProvidedEntry{
				IsReference: &tmpIsRef,
				Value:       &tmpValue,
			}
			return nil
		}

		// just init map value
		v.Provides[typeStr] = ProvidedEntry{
			IsReference: nil,
			Value:       nil,
		}
	}
	return nil
}

func (e *Endure) addCollectorsDeps(vertexID string, vertex interface{}) error {
	// hot path
	if _, ok := vertex.(Collector); !ok {
		return nil
	}

	// vertex implements Collector interface
	return e.implCollectorPath(vertexID, vertex)
}

func (e *Endure) implCollectorPath(vertexID string, vertex interface{}) error {
	const op = errors.Op("add_collectors_deps")
	collector := vertex.(Collector)
	for _, fn := range collector.Collects() {
		// what type it might depend on?
		argsTypes, err := argType(fn)
		if err != nil {
			return errors.E(op, err)
		}

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
			err = e.graph.AddDep(vertexID, removePointerAsterisk(atStr), Collects, isReference(at), at.Kind())
			if err != nil {
				return errors.E(op, err)
			}
			e.logger.Debug("adding dependency via Collects()", zap.String("vertex id", vertexID), zap.String("depends", atStr))
		}

		// get the Vertex from the graph (v)
		v := e.graph.GetVertex(vertexID)
		if v.Provides == nil {
			v.Provides = make(map[string]ProvidedEntry)
		}

		if v.Meta.FnsCollectorToInvoke == nil {
			v.Meta.FnsCollectorToInvoke = make([]string, 0, 5)
		}

		e.logger.Debug("appending collector function to invoke later", zap.String("vertex id", vertexID), zap.String("function name", getFunctionName(fn)))

		v.Meta.FnsCollectorToInvoke = append(v.Meta.FnsCollectorToInvoke, getFunctionName(fn))
	}
	return nil
}

// addEdges calculates simple graph for the dependencies
func (e *Endure) addEdges() error {
	const Op = errors.Op("add_edges")
	// vertexID for example S2
	for vertexID, vrtx := range e.graph.VerticesMap {
		// we already checked the interface satisfaction
		// and we can safely skip the OK parameter here
		init, _ := reflect.TypeOf(vrtx.Iface).MethodByName(InitMethodName)

		if init.Type == nil {
			e.logger.Fatal("internal_init method is absent in struct", zap.String("vertex id", vertexID))
			return errors.E(Op, fmt.Errorf("internal_init method is absent in struct"))
		}

		/* Add the dependencies (if) which this vertex needs to internal_init
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

func (e *Endure) addInitDeps(vertexID string, initMethod reflect.Method) error {
	const Op = errors.Op("add_init_deps")
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

		err := e.graph.AddDep(vertexID, removePointerAsterisk(initArg.String()), Init, isReference(initArg), initArg.Kind())
		if err != nil {
			return errors.E(Op, err)
		}
		e.logger.Debug("adding dependency via Init()", zap.String("vertex id", vertexID), zap.String("depends on", initArg.String()))
	}
	return nil
}
