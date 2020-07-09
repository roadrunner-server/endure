package structures

import (
	"fmt"
	"reflect"
	"sort"
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
	// FnsDependerToInvoke is the function names to invoke if type implements Register() interface
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
// when we traverse the VerticesMap, we should mark nodes as Visited or not to detect cycle
type Vertex struct {
	Id string
	// Vertex
	Iface interface{}
	// Meta information about current Vertex
	Meta Meta
	// Dependencies of the node
	Dependencies []*Vertex

	// Vertex foo4.S4 also provides (for example)
	// foo4.DB
	Provides map[string]ProvidedEntry

	// for the toposort
	NumOfDeps int

	Visited  bool
	Visiting bool
}

type ProvidedEntry struct {
	Str string
	// we need to distinguish false (default bool value) and nil --> we don't know information about reference
	IsReference *bool
	Value       *reflect.Value
	Kind        reflect.Kind
}

func (v *Vertex) AddProvider(valueKey string, value reflect.Value, isRef bool, kind reflect.Kind) error {
	if v.Provides == nil {
		v.Provides = make(map[string]ProvidedEntry)
	}

	v.Provides[valueKey] = ProvidedEntry{
		Str:         valueKey,
		IsReference: &isRef,
		Value:       &value,
		Kind:        kind,
	}
	return nil
}

// NewAL initializes adjacency list to store the VerticesMap
// example
// 1 -> 2 -> 4
// 2 -> 5
// 3 -> 6 -> 5
// 4 -> 2
// 5 -> 4
// 6 -> 6
//
// VerticesMap from the AL:
//
//+---+          +---+               +---+
//| 1 +--------->+ 2 |               | 3 |
//+-+-+          +--++               +-+-+
//  |          +-+  |             +-+  |
//  |        +-+    |           +-+    |
//  |       ++      |          ++      |
//  v     +-+       v        +-+       v
//+-+-+<--+      +--++       |       +-+-+
//| 4 |     +----+ 5 +<------+       | 6 +<-+
//+---+<----+    +---+               +-+-+  |
//                                     |    |
//                                     +----+
// BUT
// According to the topological sorting, graph should be
// 1. DIRECTED
// 2. ACYCLIC
//
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
3. Need to find VertexID to provide dependency. Example foo2.DB is actually foo2.S2 vertex
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
		panic("vertex should be in the graph")
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
		g.addToList(method, vertex, depID, isRef)

		//append depID vertex
		for j := 0; j < len(depVertices[i].Dependencies); j++ {
			tmpId := depVertices[i].Dependencies[i].Id
			if tmpId == vertex.Id {
				return nil
			}
		}

		depVertices[i].NumOfDeps++
		depVertices[i].Dependencies = append(depVertices[i].Dependencies, vertex)
	}
	return nil
}

// Add meta information to the InitDepsList or DepsList
func (g *Graph) addToList(method Kind, vertex *Vertex, depID string, isRef bool) {
	switch method {
	case Init:
		if vertex.Meta.InitDepsList == nil {
			vertex.Meta.InitDepsList = make([]DepsEntry, 0, 1)
		}
		vertex.Meta.InitDepsList = append(vertex.Meta.InitDepsList, DepsEntry{
			Name:        depID,
			IsReference: &isRef,
			Kind:        reflect.Interface,
		})
	case Depends:
		if vertex.Meta.DepsList == nil {
			vertex.Meta.DepsList = make([]DepsEntry, 0, 1)
			vertex.Meta.DepsList = append(vertex.Meta.DepsList, DepsEntry{
				Name:        depID,
				IsReference: &isRef,
				Kind:        reflect.Interface,
			})
		} else {
			// search if DepsList already contains interface dep
			for _, v := range vertex.Meta.DepsList {
				if v.Name == depID {
					continue
				}
			}
		}

	}
}

func (g *Graph) addStructDep(vertexID, depID string, method Kind, isRef bool) error {
	// vertex should always present
	vertex := g.GetVertex(vertexID)
	if vertex == nil {
		panic("vertex should be in the graph")
	}
	// but depVertex can be represented like foo2.S2 (vertexID) or like foo2.DB (vertex foo2.S2, dependency foo2.DB)
	depVertex := g.GetVertex(depID)
	if depVertex == nil {
		// here can be only 1 Dep for the struct, or PANIC!!!
		depVertex = g.FindProviders(depID)[0]
	}
	if depVertex == nil {
		return fmt.Errorf("can't find dep: %s for the vertex: %s", depID, vertexID)
	}

	// add Dependency into the List
	// to call later
	// because we should know Init method parameters for every Vertex
	g.addToList(method, vertex, depID, isRef)

	// append depID vertex
	for i := 0; i < len(depVertex.Dependencies); i++ {
		tmpId := depVertex.Dependencies[i].Id
		if tmpId == vertex.Id {
			return nil
		}
	}

	depVertex.NumOfDeps++
	depVertex.Dependencies = append(depVertex.Dependencies, vertex)
	return nil
}

func (g *Graph) AddVertex(vertexId string, vertexIface interface{}, meta Meta) {
	g.VerticesMap[vertexId] = &Vertex{
		Id:           vertexId,
		Iface:        vertexIface,
		Meta:         meta,
		Dependencies: nil,
	}
	g.Vertices = append(g.Vertices, g.VerticesMap[vertexId])
}

func (g *Graph) GetVertex(id string) *Vertex {
	return g.VerticesMap[id]
}

func (g *Graph) FindProviders(depId string) []*Vertex {
	ret := make([]*Vertex, 0, 2)
	for i := 0; i < len(g.Vertices); i++ {
		for providerId := range g.Vertices[i].Provides {
			if depId == providerId {
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
	var tmpZeroDeps Vertices

	for _, v := range ord {
		if len(v.Dependencies) == 0 {
			tmpZeroDeps = append(tmpZeroDeps, v)
		}
	}

	sort.Sort(tmpZeroDeps)

	return ord
}

func dfs(vertex *Vertex, ordered *Vertices) bool {
	if vertex.Visited {
		return false
	} else if vertex.Visiting {
		return true
	}
	vertex.Visiting = true
	for _, depV := range vertex.Dependencies {
		containsCycle := dfs(depV, ordered)
		if containsCycle {
			return true
		}
	}
	vertex.Visited = true
	vertex.Visiting = false
	*ordered = append(*ordered, vertex)
	return false
}
