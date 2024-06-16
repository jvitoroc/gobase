package main

import (
	"context"
	"errors"
	"fmt"
	"io"
)

type database struct {
	schema *Schema
}

func (d *database) run(batch string) (io.Reader, error) {
	p := &parser{}
	s, err := p.parse(batch)
	if err != nil {
		return nil, err
	}

	if len(s.Parts) == 0 {
		return nil, errors.New("empty batch")
	}

	switch s.Parts[0].Keyword {
	case "select":
		return d.selectStatement(s)
	}

	return nil, nil
}

func (d *database) selectStatement(s *statement) (io.Reader, error) {
	if err := validateSelectStatement(s); err != nil {
		return nil, err
	}

	tableName := ""
	var returningColumns []string
	// var postfixExpression []token

	for _, p := range s.Parts {
		switch p.Keyword {
		case "select":
			returningColumns, _ = p.Body.([]string)
		case "from":
			tableName, _ = p.Body.(string)
		case "where":
			// postfixExpression, _ = p.Body.([]token)
		}
	}

	t := d.schema.getTable(tableName)
	if t == nil {
		return nil, fmt.Errorf("table '%s' does not exist", tableName)
	}

	r, w := io.Pipe()

	err := t.read(context.Background(), w, returningColumns, func(row *deserializedRow) bool { return true })
	if err != nil {
		return nil, err
	}

	return r, nil
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
				goValue: c.value,
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
			b, ok := p.Body.([]string)
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
