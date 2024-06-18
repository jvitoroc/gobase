package eval

import (
	"errors"
	"fmt"
	"slices"
)

type OperatorType string
type ExpressionType string

const (
	And              OperatorType = "and"
	Or               OperatorType = "or"
	Equal            OperatorType = "equal"
	NotEqual         OperatorType = "not_equal"
	GreaterEqualThan OperatorType = "greater_equal"
	GreaterThan      OperatorType = "greater"
	LessEqualThan    OperatorType = "less_equal"
	LessThan         OperatorType = "less"
)

var operators = []OperatorType{And, Or, Equal, NotEqual, GreaterEqualThan, GreaterThan, LessEqualThan, LessThan}

func IsOperator(operator string) bool {
	return slices.Contains(operators, OperatorType(operator))
}

const (
	Operator ExpressionType = "operator"
	Operand  ExpressionType = "operand"
)

type Expression struct {
	Type     ExpressionType
	Operator OperatorType

	Identifier string
	GoValue    any

	Left  *Expression
	Right *Expression
}

type EvalResult struct {
	GoValue any
}

func (r *EvalResult) genericValueType() string {
	switch r.GoValue.(type) {
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

func (expr *Expression) Evaluate(values map[string]any) (*EvalResult, error) {
	return Evaluate(expr, values)
}

func Evaluate(expr *Expression, values map[string]any) (*EvalResult, error) {
	if expr.Type == Operand {
		if expr.Identifier != "" {
			v := values[expr.Identifier]
			if v == nil {
				return nil, fmt.Errorf("value '%s' does not exist", expr.Identifier)
			}

			return &EvalResult{
				GoValue: v,
			}, nil
		} else {
			return &EvalResult{
				GoValue: expr.GoValue,
			}, nil
		}
	}

	if expr.Type == Operator {
		switch expr.Operator {
		case And:
			return expr.evaluateAnd(values)
		case Or:
			return expr.evaluateOr(values)
		case Equal:
			return expr.evaluateEqual(values)
		case NotEqual:
			return expr.evaluateNotEqual(values)
		case GreaterThan:
			return expr.evaluateGreater(values)
		case GreaterEqualThan:
			return expr.evaluateGreaterEqual(values)
		case LessThan:
			return expr.evaluateLess(values)
		case LessEqualThan:
			return expr.evaluateLessEqual(values)
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
