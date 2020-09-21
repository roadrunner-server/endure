package structures

import (
	"errors"
	"fmt"
	"reflect"
)

type Kind int

const (
	Init Kind = iota
	Depends
)

// manages the set of services and their edges
// type of the VerticesMap: directed
type Graph struct {
	// Map with vertices to have an easy access to it
	VerticesMap map[string]*Vertex
	// List of all Vertices
	Vertices []*Vertex
}

// Meta information included into the Vertex
// May include:
// 1. Disabled info
// 2. Relation status
type Meta struct {
	// Position in code while invoking Register
	Order int
	// FnsProviderToInvoke is the function names to invoke if type implements Provides() interface
	FnsProviderToInvoke []string
	// FnsDependerToInvoke is the function names to invoke if type implements Depender() interface
	FnsDependerToInvoke []string

	// List of the vertex deps
	// foo4.DB, foo4.S4 etc.. which were found in the Init() method
	InitDepsList []DepsEntry

	// List of the vertex deps
	// foo4.DB, foo4.S4 etc.. which were found in the Depends() method
	DepsList []DepsEntry
}

type DepsEntry struct {
	Name        string
	IsReference *bool
	Kind        reflect.Kind
}

// since we can have cyclic dependencies
// when we traverse the VerticesMap, we should mark nodes as visited or not to detect cycle
type Vertex struct {
	// ID of the vertex, currently string representation of the structure name
	ID string
	// Vertex (Registered structure)
	Iface interface{}
	// Meta information about current Vertex
	Meta Meta
	// Dependencies of the node
	Dependencies []*Vertex
	// Set of entries which can vertex provide (for example, foo4 vertex can provide DB instance and logger)
	Provides map[string]ProvidedEntry

	// for the topological sort, private
	numOfDeps int
	visited   bool
	visiting  bool
}

type ProvidedEntry struct {
	Str string
	// we need to distinguish false (default bool value) and nil --> we don't know information about reference
	IsReference *bool
	Value       *reflect.Value
	Kind        reflect.Kind
}

func (v *Vertex) AddProvider(valueKey string, value reflect.Value, isRef bool, kind reflect.Kind) {
	if v.Provides == nil {
		v.Provides = make(map[string]ProvidedEntry)
	}

	v.Provides[valueKey] = ProvidedEntry{
		Str:         valueKey,
		IsReference: &isRef,
		Value:       &value,
		Kind:        kind,
	}
}

