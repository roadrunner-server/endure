package graph

type EdgeType string

const (
	InitConnection     EdgeType = "InitConnection"
	CollectsConnection EdgeType = "CollectsConnection"
)

type edge struct {
	src, dest      any
	connectionType EdgeType
}
