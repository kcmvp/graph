package graph

import (
	"fmt"
	"github.com/samber/lo"
	"github.com/tidwall/buntdb"
	"os"
	"os/user"
	"path/filepath"
	"sync"
)

const schemaPrefix = "_schema:"

// var _ Graph = (*store)(nil)
var db *buntdb.DB

func init() {
	var err error
	var u *user.User
	u, err = user.Current()
	if err != nil {
		panic(fmt.Errorf("can not get current user: %w", err))
	}
	dir := filepath.Join(u.HomeDir, "graph")
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		panic(fmt.Errorf("can not create directory: %w", err))
	}
	db, err = buntdb.Open(filepath.Join(dir, "db.data"))
	if err != nil {
		panic(fmt.Errorf("can not open database: %w", err))
	}
}

type store lo.Tuple3[*buntdb.DB, *Traits, sync.Map]

func (s *store) Traits() *Traits {
	//TODO implement me
	panic("implement me")
}

func (s *store) AddVertex(v Vertex) error {
	//TODO implement me
	panic("implement me")
}

func (s *store) Vertex(hash string) (Vertex, error) {
	//TODO implement me
	panic("implement me")
}

func (s *store) RemoveVertex(hash string) error {
	//TODO implement me
	panic("implement me")
}

func (s *store) UpdateVertex(hash string, attributes map[string]any) error {
	//TODO implement me
	panic("implement me")
}

func (s *store) AddEdge(e Edge) error {
	//TODO implement me
	panic("implement me")
}

func (s *store) Edge(hash string) (Edge, error) {
	//TODO implement me
	panic("implement me")
}

func (s *store) UpdateEdge(hash string, attributes map[string]any) error {
	//TODO implement me
	panic("implement me")
}

func (s *store) RemoveEdge(hash string) error {
	//TODO implement me
	panic("implement me")
}

func (s *store) Vertices() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *store) Edges() (int, error) {
	//TODO implement me
	panic("implement me")
}
