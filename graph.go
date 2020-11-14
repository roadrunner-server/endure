package endure

import (
	"fmt"
	"reflect"
	"sync/atomic"

	"github.com/spiral/errors"
)

type Kind int

const (
	Init Kind = iota
	Collects
)

type Vertices []*Vertex

// manages the set of services and their edges
// type of the VerticesMap: directed
type Graph struct {
	// Map with vertices to have an easy access to it
	VerticesMap map[string]*Vertex
	// List of all Vertices
	Vertices []*Vertex

	providers map[string]reflect.Value
}

type ProviderEntries []ProviderEntry
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

type CollectorEntry struct {
	in []In
	fn string
}

type In struct {
	in  reflect.Value
	dep string
}

type ProviderEntry struct {
	FunctionName  string
	ReturnTypeIds []string
}

type FnsToCall [][]string

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

type Entry struct {
	// Reference ID, structure, which provides interface dep
	RefId       string
	Name        string
	IsReference *bool
	IsDisabled  bool
	Kind        reflect.Kind
}

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

	// current state
	state uint32

	// If vertex disabled it removed from the processing (Init, Serve, Stop), but present in the graph
	IsDisabled bool
	// for the topological sort, private
	numOfDeps int
	visited   bool
	visiting  bool
}

type ProvidedEntry struct {
	Str string
	// we need to distinguish false (default bool value) and nil --> we don't know information about reference
	IsReference *bool
	Value       reflect.Value
	Kind        reflect.Kind
}

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

func (v *Vertex) RemoveProvider(valueKey string) {
	delete(v.Provides, valueKey)
}

func (v *Vertex) SetState(st State) {
	atomic.StoreUint32(&v.state, uint32(st))
}

func (v *Vertex) GetState() State {
	return State(atomic.LoadUint32(&v.state))
}

func (g *Graph) DisableById(vid string) {
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

func (g *Graph) AddGlobalProvider(providedId string, val reflect.Value) {
	g.providers[providedId] = val
}

func (g *Graph) HasVertex(name string) bool {
	_, ok := g.VerticesMap[name]
	return ok
}

/*
AddDepRev doing the following:
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
	depVertices := g.FindProviders(depID)
	if len(depVertices) == 0 {
		return errors.E(op, errors.Errorf("can't find dependency: %s for the vertex: %s", depID, vertex.ID))
	}

	// skip self
	for i := 0; i < len(depVertices); i++ {
		if depVertices[i].ID == vertex.ID {
			depVertices = append(depVertices[:i], depVertices[i+1:]...)
		}
	}

	for i := 0; i < len(depVertices); i++ {
		// add Dependency into the List
		// to call later
		// because we should know Init method parameters for every Vertex
		// for example, we should know http.Middleware dependency and later invoke all types which it implement
		// OR know Collects methods to invoke
		if g.addToList(method, vertex, depID, isRef, depVertices[i].ID, reflect.Interface) == false {
			return nil
		}

		for j := 0; j < len(depVertices[i].Dependencies); j++ {
			tmpID := depVertices[i].Dependencies[j].ID
			if tmpID == vertex.ID {
				return nil
			}
		}

		vertex.numOfDeps++
		vertex.Dependencies = append(vertex.Dependencies, depVertices[i])
	}
	return nil
}

// Add meta information to the InitDepsToInvoke or CollectsDepsToInvoke
func (g *Graph) addToList(method Kind, vertex *Vertex, depID string, isRef bool, refId string, kind reflect.Kind) bool {
	switch method {
	case Init:
		if vertex.Meta.InitDepsToInvoke == nil {
			vertex.Meta.InitDepsToInvoke = make(map[string][]Entry)
		}
		vertex.Meta.InitDepsToInvoke[refId] = append(vertex.Meta.InitDepsToInvoke[refId], Entry{
			RefId:       refId,
			Name:        depID,
			IsReference: &isRef,
			Kind:        kind,
		})
		contains := false
		for _, v := range vertex.Meta.InitDepsOrd {
			if v == refId {
				contains = true
			}
		}
		if !contains {
			vertex.Meta.InitDepsOrd = append(vertex.Meta.InitDepsOrd, refId)
		}
	case Collects:
		if vertex.Meta.CollectsDepsToInvoke == nil {
			vertex.Meta.CollectsDepsToInvoke = make(map[string][]Entry)
			vertex.Meta.CollectsDepsToInvoke[refId] = append(vertex.Meta.CollectsDepsToInvoke[refId], Entry{
				RefId:       refId,
				Name:        depID,
				IsReference: &isRef,
				Kind:        kind,
			})

			contains := false
			for _, v := range vertex.Meta.InitDepsOrd {
				if v == refId {
					contains = true
				}
			}
			if !contains {
				vertex.Meta.InitDepsOrd = append(vertex.Meta.InitDepsOrd, refId)
			}
		} else {
			if _, ok := vertex.Meta.CollectsDepsToInvoke[refId]; ok {
				return false
			}
			vertex.Meta.CollectsDepsToInvoke[refId] = append(vertex.Meta.CollectsDepsToInvoke[refId], Entry{
				RefId:       refId,
				Name:        depID,
				IsReference: &isRef,
				Kind:        kind,
			})
			contains := false
			for _, v := range vertex.Meta.InitDepsOrd {
				if v == refId {
					contains = true
				}
			}
			if !contains {
				vertex.Meta.InitDepsOrd = append(vertex.Meta.InitDepsOrd, refId)
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
		if len(tmp) > 0 {
			// here can be only 1 Dep for the struct, or PANIC!!!
			depVertex = g.FindProviders(depID)[0]
		} else {
			return fmt.Errorf("can't find dep: %s for the vertex: %s", depID, vertex.ID)
		}
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

// reset vertices to initial state
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

func (g *Graph) GetVertex(id string) *Vertex {
	return g.VerticesMap[id]
}

func (g *Graph) FindProviders(depID string) []*Vertex {
	ret := make([]*Vertex, 0, 2)
	for i := 0; i < len(g.Vertices); i++ {
		for providerID := range g.Vertices[i].Provides {
			if depID == providerID {
				ret = append(ret, g.Vertices[i])
			}
		}
	}

	// try to find directly in the graph
	if len(ret) == 0 {
		if v, ok := g.VerticesMap[depID]; ok {
			ret = append(ret, v)
		}
	}
	return ret
}

func TopologicalSort(vertices []*Vertex) ([]*Vertex, error) {
	const op = errors.Op("topological sort")
	var ord Vertices
	verticesCopy := vertices

	for len(verticesCopy) != 0 {
		vertex := verticesCopy[len(verticesCopy)-1]
		verticesCopy = verticesCopy[:len(verticesCopy)-1]
		containsCycle := dfs(vertex, &ord)
		if containsCycle {
			return nil, errors.E(op, errors.Errorf("cycle detected, please, check vertex: %s", vertex.ID))
		}
	}

	return ord, nil
}

func dfs(vertex *Vertex, ordered *Vertices) bool {
	if vertex.visited {
		return false
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
