package vertex

import (
	"reflect"
	"sync/atomic"

	"github.com/spiral/endure/pkg/fsm"
)

// CollectorEntry entry is collector interface struct which contain:
// 1. in types like func (a int, b string etc...)
// 2 function name
type CollectorEntry struct {
	In []In
	Fn string
}

// In struct represents In value, which contain:
// 1. 'in' as reflect.Value
// 2. dep name as string
type In struct {
	In  reflect.Value
	Dep string
}

// ProviderEntry is Provides interface method representation. It consists of:
// 1. Function name
// 2. Return type Ids (strings), for example foo.S2
type ProviderEntry struct {
	FunctionName  string
	ReturnTypeIds []string
}

// Entry is the general entry used in InitDepsToInvoke, CollectsDepsToInvoke, addToList and etc..
type Entry struct {
	// RefID, structure, which provides interface dep
	RefID string
	// Name of the entry
	Name string
	// IsReference, can be true, false or nil (unknown)
	IsReference *bool
	// IsDisabled retrun true if vertex returns errors.Disabled
	IsDisabled bool
	// Kind is just reflect.Kind
	Kind reflect.Kind
}

// ProviderEntries is type alias for the ProviderEntry slice
type ProviderEntries []ProviderEntry

// FnsToCall is slice with the functions which can return the same resulting set of values
// for example
// fn a() b --|
// fn c() b ----> both a and c returns b
type FnsToCall [][]string

// Merge creates FnsToCall
func (pe *ProviderEntries) Merge() FnsToCall {
	res := make(FnsToCall, len(*pe))
	hash := make(map[[10]string][]string)
	for i := 0; i < len(*pe); i++ {
		arr := [10]string{}
		for j := 0; j < len((*pe)[i].ReturnTypeIds); j++ {
			arr[j] = (*pe)[i].ReturnTypeIds[j]
		}
		hash[arr] = append(hash[arr], (*pe)[i].FunctionName)
	}

	index := 0
	for _, v := range hash {
		for i := 0; i < len(v); i++ {
			res[index] = append(res[index], v[i])
		}
		index++
	}
	if index < len(res) {
		res = res[:index]
	}
	return res
}

// CollectorEntries is type alias for the CollectorEntries slice
type CollectorEntries []CollectorEntry

// Meta information included into the Vertex
// May include:
// 1. Disabled info
// 2. Relation status
type Meta struct {
	// FnsProviderToInvoke is the function names to invoke if type implements Provides() interface
	FnsProviderToInvoke ProviderEntries
	// CollectorEntries is the function names to invoke if type implements Collector() interface
	CollectorEntries CollectorEntries

	// List of the vertex deps
	// foo4.DB, foo4.S4 etc.. which were found in the Init() method
	InitDepsToInvoke map[string][]Entry
	InitDepsOrd      []string

	// List of the vertex deps
	// foo4.DB, foo4.S4 etc.. which were found in the Collects() method
	CollectsDepsToInvoke map[string][]Entry
}

func NewMeta() Meta {
	meta := Meta{
		FnsProviderToInvoke:  make(ProviderEntries, 0),
		CollectorEntries:     make(CollectorEntries, 0),
		InitDepsToInvoke:     make(map[string][]Entry),
		InitDepsOrd:          make([]string, 0),
		CollectsDepsToInvoke: make(map[string][]Entry),
	}
	return meta
}

// Vertex is main vertex representation for the graph
// since we can have cyclic dependencies
// when we traverse the VerticesMap, we should mark nodes as visited or not to detect cycle
type Vertex struct {
	// ID of the vertex, currently string representation of the structure fn
	ID string
	// Vertex (Registered structure)
	Iface interface{}
	// Meta information about current Vertex
	Meta Meta
	// Dependencies of the node
	Dependencies []*Vertex
	// Set of entries which can vertex provide (for example, foo4 vertex can provide DB instance and logger)
	Provides map[string]ProvidedEntry

	// If vertex disabled it removed from the processing (Init, Serve, Stop), but present in the graph
	IsDisabled bool
	// for the topological sort, private
	NumOfDeps int
	Visited   bool
	Visiting  bool

	// current state
	state uint32
}

func NewVertex() *Vertex {
	vertex := &Vertex{
		ID:           "",
		Iface:        nil,
		Meta:         NewMeta(),
		Dependencies: make([]*Vertex, 0),
		Provides:     make(map[string]ProvidedEntry),
		IsDisabled:   false,
		NumOfDeps:    0,
		Visited:      false,
		Visiting:     false,
		state:        0,
	}
	return vertex
}

// ProvidedEntry is proviers helper entity
type ProvidedEntry struct {
	Str string
	// we need to distinguish false (default bool value) and nil --> we don't know information about reference
	IsReference *bool
	Value       reflect.Value
	Kind        reflect.Kind
}

// AddProvider adds an provider for a dep (vertex->vertex)
func (v *Vertex) AddProvider(valueKey string, value reflect.Value, isRef bool, kind reflect.Kind) {
	v.Provides[valueKey] = ProvidedEntry{
		Str:         valueKey,
		IsReference: &isRef,
		Value:       value,
		Kind:        kind,
	}
}

// RemoveProvider removes provider from the map
func (v *Vertex) RemoveProvider(valueKey string) {
	delete(v.Provides, valueKey)
}

// SetState sets the state for the vertex
func (v *Vertex) SetState(st fsm.State) {
	atomic.StoreUint32(&v.state, uint32(st))
}

// GetState gets the vertex state
func (v *Vertex) GetState() fsm.State {
	return fsm.State(atomic.LoadUint32(&v.state))
}
