package endure

import (
	"fmt"
	"reflect"
	"sync/atomic"

	"github.com/spiral/errors"
)

// Kind used to identify the method which invokes AddDeps
type Kind int

const (
	// Init method
	Init Kind = iota
	// Collects method
	Collects
)

// Vertices is type alias for the slice of vertices
type Vertices []*Vertex

// Graph manages the set of services and their edges
// type of the VerticesMap: directed
type Graph struct {
	// Map with vertices to have an easy access to it
	VerticesMap map[string]*Vertex
	// List of all Vertices
	Vertices []*Vertex

	providers map[string]reflect.Value
}

// ProviderEntries is type alias for the ProviderEntry slice
type ProviderEntries []ProviderEntry

// CollectorEntries is type alias for the CollectorEntries slice
type CollectorEntries []CollectorEntry

// Meta information included into the Vertex
// May include:
// 1. Disabled info
// 2. Relation status
type Meta struct {
	// Position in code while invoking Register
	Order int
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

// CollectorEntry entry is collector interface struct which contain:
// 1. in types like func (a int, b string etc...)
// 2 function name
type CollectorEntry struct {
	in []In
	fn string
}

// In struct represents In value, which contain:
// 1. 'in' as reflect.Value
// 2. dep name as string
type In struct {
	in  reflect.Value
	dep string
}

// ProviderEntry is Provides interface method representation. It consists of:
// 1. Function name
// 2. Return type Ids (strings), for example foo.S2
type ProviderEntry struct {
	FunctionName  string
	ReturnTypeIds []string
}

// FnsToCall is slice with the functions which can return the same resulting set of values
// for example
// fn a() b --|
// fn c() b ----> both a and c returns b
type FnsToCall [][]string

// Merge creates FnsToCall
func (pe *ProviderEntries) Merge() FnsToCall {
	res := make(FnsToCall, len(*pe), len(*pe))
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
	numOfDeps int
	visited   bool
	visiting  bool

	// current state
	state uint32
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
	if v.Provides == nil {
		v.Provides = make(map[string]ProvidedEntry)
	}

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
func (v *Vertex) SetState(st State) {
	atomic.StoreUint32(&v.state, uint32(st))
}

// GetState gets the vertex state
func (v *Vertex) GetState() State {
	return State(atomic.LoadUint32(&v.state))
}

// DisableByID used to disable vertex by it's ID
func (g *Graph) DisableByID(vid string) {
	v := g.VerticesMap[vid]
	for i := 0; i < len(g.Vertices); i++ {
		g.disablerHelper(g.Vertices[i], v)
	}
}

func (g *Graph) disablerHelper(vertex *Vertex, disabled *Vertex) bool {
	if vertex.ID == disabled.ID {
		return true
	}
	for i := 0; i < len(vertex.Dependencies); i++ {
		contains := g.disablerHelper(vertex.Dependencies[i], disabled)
		if contains {
			vertex.IsDisabled = true
			return true
		}
	}
	return false
}

// NewGraph initializes endure Graph
// According to the topological sorting, graph should be
// 1. DIRECTED
// 2. ACYCLIC
func NewGraph() *Graph {
	return &Graph{
		VerticesMap: make(map[string]*Vertex),
		providers:   make(map[string]reflect.Value),
	}
}

// AddGlobalProvider adds provider to the global map in the Graph structure
func (g *Graph) AddGlobalProvider(providedID string, val reflect.Value) {
	g.providers[providedID] = val
}

// HasVertex returns true or false if the vertex exists in the vertices map in the graph
func (g *Graph) HasVertex(name string) bool {
	_, ok := g.VerticesMap[name]
	return ok
}

/*
AddDep doing the following:
1. Get a vertexID (foo2.S2 for example)
2. Get a depID --> could be vertexID of vertex dep ID like foo2.DB
3. Need to find vertexID to provide dependency. Example foo2.DB is actually foo2.S2 vertex
*/
func (g *Graph) AddDep(vertex *Vertex, depID string, method Kind, isRef bool, typeKind reflect.Kind) error {
	switch typeKind {
	case reflect.Interface:
		err := g.addInterfaceDep(vertex, depID, method, isRef)
		if err != nil {
			return err
		}
	default:
		err := g.addStructDep(vertex, depID, method, isRef)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *Graph) addInterfaceDep(vertex *Vertex, depID string, method Kind, isRef bool) error {
	const op = errors.Op("add interface dep")

	// here can be a lot of deps
	depVertex := g.FindProviders(depID)
	if depVertex == nil {
		return errors.E(op, errors.Errorf("can't find dependency: %s for the vertex: %s", depID, vertex.ID))
	}

	// skip self
	if depVertex.ID == vertex.ID {
		return nil
	}

	// add Dependency into the List
	// to call later
	// because we should know Init method parameters for every Vertex
	// for example, we should know http.Middleware dependency and later invoke all types which it implement
	// OR know Collects methods to invoke
	if g.addToList(method, vertex, depID, isRef, depVertex.ID, reflect.Interface) == false {
		return nil
	}

	for j := 0; j < len(depVertex.Dependencies); j++ {
		tmpID := depVertex.Dependencies[j].ID
		if tmpID == vertex.ID {
			return nil
		}
	}

	vertex.numOfDeps++
	vertex.Dependencies = append(vertex.Dependencies, depVertex)
	return nil
}

// Add meta information to the InitDepsToInvoke or CollectsDepsToInvoke
func (g *Graph) addToList(method Kind, vertex *Vertex, depID string, isRef bool, refID string, kind reflect.Kind) bool {
	switch method {
	case Init:
		if vertex.Meta.InitDepsToInvoke == nil {
			vertex.Meta.InitDepsToInvoke = make(map[string][]Entry)
		}
		vertex.Meta.InitDepsToInvoke[refID] = append(vertex.Meta.InitDepsToInvoke[refID], Entry{
			RefID:       refID,
			Name:        depID,
			IsReference: &isRef,
			Kind:        kind,
		})
		contains := false
		for _, v := range vertex.Meta.InitDepsOrd {
			if v == refID {
				contains = true
			}
		}
		if !contains {
			vertex.Meta.InitDepsOrd = append(vertex.Meta.InitDepsOrd, refID)
		}
	case Collects:
		if vertex.Meta.CollectsDepsToInvoke == nil {
			vertex.Meta.CollectsDepsToInvoke = make(map[string][]Entry)
			vertex.Meta.CollectsDepsToInvoke[refID] = append(vertex.Meta.CollectsDepsToInvoke[refID], Entry{
				RefID:       refID,
				Name:        depID,
				IsReference: &isRef,
				Kind:        kind,
			})

			contains := false
			for _, v := range vertex.Meta.InitDepsOrd {
				if v == refID {
					contains = true
				}
			}
			if !contains {
				vertex.Meta.InitDepsOrd = append(vertex.Meta.InitDepsOrd, refID)
			}
		} else {
			if _, ok := vertex.Meta.CollectsDepsToInvoke[refID]; ok {
				return false
			}
			vertex.Meta.CollectsDepsToInvoke[refID] = append(vertex.Meta.CollectsDepsToInvoke[refID], Entry{
				RefID:       refID,
				Name:        depID,
				IsReference: &isRef,
				Kind:        kind,
			})
			contains := false
			for _, v := range vertex.Meta.InitDepsOrd {
				if v == refID {
					contains = true
				}
			}
			if !contains {
				vertex.Meta.InitDepsOrd = append(vertex.Meta.InitDepsOrd, refID)
			}
		}
	}
	return true
}

func (g *Graph) addStructDep(vertex *Vertex, depID string, method Kind, isRef bool) error {
	const op = errors.Op("add structure dep")
	// vertex should always present

	// but depVertex can be represented like foo2.S2 (vertexID) or like foo2.DB (vertex foo2.S2, dependency foo2.DB)
	depVertex := g.GetVertex(depID)
	if depVertex == nil {
		tmp := g.FindProviders(depID)
		if tmp == nil {
			return errors.E(op, errors.Errorf("can't find dep: %s for the vertex: %s", depID, vertex.ID))
		}
		depVertex = tmp
	}

	// add Dependency into the List
	// to call later
	// because we should know Init method parameters for every Vertex
	if !g.addToList(method, vertex, depID, isRef, depVertex.ID, reflect.Struct) {
		return nil
	}

	// append depID vertex
	for i := 0; i < len(depVertex.Dependencies); i++ {
		tmpID := depVertex.Dependencies[i].ID
		if tmpID == vertex.ID {
			return nil
		}
	}

	vertex.numOfDeps++
	vertex.Dependencies = append(vertex.Dependencies, depVertex)
	return nil
}

// Reset resets vertices to initial state
func (g *Graph) Reset(vertex *Vertex) []*Vertex {
	// restore number of dependencies for the root
	vertex.numOfDeps = len(vertex.Dependencies)
	vertex.visiting = false
	vertex.visited = false
	vertex.SetState(Uninitialized)
	vertices := make([]*Vertex, 0, 5)
	vertices = append(vertices, vertex)

	tmp := make(map[string]*Vertex)

	g.depthFirstSearch(vertex.Dependencies, tmp)

	for _, v := range tmp {
		vertices = append(vertices, v)
	}
	return vertices
}

// actually this is DFS just to reset all vertices to initial state after topological sort
func (g *Graph) depthFirstSearch(deps []*Vertex, tmp map[string]*Vertex) {
	for i := 0; i < len(deps); i++ {
		deps[i].visited = false
		deps[i].visiting = false
		deps[i].numOfDeps = len(deps)
		tmp[deps[i].ID] = deps[i]
		g.depthFirstSearch(deps[i].Dependencies, tmp)
	}
}

// AddVertex adds an vertex to the graph with its ID, value and meta information
func (g *Graph) AddVertex(vertexID string, vertexIface interface{}, meta Meta) {
	v := &Vertex{
		ID:           vertexID,
		Iface:        vertexIface,
		Meta:         meta,
		Dependencies: nil,
	}
	v.SetState(Uninitialized)
	g.VerticesMap[vertexID] = v

	g.Vertices = append(g.Vertices, g.VerticesMap[vertexID])
}

// GetVertex returns vertex by its ID
func (g *Graph) GetVertex(id string) *Vertex {
	return g.VerticesMap[id]
}

// FindProviders finds provider deps for the vertex and returns dependent vertices
func (g *Graph) FindProviders(depID string) *Vertex {
	for i := 0; i < len(g.Vertices); i++ {
		for providerID := range g.Vertices[i].Provides {
			if depID == providerID {
				return g.Vertices[i]
			}
		}
	}

	// try to find directly in the graph
	if _, ok := g.VerticesMap[depID]; ok {
		return g.VerticesMap[depID]
	}

	return nil
}

// TopologicalSort topologically sort the graph and return slice of the sorted vertices
func TopologicalSort(vertices []*Vertex) ([]*Vertex, error) {
	const op = errors.Op("topological sort")
	var ord Vertices
	verticesCopy := vertices

	for len(verticesCopy) != 0 {
		vertex := verticesCopy[len(verticesCopy)-1]
		verticesCopy = verticesCopy[:len(verticesCopy)-1]
		containsCycle := dfs(vertex, &ord)
		if containsCycle {
			// If we found a cycle, print involved vertices
			for i := 0; i < len(vertices); i++ {
				if vertices[i].visited == false {
					fmt.Println(vertices[i].ID)
				}
			}
			return nil, errors.E(op, errors.Errorf("cycle detected, please, check vertex: %s", vertex.ID))
		}
	}

	return ord, nil
}

func dfs(vertex *Vertex, ordered *Vertices) bool {
	// if vertex already visited, we reach the end, return
	if vertex.visited {
		return false
		// if vertex is visiting, than we have cyclic dep, because we reach the same vertex from two different dependencies
	} else if vertex.visiting {
		return true
	}
	vertex.visiting = true
	for _, depV := range vertex.Dependencies {
		containsCycle := dfs(depV, ordered)
		if containsCycle {
			return true
		}
	}
	vertex.visited = true
	vertex.visiting = false
	*ordered = append(*ordered, vertex)
	return false
}
