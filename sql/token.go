package sql

import (
	"slices"
	"strconv"
)

type token struct {
	_type    tokenType
	strValue string
	goValue  any

	line   int
	column int
}

var tokenNoop token

func (tk *token) isParenthesis() bool {
	return tk.isLeftParenthesis() || tk.isRightParenthesis()
}

func (tk *token) isLeftParenthesis() bool {
	return tk._type == leftParenthesis
}

func (tk *token) isRightParenthesis() bool {
	return tk._type == rightParenthesis
}

var (
	logicalOperators    = []tokenType{and, or}
	comparisonOperators = []tokenType{equal, notEqual, greaterEqual, greater, less, lessEqual}
	operands            = []tokenType{identifier, numberLiteral, stringLiteral, booleanLiteral}
)

var precedence = map[tokenType]int{
	equal:        1,
	notEqual:     1,
	greaterEqual: 1,
	greater:      1,
	less:         1,
	lessEqual:    1,
	and:          2,
	or:           3,
}

func (tk *token) hasLowerOrSamePrecedenceThan(tk1 token) bool {
	l, lok := precedence[tk._type]
	r, rok := precedence[tk1._type]

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

var literalTypes = []tokenType{numberLiteral, stringLiteral, booleanLiteral}

func (tk *token) isLiteral() bool {
	return slices.Contains(literalTypes, tk._type)
}

func (tk *token) convertToGoType() (v any, err error) {
	switch tk._type {
	case numberLiteral:
		v, err = strconv.ParseFloat(tk.strValue, 64)
	case booleanLiteral:
		v, err = strconv.ParseBool(tk.strValue)
	case stringLiteral:
		v = tk.strValue
	}

	return
}
