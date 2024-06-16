package main

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type Schema struct {
	mu     sync.Mutex
	Tables []*Table

	rootDir string
}

type newColumn struct {
	name  string
	_type columnType
}

func (s *Schema) createTable(name string, columns []*newColumn) (*Table, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.hasTable(name) {
		return nil, fmt.Errorf("table with name '%s' already exists", name)
	}

	c := make([]*Column, len(columns))
	for i := range columns {
		c[i] = &Column{
			ID:   uuid.New().ID(),
			Name: columns[i].name,
			Type: columns[i]._type,
		}
	}

	t := &Table{
		ID:      uuid.New().ID(),
		Name:    name,
		Columns: c,

		rootDir: s.rootDir,
	}

	s.Tables = append(s.Tables, t)

	return t, nil
}

func (s *Schema) getTable(name string) *Table {
	for _, t := range s.Tables {
		if t.Name == name {
			return t
		}
	}

	return nil
}

func (s *Schema) hasTable(name string) bool {
	for _, t := range s.Tables {
		if t.Name == name {
			return true
		}
	}

	return false
}
