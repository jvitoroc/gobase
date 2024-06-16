package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
)

type database struct {
	schema *Schema
}

func (d *database) initialize(rootDir string) error {
	file, err := os.OpenFile(path.Join(rootDir, "schema"), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	sch := &Schema{rootDir: rootDir}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(sch)
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	d.schema = sch

	return nil
}

func (d *database) run(r io.Writer, batch string) error {
	p := &parser{}
	sts, err := p.parse(batch)
	if err != nil {
		return err
	}

	if len(sts) == 0 {
		return errors.New("empty batch")
	}

	for i, s := range sts {
		if len(s.Parts) == 0 {
			return fmt.Errorf("empty statement #%d", i+1)
		}

		switch s.Parts[0].Keyword {
		case "create table":
			return d.createTableStatement(r, s)
		case "select":
			return d.selectStatement(r, s)
		case "insert into":
			return d.insertIntoStatement(r, s)
		default:
			return fmt.Errorf("invalid statement #%d", i+1)
		}
	}

	return nil
}

func (d *database) createTableStatement(r io.Writer, s *statement) error {
	if err := validateCreateTableStatement(s); err != nil {
		return err
	}

	tableName := ""
	var columns []*newColumn

	for _, p := range s.Parts {
		switch p.Keyword {
		case "create table":
			tableName = p.Body.(string)
		case "definitions":
			columns = p.Body.([]*newColumn)
		}
	}

	_, err := d.schema.createTable(tableName, columns)
	if err != nil {
		return err
	}

	return nil
}

func (d *database) insertIntoStatement(r io.Writer, s *statement) error {
	if err := validateInsertIntoStatement(s); err != nil {
		return err
	}

	tableName := ""
	var values []string

	for _, p := range s.Parts {
		switch p.Keyword {
		case "insert into":
			tableName = p.Body.(string)
		case "values":
			values = p.Body.([]string)
		}
	}

	t := d.schema.getTable(tableName)
	if t == nil {
		return fmt.Errorf("table with name '%s' does not exist", tableName)
	}

	err := t.insert(values)
	if err != nil {
		return err
	}

	return nil
}

func (d *database) selectStatement(r io.Writer, s *statement) error {
	if err := validateSelectStatement(s); err != nil {
		return err
	}

	tableName := ""
	var returningColumns []string
	var filter *expression

	for _, p := range s.Parts {
		switch p.Keyword {
		case "select":
			returningColumns, _ = p.Body.([]string)
		case "from":
			tableName, _ = p.Body.(string)
		case "where":
			filter, _ = p.Body.(*expression)
		}
	}

	t := d.schema.getTable(tableName)
	if t == nil {
		return fmt.Errorf("table with name '%s' does not exist", tableName)
	}

	err := t.read(context.Background(), r, returningColumns, func(row *deserializedRow) (bool, error) {
		r, err := evaluateBooleanExpressionAgainstRow(row, filter)
		if err != nil {
			return false, err
		}

		if res, ok := r.goValue.(bool); ok {
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

type result struct {
	goValue any
}

func (r *result) genericValueType() string {
	switch r.goValue.(type) {
	case float64:
		return "number"
	case bool:
		return "bool"
	case string:
		return "string"
	default:
		return ""
	}
}

func evaluateBooleanExpressionAgainstRow(row *deserializedRow, expr *expression) (*result, error) {
	if expr._type == Operand {
		if expr.valueType == "name" {
			c := row.getColumn(expr.strValue)
			if c == nil {
				return nil, fmt.Errorf("column '%s' does not exist", expr.strValue)
			}

			return &result{
				goValue: c.Value,
			}, nil
		} else {
			return &result{
				goValue: expr.goValue,
			}, nil
		}
	}

	if expr._type == Operator {
		switch expr.operator {
		case "and":
			left, err := evaluateBooleanExpressionAgainstRow(row, expr.left)
			if err != nil {
				return nil, err
			}

			l, ok := left.goValue.(bool)
			if !ok {
				return nil, errors.New("both sides of an logical operation must be boolean values")
			}

			if !l {
				return &result{goValue: false}, nil
			}

			right, err := evaluateBooleanExpressionAgainstRow(row, expr.right)
			if err != nil {
				return nil, err
			}

			r, ok := right.goValue.(bool)
			if !ok {
				return nil, errors.New("both sides of an logical operation must be boolean values")
			}

			return &result{goValue: l && r}, nil
		case "or":
			left, err := evaluateBooleanExpressionAgainstRow(row, expr.left)
			if err != nil {
				return nil, err
			}

			l, ok := left.goValue.(bool)
			if !ok {
				return nil, errors.New("both sides of an logical operation must be boolean values")
			}

			if l {
				return &result{goValue: true}, nil
			}

			right, err := evaluateBooleanExpressionAgainstRow(row, expr.right)
			if err != nil {
				return nil, err
			}

			r, ok := right.goValue.(bool)
			if !ok {
				return nil, errors.New("both sides of an logical operation must be boolean values")
			}

			return &result{goValue: l || r}, nil
		case "equal":
			left, err := evaluateBooleanExpressionAgainstRow(row, expr.left)
			if err != nil {
				return nil, err
			}

			right, err := evaluateBooleanExpressionAgainstRow(row, expr.right)
			if err != nil {
				return nil, err
			}

			return &result{goValue: left.goValue == right.goValue}, nil
		case "not_equal":
			left, err := evaluateBooleanExpressionAgainstRow(row, expr.left)
			if err != nil {
				return nil, err
			}

			right, err := evaluateBooleanExpressionAgainstRow(row, expr.right)
			if err != nil {
				return nil, err
			}

			return &result{goValue: left.goValue != right.goValue}, nil
		case "greater":
			left, err := evaluateBooleanExpressionAgainstRow(row, expr.left)
			if err != nil {
				return nil, err
			}

			right, err := evaluateBooleanExpressionAgainstRow(row, expr.right)
			if err != nil {
				return nil, err
			}

			if !(left.genericValueType() == "number" && left.genericValueType() == right.genericValueType()) {
				return nil, errors.New("both sides of a comparison operation must be numbers")
			}

			return &result{goValue: greaterThan(left.goValue, right.goValue)}, nil
		case "greater_equal":
			left, err := evaluateBooleanExpressionAgainstRow(row, expr.left)
			if err != nil {
				return nil, err
			}

			right, err := evaluateBooleanExpressionAgainstRow(row, expr.right)
			if err != nil {
				return nil, err
			}

			if !(left.genericValueType() == "number" && left.genericValueType() == right.genericValueType()) {
				return nil, errors.New("both sides of a comparison operation must be numbers")
			}

			return &result{goValue: greaterOrEqualThan(left.goValue, right.goValue)}, nil
		case "less":
			left, err := evaluateBooleanExpressionAgainstRow(row, expr.left)
			if err != nil {
				return nil, err
			}

			right, err := evaluateBooleanExpressionAgainstRow(row, expr.right)
			if err != nil {
				return nil, err
			}

			if !(left.genericValueType() == "number" && left.genericValueType() == right.genericValueType()) {
				return nil, errors.New("both sides of a comparison operation must be numbers")
			}

			return &result{goValue: greaterThan(right.goValue, left.goValue)}, nil
		case "less_equal":
			left, err := evaluateBooleanExpressionAgainstRow(row, expr.left)
			if err != nil {
				return nil, err
			}

			right, err := evaluateBooleanExpressionAgainstRow(row, expr.right)
			if err != nil {
				return nil, err
			}

			if !(left.genericValueType() == "number" && left.genericValueType() == right.genericValueType()) {
				return nil, errors.New("both sides of a comparison operation must be numbers")
			}

			return &result{goValue: greaterOrEqualThan(right.goValue, left.goValue)}, nil
		}
	}

	return nil, errors.New("unknown expression type")
}

func greaterThan(left, right any) bool {
	l := left.(float64)
	r := right.(float64)

	return l > r
}

func greaterOrEqualThan(left, right any) bool {
	l := left.(float64)
	r := right.(float64)

	return l >= r
}

func validateSelectStatement(s *statement) error {
	hasSelect := false
	hasFrom := false

	for _, p := range s.Parts {
		switch p.Keyword {
		case "select":
			b, ok := p.Body.([]*expression)
			if !ok {
				return errors.New("invalid type for SELECT body")
			}

			if len(b) == 0 {
				return errors.New("must provide columns after keyword SELECT")
			}

			hasSelect = true
		case "from":
			b, ok := p.Body.(string)
			if !ok {
				return errors.New("invalid type for FROM body")
			}

			if len(b) == 0 {
				return errors.New("must provide table name to be read after keyword FROM")
			}

			hasFrom = true
		case "where":
			b, ok := p.Body.(*expression)
			if !ok {
				return errors.New("invalid type for WHERE body")
			}

			if b == nil {
				return errors.New("must provide a filter expression after keyword WHERE")
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

func validateCreateTableStatement(s *statement) error {
	hasCreateTable := false
	hasDefinitions := false

	for _, p := range s.Parts {
		switch p.Keyword {
		case "create table":
			b, ok := p.Body.(string)
			if !ok {
				return errors.New("invalid name for table")
			}

			if len(b) == 0 {
				return errors.New("must provide name for table")
			}

			hasCreateTable = true
		case "definitions":
			b, ok := p.Body.([]*newColumn)
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

func validateInsertIntoStatement(s *statement) error {
	hasInsertInto := false
	hasValues := false

	for _, p := range s.Parts {
		switch p.Keyword {
		case "insert into":
			b, ok := p.Body.(string)
			if !ok {
				return errors.New("invalid table name")
			}

			if len(b) == 0 {
				return errors.New("must provide table name for writing")
			}

			hasInsertInto = true
		case "values":
			b, ok := p.Body.([]string)
			if !ok {
				return errors.New("invalid values")
			}

			if len(b) == 0 {
				return errors.New("must provide values to be insert into table")
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
