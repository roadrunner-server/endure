package registar

import (
	"reflect"
)

type Registar struct {
	// id - plugin
	// values - types, which plugin have
	types map[reflect.Type]*registarEntry
}

func New() *Registar {
	return &Registar{
		types: make(map[reflect.Type]*registarEntry),
	}
}

func (r *Registar) Insert(plugin any, retType reflect.Type, method string) {
	key := reflect.TypeOf(plugin)
	if _, ok := r.types[key]; !ok {
		r.types[key] = &registarEntry{}
	}

	r.types[key].returnedTypes = append(r.types[key].returnedTypes, &returnedType{
		retType: retType,
		method:  method,
	})

	r.types[key].plugin = plugin
}

func (r *Registar) Update(plugin any, tp reflect.Type, value func() reflect.Value) {
	key := reflect.TypeOf(plugin)
	if _, ok := r.types[key]; !ok {
		return
	}

	// returned types
	types := r.types[key].returnedTypes

	for i := 0; i < len(types); i++ {
		if types[i].retType == tp {
			types[i].value = value
		}
	}
}

func (r *Registar) Value(plugin any, tp reflect.Type) (reflect.Value, bool) {
	key := reflect.TypeOf(plugin)
	if _, ok := r.types[key]; !ok {
		return reflect.Value{}, false
	}
	// returned types
	types := r.types[key].returnedTypes

	for i := 0; i < len(types); i++ {
		if types[i].retType == tp {
			return types[i].value(), true
		}
	}

	// initialized values for the particular type
	return reflect.Value{}, false
}

// TypeValue check that there are plugins (with Provides) that implement all types
func (r *Registar) TypeValue(plugin any, tp reflect.Type) (reflect.Value, bool) {
	key := reflect.TypeOf(plugin)
	if _, ok := r.types[key]; !ok {
		return reflect.Value{}, false
	}

	retTp := r.types[key]

	for i := 0; i < len(retTp.returnedTypes); i++ {
		if retTp.returnedTypes[i].retType.Implements(tp) {
			return retTp.returnedTypes[i].value(), true
		}
	}

	return reflect.Value{}, false
}

func (r *Registar) Remove(plugin any) {
	delete(r.types, reflect.TypeOf(plugin))
}

// Implements check that there are plugins (with Provides) that implement all types
func (r *Registar) Implements(types ...reflect.Type) []*implements {
	// matchingTypes are plugins (any)
	var matchingTypes []*implements

	// range over all registered types (basically all that we know about plugins and providers)
	for k, entry := range r.types {
		var methods []string
		implementsAllTypes := true
		// iterate over types, provided by the user
		// plugin (w or w/o provides) should implement this type
		for i := 0; i < len(types); i++ {
			requiredType := types[i]

			// our plugin might implement one of the needed types
			// if not, check if the plugin provides some types which might implement the type
			if !k.Implements(requiredType) {
				implemented := false
				// here we check that provides
				for j := 0; j < len(entry.returnedTypes); j++ {
					provided := entry.returnedTypes[j]
					if provided.retType.Implements(requiredType) {
						// plan -> call method
						implemented = true
						methods = append(methods, provided.method)
						break
					}
				}

				if !implemented {
					implementsAllTypes = false
					methods = nil
					break
				}
			}
		}

		if implementsAllTypes {
			matchingTypes = append(matchingTypes, &implements{
				plugin:  entry.Plugin(),
				methods: methods,
			})
		}
	}

	return matchingTypes
}
