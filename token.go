package main

import (
	"slices"
)

type token struct {
	_type    string
	valueStr string
	valueGo  any

	line   int
	column int
}

var tokenNoop token

func (tk *token) isParenthesis() bool {
	return tk.isLeftParenthesis() || tk.isRightParenthesis()
}

func (tk *token) isLeftParenthesis() bool {
	return tk._type == "left_parenthesis"
}

func (tk *token) isRightParenthesis() bool {
	return tk._type == "right_parenthesis"
}

var (
	logicalOperators    = []string{"and", "or"}
	comparisonOperators = []string{"equal", "not_equal", "greater_equal", "greater", "less", "less_equal"}
	operands            = []string{"name", "decimal_literal", "integer_literal", "string_literal", "boolean_literal"}
)

var precedence = map[string]int{
	"comparison": 1,
	"and":        2,
	"or":         3,
}

func (tk *token) precedenceCategory() string {
	if tk.isComparisonOperator() {
		return "comparison"
	}

	return tk._type
}

func (tk *token) hasLowerOrSamePrecedenceThan(tk1 token) bool {
	l, lok := precedence[tk.precedenceCategory()]
	r, rok := precedence[tk1.precedenceCategory()]

	if !lok || !rok {
		return false
	}

	return l >= r
}

func (tk *token) isPredicateToken() bool {
	return tk.isOperator() || tk.isOperand() || tk.isParenthesis()
}

func (tk *token) isLogicalOperator() bool {
	return slices.Contains(logicalOperators, tk._type)
}

func (tk *token) isComparisonOperator() bool {
	return slices.Contains(comparisonOperators, tk._type)
}

func (tk *token) isOperand() bool {
	return slices.Contains(operands, tk._type)
}

func (tk *token) isOperator() bool {
	return tk.isComparisonOperator() || tk.isLogicalOperator()
}
