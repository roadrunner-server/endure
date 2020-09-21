package endure

import (
	"reflect"

	"github.com/spiral/endure/structures"
	"go.uber.org/zap"
)

/*
addProviders:
Adds a provided type via the Provider interface. And adding:
1. Key to the `Vertex Provides` map with empty ProvidedEntry, because we use key at the Init stage and fill the map with
actual type after FnsProviderToInvoke will be invoked
2. FnsProviderToInvoke --> is the list of the Provided function to invoke via the reflection
*/
func (e *Endure) addProviders(vertexID string, vertex interface{}) error {
	if provider, ok := vertex.(Provider); ok {
		for _, fn := range provider.Provides() {
			ret, err := dependersReturnType(fn)
			if err != nil {
				return err
			}

			typeStr := removePointerAsterisk(ret.String())
			// get the Vertex from the graph (gVertex)
			gVertex := e.graph.GetVertex(vertexID)
			if gVertex.Provides == nil {
				gVertex.Provides = make(map[string]structures.ProvidedEntry)
			}

			if gVertex.Meta.FnsProviderToInvoke == nil {
				gVertex.Meta.FnsProviderToInvoke = make([]string, 0, 5)
			}

			gVertex.Meta.FnsProviderToInvoke = append(gVertex.Meta.FnsProviderToInvoke, getFunctionName(fn))

			// Interface dep
			/*
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
	// vertexID for example S2
	for vertexID, vrtx := range e.graph.VerticesMap {
		// we already checked the interface satisfaction
		// and we can safely skip the OK parameter here
		init, _ := reflect.TypeOf(vrtx.Iface).MethodByName(InitMethodName)

		/* Add the dependencies (if) which this vertex needs to init
		Information we know at this step is:
		1. vertexID
		2. Vertex structure value (interface)
		3. Provided type
		4. Provided type String name
		5. Name of the dependencies which we should found
		We add 3 and 4 points to the Vertex
		*/
		err := e.addDependersDeps(vertexID, vrtx.Iface)
		if err != nil {
			return err
		}

		/*
			At this step we know (and build) all dependencies via the Depends interface and connected all providers
			to it's dependencies.
			The next step is to calculate dependencies provided by the Init() method
			for example S1.Init(foo2.DB) S1 --> foo2.S2 (not foo2.DB, because vertex which provides foo2.DB is foo2.S2)
		*/
		err = e.addInitDeps(vertexID, init)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *Endure) addDependersDeps(vertexID string, vertex interface{}) error {
	if register, ok := vertex.(Depender); ok {
		for _, fn := range register.Depends() {
			// what type it might depend on?
			argsTypes, err := argType(fn)
			if err != nil {
				return err
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
				// depends at interface via Dependers
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
				err = e.graph.AddDep(vertexID, removePointerAsterisk(atStr), structures.Depends, isReference(at), at.Kind())
				if err != nil {
					return err
				}
				e.logger.Debug("adding dependency via Depends()", zap.String("vertex id", vertexID), zap.String("depends", atStr))
			}

			// get the Vertex from the graph (gVertex)
			gVertex := e.graph.GetVertex(vertexID)
			if gVertex.Provides == nil {
				gVertex.Provides = make(map[string]structures.ProvidedEntry)
			}

			if gVertex.Meta.FnsDependerToInvoke == nil {
				gVertex.Meta.FnsDependerToInvoke = make([]string, 0, 5)
			}

			e.logger.Debug("appending depender function to invoke later", zap.String("vertex id", vertexID), zap.String("function name", getFunctionName(fn)))

			gVertex.Meta.FnsDependerToInvoke = append(gVertex.Meta.FnsDependerToInvoke, getFunctionName(fn))
		}
	}

	return nil
}

func (e *Endure) addInitDeps(vertexID string, initMethod reflect.Method) error {
	// S2 init args
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
			return err
		}
		e.logger.Debug("adding dependency via Init()", zap.String("vertex id", vertexID), zap.String("depends", initArg.String()))
	}
	return nil
}
