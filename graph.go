package graph

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const DefaultVertexRune = '_'
const DefaultEdgeRune = '#'

var (
	ErrVertexNotFound      = errors.New("vertex not found")
	ErrVertexAlreadyExists = errors.New("vertex already exists")
	ErrEdgeNotFound        = errors.New("edge not found")
	ErrEdgeAlreadyExists   = errors.New("edge already exists")
	ErrEdgeCreatesCycle    = errors.New("edge would create a cycle")
	ErrVertexHasEdges      = errors.New("vertex has edges")
)

// schema returns the schema of the graph, includes the vertex and edge types.
// and those data all save in the store as key-value pairs in the db
// vertex schema pairs naming standard is "v_<vertex_type>_<attribute_name>:<short_name>"
// edge schema pairs naming standard is "e_<vertex_type>_<vertex_type>_<attribute_name>:<short_name>"
func schema() map[string]string {
	panic("@todo")
}

type Vertex interface {
	// ID returns the unique hash of the vertex. This hash is used to identify
	// the vertex in the graph. It must not contain Graph.EdgeSeparator().
	ID() string
	// VertexType returns the type of the vertex. This is used to determine the
	// type of the vertex when creating a new vertex.
	// This is the type level method, which means that all vertices of the same type must have the same vertex type.
	// e.g. for a vertex type of "user", this method returns 'u'.
	VertexType() rune
}

// hash returns the hash of the vertex, which is a combination of the vertex type and the ID.
// Vertex hashes are always prefixed with Vertex type rune and an underscore. so make sure there is no '#' in the ID,
// otherwise the hash will be invalid.
func hash(v Vertex, r rune) string {
	return fmt.Sprintf("v%s%s%s", v.VertexType(), r, v.ID())
}

type Vertexes[T Vertex] struct {
	data   []T
	schema map[string]string
}

type Edge map[string]any

func newEdge(sh, th string, attrs map[string]any) Edge {
	if len(strings.TrimSpace(sh)) == 0 || len(strings.TrimSpace(th)) == 0 {
		panic("source and target vertex hashes must not be empty")
	}
	if attrs == nil {
		attrs = make(map[string]any)
	}
	attrs["_s"] = sh
	attrs["_t"] = sh
	attrs["_b"] = 0
	return attrs
}

// Source returns the hash of the source vertex of the edge.
func (e Edge) Source() string {
	return e["_s"].(string)
}

func (e Edge) Hash() string {
	return e["_k"].(string)
}

// Target returns the target vertex hash of the edge.
func (e Edge) Target() string {
	return e["_t"].(string)
}

// Bidirectional indicate this edge is bidirectional or not.
func (e Edge) Bidirectional() bool {
	if v, ok := e["_b"]; ok {
		if b, ok := v.(int64); ok {
			return b == 1
		}
	}
	return false
}

type Graph interface {
	// Traits returns the graph's traits. Those traits must be set when creating
	// a graph using New.
	Traits() *Traits

	// AddVertex creates a new vertex in the graph. If the vertex already exists
	// in the graph, ErrVertexAlreadyExists will be returned. a valid hash key of T should
	// not contain EdgeSeparator() and must be unique.
	AddVertex(v Vertex) (string, error)

	// Vertex returns the vertex with the given hash or ErrVertexNotFound if it
	// doesn't exist.
	Vertex(hash string) (Vertex, error)

	// RemoveVertex removes the vertex with the given hash value from the graph.
	// The vertex is not allowed to have edges and thus must be disconnected.
	// Potential edges must be removed first. Otherwise, ErrVertexHasEdges will
	// be returned. If the vertex doesn't exist, ErrVertexNotFound is returned.
	RemoveVertex(hash string) error

	UpdateVertex(hash string, attributes map[string]any) error

	// AddEdge creates an edge between the sh and the th vertex.
	// If either vertex cannot be found, ErrVertexNotFound will be returned. If
	// the edge already exists, ErrEdgeAlreadyExists will be returned. If cycle
	// prevention has been activated using PreventCycles and if adding the edge
	// would create a cycle, ErrEdgeCreatesCycle will be returned.
	AddEdge(source, target string, attrs map[string]any) error

	// Edge returns the edge joining two given vertices or ErrEdgeNotFound if
	// the edge doesn't exist. In an undirected graph, an edge with swapped
	// sh and th vertices does match.
	Edge(hash string) (Edge, error)

	// UpdateEdge updates the edge joining the two given vertices with the data
	// overwrite the existing attributes using the EdgeAttributes option.
	UpdateEdge(hash string, attributes map[string]any) error

	// RemoveEdge removes the edge between the given sh and th vertices.
	// If the edge cannot be found, ErrEdgeNotFound will be returned.
	RemoveEdge(hash string) error

	// Vertices returns the number of vertices in the graph.
	Vertices() (int, error)

	// Edges returns the number of edges in the graph.
	Edges() (int, error)
	//
}

// New creates a new graph with vertices of type T, identified by hash values of
// type K. These hash values will be obtained using the provided hash function.
//
// The graph will use the default in-memory store for persisting vertices and
// edges. To use a different [Store], use [NewWithStore].
func New(options ...func(*Traits)) Graph {
	//return NewWithStore(hash, newMemoryStore[K, T](), options...)
	//return newStore(options...)
	panic("@todo")
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
//
//	func (e Edge) MarshalJSON() ([]byte, error) {
//		type Alias Edge
//		return json.Marshal(&struct {
//			Attributes map[string]any `json:"attributes"`
//			*Alias
//		}{
//			Attributes: e.attributes,
//			Alias:      (*Alias)(e),
//		})
//	}
//
//	func (e *Edge) UnmarshalJSON(data []byte) error {
//		type Alias Edge
//		aux := &struct {
//			Attributes map[string]any `json:"attributes"`
//			*Alias
//		}{
//			Alias: (*Alias)(e),
//		}
//		if err := json.Unmarshal(data, &aux); err != nil {
//			return err
//		}
//		e.attributes = aux.Attributes
//		return nil
//	}
//
//	func (e *Edge) Attribute(key rune, v any) {
//		e.attributes[string(key)] = v
//	}
func (e *Edge) StringAttr(key string) string {
	panic("@todo")
}

func (e *Edge) IntAttr(key string) int64 {
	panic("@todo")
}

func (e *Edge) UintAttr(key string) uint64 {
	panic("@todo")
}

func (e *Edge) FloatAttr(key string) float64 {
	panic("@todo")
}

func (e *Edge) BoolAttr(key string) bool {
	panic("@todo")
}

func (e *Edge) TimeAttr(key string) time.Time {
	panic("@todo")
}
