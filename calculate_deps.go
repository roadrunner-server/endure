package cascade

import (
	"reflect"

	"github.com/spiral/cascade/structures"
	"go.uber.org/zap"
)

/*
addProviders:
Adds a provided type via the Provider interface. And adding:
1. Key to the `Vertex Provides` map with empty ProvidedEntry, because we use key at the Init stage and fill the map with
actual type after FnsProviderToInvoke will be invoked
2. FnsProviderToInvoke --> is the list of the Provided function to invoke via the reflection
*/
func (c *Cascade) addProviders(vertexID string, vertex interface{}) error {
	if provider, ok := vertex.(Provider); ok {
		for _, fn := range provider.Provides() {
			ret, err := providersReturnType(fn)
			if err != nil {
				// todo: delete gVertex
				return err
			}

			typeStr := removePointerAsterisk(ret.String())
			// get the Vertex from the graph (gVertex)
			gVertex := c.graph.GetVertex(vertexID)
			if gVertex.Provides == nil {
				gVertex.Provides = make(map[string]structures.ProvidedEntry)
			}

			if gVertex.Meta.FnsProviderToInvoke == nil {
				gVertex.Meta.FnsProviderToInvoke = make([]string, 0, 5)
			}

			gVertex.Meta.FnsProviderToInvoke = append(gVertex.Meta.FnsProviderToInvoke, getFunctionName(fn))

			if ret.Kind() == reflect.Interface {
				if reflect.TypeOf(vertex).Implements(ret) {
					tmpValue := reflect.ValueOf(vertex)
					tmpIsRef := isReference(ret)
					gVertex.Provides[typeStr] = structures.ProvidedEntry{
						IsReference: &tmpIsRef,
						Value:       &tmpValue,
					}
				} else {
					gVertex.Provides[typeStr] = structures.ProvidedEntry{
						IsReference: nil,
						Value:       nil,
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
func (c *Cascade) addEdges() error {
	// vertexID for example S2
	for vertexID, vrtx := range c.graph.Graph {
		// we already checked the interface satisfaction
		// and we can safely skip the OK parameter here
		init, _ := reflect.TypeOf(vrtx.Iface).MethodByName(InitMethodName)

		/* Add the dependencies (if) which this vertex needs to init
		Information we know at this step is:
		1. VertexId
		2. Vertex structure value (interface)
		3. Provided type
		4. Provided type String name
		5. Name of the dependencies which we should found
		We add 3 and 4 points to the Vertex
		*/
		err := c.addRegisterDeps(vertexID, vrtx.Iface)
		if err != nil {
			return err
		}

		/*
			At this step we know (and build) all dependencies via the Depends interface and connected all providers
			to it's dependencies.
			The next step is to calculate dependencies provided by the Init() method
			for example S1.Init(foo2.DB) S1 --> foo2.S2 (not foo2.DB, because vertex which provides foo2.DB is foo2.S2)
		*/
		err = c.addInitDeps(vertexID, init)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Cascade) addRegisterDeps(vertexID string, vertex interface{}) error {
	if register, ok := vertex.(Register); ok {
		for _, fn := range register.Depends() {
			// what type it might depend on?
			argsTypes, err := argType(fn)
			if err != nil {
				// todo: delete vertex
				return err
			}

			// at is like foo2.S2
			// we already checked argsTypes len
			for _, at := range argsTypes {
				// check if type is primitive type
				if isPrimitive(at.String()) {
					c.logger.Panic("primitive type in the function parameters", zap.String("vertex id", vertexID), zap.String("type", at.String()))
				}
				atStr := at.String()
				if vertexID == atStr {
					continue
				}
				// if we found, that some structure depends on some type
				// we also save it in the `depends` section
				// name s1 (for example)
				// vertex - S4 func

				// we store pointer in the Deps structure in the isRef field
				err = c.graph.AddDep(vertexID, removePointerAsterisk(atStr), structures.Depends, isReference(at))
				if err != nil {
					return err
				}
				c.logger.Debug("adding dependency via Depends()", zap.String("vertex id", vertexID), zap.String("depends", atStr))
			}

			// get the Vertex from the graph (gVertex)
			gVertex := c.graph.GetVertex(vertexID)
			if gVertex.Provides == nil {
				gVertex.Provides = make(map[string]structures.ProvidedEntry)
			}

			if gVertex.Meta.FnsRegisterToInvoke == nil {
				gVertex.Meta.FnsRegisterToInvoke = make([]string, 0, 5)
			}

			c.logger.Debug("appending register function to invoke later", zap.String("vertex id", vertexID), zap.String("function name", getFunctionName(fn)))

			gVertex.Meta.FnsRegisterToInvoke = append(gVertex.Meta.FnsRegisterToInvoke, getFunctionName(fn))
		}
	}

	return nil
}

func (c *Cascade) addInitDeps(vertexID string, initMethod reflect.Method) error {
	// S2 init args
	initArgs, err := functionParameters(initMethod)
	if err != nil {
		return err
	}

	// iterate over all function parameters
	for _, initArg := range initArgs {
		if isPrimitive(initArg.String()) {
			c.logger.Panic("primitive type in the function parameters", zap.String("vertex id", vertexID), zap.String("type", initArg.String()))
			continue
		}
		// receiver
		if vertexID == removePointerAsterisk(initArg.String()) {
			continue
		}
		if initArg.Kind() == reflect.Interface {
			for i := 0; i < len(c.graph.Vertices); i++ {
				// if type implements interface we should add this struct as provider of the interface
				if reflect.TypeOf(c.graph.Vertices[i].Iface).Implements(initArg) {
					// skip double add
					if _, ok := c.graph.Vertices[i].Provides[removePointerAsterisk(initArg.String())]; ok {
						continue
					}
					tmpIsRef := isReference(initArg)
					tmpValue := reflect.ValueOf(c.graph.Vertices[i].Iface)
					if c.graph.Vertices[i].Provides != nil {
						c.graph.Vertices[i].Provides[removePointerAsterisk(initArg.String())] = structures.ProvidedEntry{
							IsReference: &tmpIsRef,
							Value:       &tmpValue,
						}
					} else {
						c.graph.Vertices[i].Provides = make(map[string]structures.ProvidedEntry)
						c.graph.Vertices[i].Provides[removePointerAsterisk(initArg.String())] = structures.ProvidedEntry{
							IsReference: &tmpIsRef,
							Value:       &tmpValue,
						}
					}
				}
			}
		}

		err = c.graph.AddDep(vertexID, removePointerAsterisk(initArg.String()), structures.Init, isReference(initArg))
		if err != nil {
			return err
		}
		c.logger.Debug("adding dependency via Init()", zap.String("vertex id", vertexID), zap.String("depends", initArg.String()))
	}
	return nil
}
