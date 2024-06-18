package sql

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

const (
	Operator ExpressionType = "operator"
	Operand  ExpressionType = "operand"
)

type Expression struct {
	Type     ExpressionType
	Operator OperatorType

	ValueType string
	StrValue  string
	GoValue   any

	Left  *Expression
	Right *Expression
}
