package registar

import (
	"reflect"
	"sort"
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

func (r *Registar) Insert(plugin any, retType reflect.Type, method string, weight uint) {
	key := reflect.TypeOf(plugin)
	if _, ok := r.types[key]; !ok {
		r.types[key] = &registarEntry{}
	}

	r.types[key].returnedTypes = append(r.types[key].returnedTypes, &returnedType{
		retType: retType,
		method:  method,
	})

	r.types[key].weight = weight
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
			if retTp.returnedTypes[i].value == nil {
				return reflect.Value{}, false
			}
			return retTp.returnedTypes[i].value(), true
		}
	}

	return reflect.Value{}, false
}

func (r *Registar) Remove(plugin any) {
	delete(r.types, reflect.TypeOf(plugin))
}

// Implements check that there are plugins (with Provides) that implement all types
func (r *Registar) Implements(tp reflect.Type) []*implements {
	var impl []*implements
	// range over all registered types (basically all that we know about plugins and providers)
	for k, entry := range r.types {
		// iterate over types, provided by the user
		// plugin (w or w/o provides) should implement this type

		// our plugin might implement one of the needed types
		// if not, check if the plugin provides some types which might implement the type
		if k.Implements(tp) {
			impl = append(impl, &implements{
				plugin: entry.Plugin(),
				weight: entry.Weight(),
			})
			continue
		}

		// here we check that provides
		for j := 0; j < len(entry.returnedTypes); j++ {
			provided := entry.returnedTypes[j]
			if provided.retType.Implements(tp) {
				impl = append(impl,
					&implements{
						plugin:  entry.Plugin(),
						weight:  entry.Weight(),
						methods: provided.method,
					},
				)
			}
		}
	}

	// sort by weight
	sort.Slice(impl, func(i, j int) bool {
		return impl[i].weight > impl[j].weight
	})

	return impl
}

func (r *Registar) ImplementsExcept(tp reflect.Type, plugin any) []*implements {
	var impl []*implements
	excl := reflect.TypeOf(plugin)
	// range over all registered types (basically all that we know about plugins and providers)
	for k, entry := range r.types {
		if k == excl {
			continue
		}
		// iterate over types, provided by the user
		// plugin (w or w/o provides) should implement this type

		// our plugin might implement one of the needed types
		// if not, check if the plugin provides some types which might implement the type
		if k.Implements(tp) {
			impl = append(impl, &implements{
				plugin: entry.Plugin(),
				weight: entry.Weight(),
			})
			continue
		}

		// here we check that provides
		for j := 0; j < len(entry.returnedTypes); j++ {
			provided := entry.returnedTypes[j]
			if provided.retType.Implements(tp) {
				impl = append(impl,
					&implements{
						plugin:  entry.Plugin(),
						weight:  entry.Weight(),
						methods: provided.method,
					},
				)
			}
		}
	}

	// sort by weight
	sort.Slice(impl, func(i, j int) bool {
		return impl[i].weight > impl[j].weight
	})

	return impl
}
