package main

// type operator string
type expressionType string

// const (
// 	And              operator = "and"
// 	Or               operator = "or"
// 	Equal            operator = "equal"
// 	NotEqual         operator = "not_equal"
// 	GreaterEqualThan operator = "greater_or_equal_than"
// 	GreaterThan      operator = "greater_than"
// 	LessEqualThan    operator = "less_or_equal_than"
// 	LessThan         operator = "less_than"
// )

const (
	Operator expressionType = "operator"
	Operand  expressionType = "operand"
)

type expression struct {
	_type    expressionType
	operator string

	valueType string
	strValue  string
	goValue   any

	left  *expression
	right *expression
}