// NewGraph initializes endure Graph
// According to the topological sorting, graph should be
// 1. DIRECTED
// 2. ACYCLIC
func NewGraph() *Graph {
	return &Graph{
		VerticesMap: make(map[string]*Vertex),
	}
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
func (g *Graph) AddDep(vertexID, depID string, method Kind, isRef bool, typeKind reflect.Kind) error {
	switch typeKind {
	case reflect.Interface:
		err := g.addInterfaceDep(vertexID, depID, method, isRef)
		if err != nil {
			return err
		}
	default:
		err := g.addStructDep(vertexID, depID, method, isRef)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *Graph) addInterfaceDep(vertexID, depID string, method Kind, isRef bool) error {
	// vertex should always present
	vertex := g.GetVertex(vertexID)
	if vertex == nil {
		return errors.New("vertex should be in the graph")
	}

	// here can be a lot of deps
	depVertices := g.FindProviders(depID)

	if depVertices == nil {
		return fmt.Errorf("can't find dep: %s for the vertex: %s", depID, vertexID)
	}

	for i := 0; i < len(depVertices); i++ {
		// add Dependency into the List
		// to call later
		// because we should know Init method parameters for every Vertex
		// for example, we should know http.Middleware dependency and later invoke all types which it implement
		// OR know Depends methods to invoke
		g.addToList(method, vertex, depID, isRef, reflect.Interface)

		// append depID vertex
		for j := 0; j < len(depVertices[i].Dependencies); j++ {
			tmpID := depVertices[i].Dependencies[i].ID
			if tmpID == vertex.ID {
				return nil
			}
		}

		depVertices[i].numOfDeps++
		depVertices[i].Dependencies = append(depVertices[i].Dependencies, vertex)
	}
	return nil
}

// Add meta information to the InitDepsList or DepsList
func (g *Graph) addToList(method Kind, vertex *Vertex, depID string, isRef bool, kind reflect.Kind) {
	switch method {
	case Init:
		if vertex.Meta.InitDepsList == nil {
			vertex.Meta.InitDepsList = make([]DepsEntry, 0, 1)
		}
		vertex.Meta.InitDepsList = append(vertex.Meta.InitDepsList, DepsEntry{
			Name:        depID,
			IsReference: &isRef,
			Kind:        kind,
		})
	case Depends:
		if vertex.Meta.DepsList == nil {
			vertex.Meta.DepsList = make([]DepsEntry, 0, 1)
			vertex.Meta.DepsList = append(vertex.Meta.DepsList, DepsEntry{
				Name:        depID,
				IsReference: &isRef,
				Kind:        kind,
			})
		} else {
			// search if DepsList already contains interface dep
			for _, v := range vertex.Meta.DepsList {
				if v.Name == depID {
					continue
				}

				vertex.Meta.DepsList = append(vertex.Meta.DepsList, DepsEntry{
					Name:        depID,
					IsReference: &isRef,
					Kind:        kind,
				})
			}
		}
	}
}

func (g *Graph) addStructDep(vertexID, depID string, method Kind, isRef bool) error {
	// vertex should always present
	vertex := g.GetVertex(vertexID)
	if vertex == nil {
		return errors.New("vertex should be in the graph")
	}
	// but depVertex can be represented like foo2.S2 (vertexID) or like foo2.DB (vertex foo2.S2, dependency foo2.DB)
	depVertex := g.GetVertex(depID)
	if depVertex == nil {
		tmp := g.FindProviders(depID)
		if len(tmp) > 0 {
			// here can be only 1 Dep for the struct, or PANIC!!!
			depVertex = g.FindProviders(depID)[0]
		} else {
			return fmt.Errorf("can't find dep: %s for the vertex: %s", depID, vertexID)
		}
	}

	// add Dependency into the List
	// to call later
	// because we should know Init method parameters for every Vertex
	g.addToList(method, vertex, depID, isRef, reflect.Struct)

	// append depID vertex
	for i := 0; i < len(depVertex.Dependencies); i++ {
		tmpID := depVertex.Dependencies[i].ID
		if tmpID == vertex.ID {
			return nil
		}
	}

	depVertex.numOfDeps++
	depVertex.Dependencies = append(depVertex.Dependencies, vertex)
	return nil
}

// reset vertices to initial state
func (g *Graph) Reset(vertex *Vertex) []*Vertex {
	// restore number of dependencies for the root
	vertex.numOfDeps = len(vertex.Dependencies)
	vertex.visiting = false
	vertex.visited = false
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
	g.VerticesMap[vertexID] = &Vertex{
		ID:           vertexID,
		Iface:        vertexIface,
		Meta:         meta,
		Dependencies: nil,
	}
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
	return ret
}

type Vertices []*Vertex

func (v Vertices) Len() int {
	return len(v)
}
func (v Vertices) Less(i, j int) bool {
	return v[i].Meta.Order < v[j].Meta.Order
}
func (v Vertices) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func TopologicalSort(vertices []*Vertex) []*Vertex {
	var ord Vertices
	verticesCopy := vertices

	for len(verticesCopy) != 0 {
		vertex := verticesCopy[len(verticesCopy)-1]
		verticesCopy = verticesCopy[:len(verticesCopy)-1]
		containsCycle := dfs(vertex, &ord)
		if containsCycle {
			return nil
		}
	}

	return ord
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
