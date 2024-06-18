package eval

import "errors"

func (expr *Expression) evaluateAnd(values map[string]any) (*EvalResult, error) {
	left, err := Evaluate(expr.Left, values)
	if err != nil {
		return nil, err
	}

	l, ok := left.GoValue.(bool)
	if !ok {
		return nil, errors.New("both sides of an logical operation must be boolean values")
	}

	if !l {
		return &EvalResult{GoValue: false}, nil
	}

	right, err := Evaluate(expr.Right, values)
	if err != nil {
		return nil, err
	}

	r, ok := right.GoValue.(bool)
	if !ok {
		return nil, errors.New("both sides of an logical operation must be boolean values")
	}

	return &EvalResult{GoValue: l && r}, nil
}

func (expr *Expression) evaluateOr(values map[string]any) (*EvalResult, error) {
	left, err := Evaluate(expr.Left, values)
	if err != nil {
		return nil, err
	}

	l, ok := left.GoValue.(bool)
	if !ok {
		return nil, errors.New("both sides of an logical operation must be boolean values")
	}

	if l {
		return &EvalResult{GoValue: true}, nil
	}

	right, err := Evaluate(expr.Right, values)
	if err != nil {
		return nil, err
	}

	r, ok := right.GoValue.(bool)
	if !ok {
		return nil, errors.New("both sides of an logical operation must be boolean values")
	}

	return &EvalResult{GoValue: l || r}, nil
}

func (expr *Expression) evaluateEqual(values map[string]any) (*EvalResult, error) {
	left, err := Evaluate(expr.Left, values)
	if err != nil {
		return nil, err
	}

	right, err := Evaluate(expr.Right, values)
	if err != nil {
		return nil, err
	}

	return &EvalResult{GoValue: left.GoValue == right.GoValue}, nil
}

func (expr *Expression) evaluateNotEqual(values map[string]any) (*EvalResult, error) {
	left, err := Evaluate(expr.Left, values)
	if err != nil {
		return nil, err
	}

	right, err := Evaluate(expr.Right, values)
	if err != nil {
		return nil, err
	}

	return &EvalResult{GoValue: left.GoValue != right.GoValue}, nil
}

func (expr *Expression) evaluateGreater(values map[string]any) (*EvalResult, error) {
	left, err := Evaluate(expr.Left, values)
	if err != nil {
		return nil, err
	}

	right, err := Evaluate(expr.Right, values)
	if err != nil {
		return nil, err
	}

	if !(left.genericValueType() == "number" && left.genericValueType() == right.genericValueType()) {
		return nil, errors.New("both sides of a comparison operation must be numbers")
	}

	return &EvalResult{GoValue: greaterThan(left.GoValue, right.GoValue)}, nil
}

func (expr *Expression) evaluateGreaterEqual(values map[string]any) (*EvalResult, error) {
	left, err := Evaluate(expr.Left, values)
	if err != nil {
		return nil, err
	}

	right, err := Evaluate(expr.Right, values)
	if err != nil {
		return nil, err
	}

	if !(left.genericValueType() == "number" && left.genericValueType() == right.genericValueType()) {
		return nil, errors.New("both sides of a comparison operation must be numbers")
	}

	return &EvalResult{GoValue: greaterThan(left.GoValue, right.GoValue)}, nil
}

func (expr *Expression) evaluateLess(values map[string]any) (*EvalResult, error) {
	left, err := Evaluate(expr.Left, values)
	if err != nil {
		return nil, err
	}

	right, err := Evaluate(expr.Right, values)
	if err != nil {
		return nil, err
	}

	if !(left.genericValueType() == "number" && left.genericValueType() == right.genericValueType()) {
		return nil, errors.New("both sides of a comparison operation must be numbers")
	}

	return &EvalResult{GoValue: greaterThan(right.GoValue, left.GoValue)}, nil
}

func (expr *Expression) evaluateLessEqual(values map[string]any) (*EvalResult, error) {
	left, err := Evaluate(expr.Left, values)
	if err != nil {
		return nil, err
	}

	right, err := Evaluate(expr.Right, values)
	if err != nil {
		return nil, err
	}

	if !(left.genericValueType() == "number" && left.genericValueType() == right.genericValueType()) {
		return nil, errors.New("both sides of a comparison operation must be numbers")
	}

	return &EvalResult{GoValue: greaterOrEqualThan(right.GoValue, left.GoValue)}, nil
}
