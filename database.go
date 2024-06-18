package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/jvitoroc/gobase/eval"
	"github.com/jvitoroc/gobase/schema"
	"github.com/jvitoroc/gobase/sql"
)

type database struct {
	schema *schema.Schema
}

func (d *database) initialize(rootDir string) error {
	file, err := os.OpenFile(path.Join(rootDir, "schema"), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	sch := schema.NewSchema(rootDir)

	decoder := json.NewDecoder(file)
	err = decoder.Decode(sch)
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	d.schema = sch

	return nil
}

func (d *database) run(r io.Writer, batch string) error {
	p := sql.NewParser(batch)
	sts, err := p.Parse()
	if err != nil {
		return err
	}

	if len(sts) == 0 {
		return errors.New("empty batch")
	}

	for i, s := range sts {
		if len(s.Clauses) == 0 {
			return fmt.Errorf("empty Statement #%d", i+1)
		}

		switch s.Clauses[0].Type {
		case sql.CreateTable:
			if err := d.createTableStatement(r, s); err != nil {
				return err
			}
		case sql.Select:
			if err := d.selectStatement(r, s); err != nil {
				return err
			}
		case sql.InsertInto:
			if err := d.InsertIntoStatement(r, s); err != nil {
				return err
			}
		default:
			return fmt.Errorf("invalid Statement #%d", i+1)
		}
	}

	return nil
}

func (d *database) createTableStatement(r io.Writer, s *sql.Statement) error {
	if err := validateCreateTableStatement(s); err != nil {
		return err
	}

	tableName := ""
	var columns []*schema.NewColumn

	for _, p := range s.Clauses {
		switch p.Type {
		case sql.CreateTable:
			tableName = p.Body.(string)
		case sql.Definitions:
			columns = p.Body.([]*schema.NewColumn)
		}
	}

	_, err := d.schema.CreateTable(tableName, columns)
	if err != nil {
		return err
	}

	return nil
}

func (d *database) InsertIntoStatement(r io.Writer, s *sql.Statement) error {
	if err := validateInsertIntoStatement(s); err != nil {
		return err
	}

	tableName := ""
	var values []string

	for _, p := range s.Clauses {
		switch p.Type {
		case sql.InsertInto:
			tableName = p.Body.(string)
		case sql.Values:
			values = p.Body.([]string)
		}
	}

	t := d.schema.GetTable(tableName)
	if t == nil {
		return fmt.Errorf("table with name '%s' does not exist", tableName)
	}

	err := t.Insert(values)
	if err != nil {
		return err
	}

	return nil
}

func (d *database) selectStatement(r io.Writer, s *sql.Statement) error {
	if err := validateSelectStatement(s); err != nil {
		return err
	}

	tableName := ""
	var returningColumns []string
	var filter *eval.Expression

	for _, p := range s.Clauses {
		switch p.Type {
		case sql.Select:
			returningColumns, _ = p.Body.([]string)
		case sql.From:
			tableName, _ = p.Body.(string)
		case sql.Where:
			filter, _ = p.Body.(*eval.Expression)
		}
	}

	t := d.schema.GetTable(tableName)
	if t == nil {
		return fmt.Errorf("table with name '%s' does not exist", tableName)
	}

	err := t.Read(context.Background(), r, returningColumns, func(row *schema.DeserializedRow) (bool, error) {
		r, err := eval.Evaluate(filter, row.Map())
		if err != nil {
			return false, err
		}

		if res, ok := r.GoValue.(bool); ok {
			if ok {
				return res, nil
			}
		}

		return false, errors.New("WHERE clause is invalid, must result in a boolean result")
	})
	if err != nil {
		return err
	}

	return nil
}

func validateSelectStatement(s *sql.Statement) error {
	hasSelect := false
	hasFrom := false

	for _, p := range s.Clauses {
		switch p.Type {
		case sql.Select:
			b, ok := p.Body.([]*eval.Expression)
			if !ok {
				return errors.New("invalid type for SELECT body")
			}

			if len(b) == 0 {
				return errors.New("must provide columns after keyword SELECT")
			}

			hasSelect = true
		case sql.From:
			b, ok := p.Body.(string)
			if !ok {
				return errors.New("invalid type for FROM body")
			}

			if len(b) == 0 {
				return errors.New("must provide table name to be read after keyword FROM")
			}

			hasFrom = true
		case sql.Where:
			b, ok := p.Body.(*eval.Expression)
			if !ok {
				return errors.New("invalid type for WHERE body")
			}

			if b == nil {
				return errors.New("must provide a filter Expression after keyword WHERE")
			}
		}
	}

	if !hasSelect {
		return errors.New("missing SELECT clause")
	}

	if !hasFrom {
		return errors.New("missing FROM clause")
	}

	return nil
}

func validateCreateTableStatement(s *sql.Statement) error {
	hasCreateTable := false
	hasDefinitions := false

	for _, p := range s.Clauses {
		switch p.Type {
		case sql.CreateTable:
			b, ok := p.Body.(string)
			if !ok {
				return errors.New("invalid name for table")
			}

			if len(b) == 0 {
				return errors.New("must provide name for table")
			}

			hasCreateTable = true
		case sql.Definitions:
			b, ok := p.Body.([]*schema.NewColumn)
			if !ok {
				return errors.New("invalid definitions")
			}

			if len(b) == 0 {
				return errors.New("must provide definitions for table")
			}

			hasDefinitions = true
		}
	}

	if !hasCreateTable {
		return errors.New("missing CREATE TABLE clause")
	}

	if !hasDefinitions {
		return errors.New("missing DEFINITIONS clause")
	}

	return nil
}

func validateInsertIntoStatement(s *sql.Statement) error {
	hasInsertInto := false
	hasValues := false

	for _, p := range s.Clauses {
		switch p.Type {
		case sql.InsertInto:
			b, ok := p.Body.(string)
			if !ok {
				return errors.New("invalid table name")
			}

			if len(b) == 0 {
				return errors.New("must provide table name for writing")
			}

			hasInsertInto = true
		case sql.Values:
			b, ok := p.Body.([]string)
			if !ok {
				return errors.New("invalid values")
			}

			if len(b) == 0 {
				return errors.New("must provide values to be Insert into table")
			}

			hasValues = true
		}
	}

	if !hasInsertInto {
		return errors.New("missing INSERT INTO clause")
	}

	if !hasValues {
		return errors.New("missing VALUES clause")
	}

	return nil
}
