package cascade

import (
	"reflect"

	"github.com/rs/zerolog/log"
	"github.com/spiral/cascade/structures"
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

			gVertex.Provides[typeStr] = structures.ProvidedEntry{
				IsReference: nil,
				Value:       nil,
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
					log.Fatal().
						Str("vertexID", vertexID).
						Str("type", at.String()).
						Msg("primitive type in the function parameters")
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
				c.logger.Info().
					Str("vertexID", vertexID).
					Str("depends", atStr).
					Msg("adding dependency via Depends()")
			}

			// get the Vertex from the graph (gVertex)
			gVertex := c.graph.GetVertex(vertexID)
			if gVertex.Provides == nil {
				gVertex.Provides = make(map[string]structures.ProvidedEntry)
			}

			if gVertex.Meta.FnsRegisterToInvoke == nil {
				gVertex.Meta.FnsRegisterToInvoke = make([]string, 0, 5)
			}

			c.logger.Info().
				Str("vertexID", vertexID).
				Str("function name", getFunctionName(fn)).
				Msg("appending register function to invoke later")

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
			log.Fatal().
				Str("vertexID", vertexID).
				Str("type", initArg.String()).
				Msg("primitive type in the function parameters")
			continue
		}
		// receiver
		if vertexID == removePointerAsterisk(initArg.String()) {
			continue
		}


		err = c.graph.AddDep(vertexID, removePointerAsterisk(initArg.String()), structures.Init, isReference(initArg))
		if err != nil {
			return err
		}
		c.logger.Info().
			Str("vertexID", vertexID).
			Str("depends", initArg.String()).
			Msg("adding dependency via Init()")
	}
	return nil
}
