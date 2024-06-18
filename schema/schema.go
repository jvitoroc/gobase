package schema

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type Schema struct {
	mu     sync.Mutex
	tables []*Table

	rootDir string
}

func NewSchema(rootDir string) *Schema {
	return &Schema{rootDir: rootDir}
}

type NewColumn struct {
	Name string
	Type ColumnType
}

func (s *Schema) CreateTable(name string, columns []*NewColumn) (*Table, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.hasTable(name) {
		return nil, fmt.Errorf("table with name '%s' already exists", name)
	}

	c := make([]*Column, len(columns))
	for i := range columns {
		c[i] = &Column{
			ID:   uuid.New().ID(),
			Name: columns[i].Name,
			Type: columns[i].Type,
		}
	}

	t := &Table{
		ID:      uuid.New().ID(),
		Name:    name,
		Columns: c,

		rootDir: s.rootDir,
	}

	s.tables = append(s.tables, t)

	return t, nil
}

func (s *Schema) GetTable(name string) *Table {
	for _, t := range s.tables {
		if t.Name == name {
			return t
		}
	}

	return nil
}

func (s *Schema) hasTable(name string) bool {
	for _, t := range s.tables {
		if t.Name == name {
			return true
		}
	}

	return false
}
