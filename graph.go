package graph

import (
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrVertexNotFound      = errors.New("vertex not found")
	ErrVertexAlreadyExists = errors.New("vertex already exists")
	ErrEdgeNotFound        = errors.New("edge not found")
	ErrEdgeAlreadyExists   = errors.New("edge already exists")
	ErrEdgeCreatesCycle    = errors.New("edge would create a cycle")
	ErrVertexHasEdges      = errors.New("vertex has edges")
)

type Vertex interface {
	// Hash returns the unique hash of the vertex. This hash is used to identify
	// the vertex in the graph. It must not contain Graph.EdgeSeparator().
	Hash() string
	// VertexType returns the type of the vertex. This is used to determine the
	// type of the vertex when creating a new vertex.
	// This is the type level method, which means that all vertices of the same type must have the same vertex type.
	// e.g. for a vertex type of "user", this method returns 'u'.
	VertexType() rune
	// Schema returns the registry of the vertex. The keys of the map are the
	// attribute names from the types which implement the Vertex interface, and the
	// values are the types of the vertex attributes.
	// This is a type level method, which means that all vertices of the same type must have the same registry.
	// e.g. for a vertex type of "user", this method returns map[string]rune{"name": 's', "age": 'i', "created_at": 't'}.
	Schema() map[string]rune
}

type EdgeChar func() rune

type Graph[T Vertex] interface {
	// Traits returns the graph's traits. Those traits must be set when creating
	// a graph using New.
	Traits() *Traits

	// AddVertex creates a new vertex in the graph. If the vertex already exists
	// in the graph, ErrVertexAlreadyExists will be returned. a valid hash key of T should
	// not contain EdgeSeparator() and must be unique.
	AddVertex(value T) error

	// Vertex returns the vertex with the given hash or ErrVertexNotFound if it
	// doesn't exist.
	Vertex(hash string) (T, error)

	Vertices() ([]string, error)

	// RemoveVertex removes the vertex with the given hash value from the graph.
	//
	// The vertex is not allowed to have edges and thus must be disconnected.
	// Potential edges must be removed first. Otherwise, ErrVertexHasEdges will
	// be returned. If the vertex doesn't exist, ErrVertexNotFound is returned.
	RemoveVertex(hash string) error

	UpdateVertex(hash string, attributes map[rune]any) error

	// EdgeSeparator returns the character used to separate the source and target
	EdgeSeparator() rune
	// AddEdge creates an edge between the source and the target vertex.
	//
	// If either vertex cannot be found, ErrVertexNotFound will be returned. If
	// the edge already exists, ErrEdgeAlreadyExists will be returned. If cycle
	// prevention has been activated using PreventCycles and if adding the edge
	// would create a cycle, ErrEdgeCreatesCycle will be returned.
	AddEdge(e Edge) error

	// Edge returns the edge joining two given vertices or ErrEdgeNotFound if
	// the edge doesn't exist. In an undirected graph, an edge with swapped
	// source and target vertices does match.
	Edge(hash string) (Edge, error)

	// Edges returns all edges in the graph. The order of the edges is not
	Edges() ([]string, error)

	// UpdateEdge updates the edge joining the two given vertices with the data
	// overwrite the existing attributes using the EdgeAttributes option.
	UpdateEdge(hash string, attributes map[rune]any) error

	// RemoveEdge removes the edge between the given source and target vertices.
	// If the edge cannot be found, ErrEdgeNotFound will be returned.
	RemoveEdge(hash string) error

	// Order returns the number of vertices in the graph.
	Order() (int, error)

	// Size returns the number of edges in the graph.
	Size() (int, error)
}

// New creates a new graph with vertices of type T, identified by hash values of
// type K. These hash values will be obtained using the provided hash function.
//
// The graph will use the default in-memory store for persisting vertices and
// edges. To use a different [Store], use [NewWithStore].
func New[T Vertex](options ...func(*Traits)) Graph[T] {
	//return NewWithStore(hash, newMemoryStore[K, T](), options...)
	return newStore[T](options...)
}

// NewWithStore creates a new graph same as [New] but uses the provided store
// instead of the default memory store.

// NewLike creates a graph that is "like" the given graph: It has the same type,
// the same hashing function, and the same traits. The new graph is independent
// of the original graph and uses the default in-memory storage.
//
//	g := graph.New(graph.IntHash, graph.Directed())
//	h := graph.NewLike(g)
//
// In the example above, h is a new directed graph of integers derived from g.

// Edge represents an edge that joins two vertices. Even though these edges are
// always referred to as source and target, whether the graph is directed or not
// is determined by its traits.
type Edge struct {
	Source     string `json:"-"` // Source vertex hash
	Target     string `json:"_r"`
	attributes map[string]any
}

func (e *Edge) MarshalJSON() ([]byte, error) {
	type Alias Edge
	return json.Marshal(&struct {
		Attributes map[string]any `json:"attributes"`
		*Alias
	}{
		Attributes: e.attributes,
		Alias:      (*Alias)(e),
	})
}

func (e *Edge) UnmarshalJSON(data []byte) error {
	type Alias Edge
	aux := &struct {
		Attributes map[string]any `json:"attributes"`
		*Alias
	}{
		Alias: (*Alias)(e),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	e.attributes = aux.Attributes
	return nil
}

func NewEdge(source, target string) *Edge {
	return &Edge{
		Source:     source,
		Target:     target,
		attributes: make(map[string]any),
	}
}

func (e *Edge) Attribute(key rune, v any) {
	e.attributes[string(key)] = v
}

func (e *Edge) StringAttr(key rune) string {
	panic("@todo")
}

func (e *Edge) IntAttr(key rune) int64 {
	panic("@todo")
}

func (e *Edge) UintAttr(key rune) uint64 {
	panic("@todo")
}

func (e *Edge) FloatAttr(key rune) float64 {
	panic("@todo")
}

func (e *Edge) BoolAttr(key rune) bool {
	panic("@todo")
}

func (e *Edge) TimeAttr(key rune) time.Time {
	panic("@todo")
}
