package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRidiculousSelect(t *testing.T) {
	p := newParser(`SELeCT foo    , bar    FROM       jobs  where (foo == 
		 "  bbbbasdasd asd asd ") or (bar >= 1.0);`)
	q, err := p.parse()
	if err != nil {
		t.Error(err)
	}

	diff := cmp.Diff(q, []*statement{
		{
			Clauses: []*Clause{
				{
					Name: "select",
					Body: []*expression{
						{_type: "operand", valueType: "name", strValue: "foo"},
						{_type: "operand", valueType: "name", strValue: "bar"},
					},
				},
				{
					Name: "from",
					Body: "jobs",
				},
				{
					Name: "where",
					Body: &expression{
						_type:    "operator",
						operator: "or",
						left: &expression{
							_type:    "operator",
							operator: "equal",
							left:     &expression{_type: "operand", valueType: "name", strValue: "foo"},
							right: &expression{
								_type:     "operand",
								valueType: "string_literal",
								strValue:  "  bbbbasdasd asd asd ",
								goValue:   string("  bbbbasdasd asd asd "),
							},
						},
						right: &expression{
							_type:    "operator",
							operator: "greater_equal",
							left:     &expression{_type: "operand", valueType: "name", strValue: "bar"},
							right: &expression{
								_type:     "operand",
								valueType: "number_literal",
								strValue:  "1.0",
								goValue:   float64(1),
							},
						},
					},
				},
			},
		},
	}, cmp.AllowUnexported(token{}, expression{}))
	if diff != "" {
		t.Error(diff)
	}
}

func TestCreateTable(t *testing.T) {
	s, err := newParser(`
		CREATE TABLE foo DEFINITIONS (
			foo bool,
			bar int,
			baz string
		);
	`).parse()
	if err != nil {
		t.Error(err)
	}

	diff := cmp.Diff(s, []*statement{
		{
			Clauses: []*Clause{
				{
					Name: "create table",
					Body: "foo",
				},
				{
					Name: "definitions",
					Body: []*newColumn{
						{name: "foo", _type: BoolType},
						{name: "bar", _type: Int32Type},
						{name: "baz", _type: StringType},
					},
				},
			},
		},
	}, cmp.AllowUnexported(token{}, expression{}, newColumn{}))
	if diff != "" {
		t.Error(diff)
	}
}

func TestInsertInto(t *testing.T) {
	s, err := newParser(`
		INSERT INTO foo VALUES (true, 123, "foobarbaz");
	`).parse()
	if err != nil {
		t.Error(err)
	}

	diff := cmp.Diff(s, []*statement{
		{
			Clauses: []*Clause{
				{
					Name: "insert into",
					Body: "foo",
				},
				{
					Name: "values",
					Body: []string{
						"true",
						"123",
						"foobarbaz",
					},
				},
			},
		},
	}, cmp.AllowUnexported(token{}, expression{}))
	if diff != "" {
		t.Error(diff)
	}
}

func TestMultipleStatements(t *testing.T) {
	s, err := newParser(`
		CREATE TABLE foo DEFINITIONS (
			foo bool,
			bar int,
			baz string
		);

		INSERT INTO foo VALUES (true, 123, "foobarbaz");

		SELECT foo, bar, baz FROM foo WHERE foo != false AND bar > 100;
	`).parse()
	if err != nil {
		t.Error(err)
	}

	diff := cmp.Diff(s, []*statement{
		{
			Clauses: []*Clause{
				{
					Name: "create table",
					Body: "foo",
				},
				{
					Name: "definitions",
					Body: []*newColumn{
						{name: "foo", _type: BoolType},
						{name: "bar", _type: Int32Type},
						{name: "baz", _type: StringType},
					},
				},
			},
		},
		{
			Clauses: []*Clause{
				{
					Name: "insert into",
					Body: "foo",
				},
				{
					Name: "values",
					Body: []string{
						"true",
						"123",
						"foobarbaz",
					},
				},
			},
		},
		{
			Clauses: []*Clause{
				{
					Name: "select",
					Body: []*expression{
						{_type: Operand, valueType: "name", strValue: "foo"},
						{_type: Operand, valueType: "name", strValue: "bar"},
						{_type: Operand, valueType: "name", strValue: "baz"},
					},
				},
				{
					Name: "from",
					Body: "foo",
				},
				{
					Name: "where",
					Body: &expression{
						_type:    Operator,
						operator: "and",
						left: &expression{
							_type:    Operator,
							operator: "not_equal",
							left: &expression{
								_type:     Operand,
								valueType: "name",
								strValue:  "foo",
							},
							right: &expression{
								_type:     Operand,
								valueType: "boolean_literal",
								strValue:  "false",
								goValue:   false,
							},
						},
						right: &expression{
							_type:    Operator,
							operator: "greater",
							left: &expression{
								_type:     Operand,
								valueType: "name",
								strValue:  "bar",
							},
							right: &expression{
								_type:     Operand,
								valueType: "number_literal",
								strValue:  "100",
								goValue:   float64(100),
							},
						},
					},
				},
			},
		},
	}, cmp.AllowUnexported(token{}, expression{}, newColumn{}))
	if diff != "" {
		t.Error(diff)
	}
}

func Test_checkBooleanExpressionSyntax(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name    string
		args    args
		wantErr string
	}{
		{
			name: "empty",
			args: args{
				input: "",
			},
			wantErr: "empty expression",
		},
		{
			name: "invalid token",
			args: args{
				input: "(select",
			},
			wantErr: "'select' at 1:2 is not valid as part of an expression",
		},
		{
			name: "invalid beginning of expression",
			args: args{
				input: "and select",
			},
			wantErr: "can't start expression with operator 'and' at 1:1",
		},
		{
			name: "invalid token after operand",
			args: args{
				input: "asdasda 121",
			},
			wantErr: "expected operator after 'asdasda' at 1:9",
		},
		{
			name: "invalid token after operator",
			args: args{
				input: "asdasda == != asdasda",
			},
			wantErr: "expected operand after '==' at 1:12",
		},
		{
			name: "empty parentheses",
			args: args{
				input: "()",
			},
			wantErr: "empty parentheses at 1:2",
		},
		{
			name: "invalid token after left_parenthesis",
			args: args{
				input: "(and",
			},
			wantErr: "an operator is not allowed to be positioned at 1:2 after an opening parenthesis",
		},
		{
			name: "ending expression with operator",
			args: args{
				input: ") and",
			},
			wantErr: "can't end expression with an operator 'and' at 1:3",
		},
		{
			name: "happy path",
			args: args{
				input: "(foo == \"bbbbasdasd asd asd\") or (bar >= 1.0)",
			},
			wantErr: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := mustTokenize(tt.args.input)
			err := checkBooleanExpressionSyntax(input)
			gotErr := ""
			if err != nil {
				gotErr = err.Error()
			}
			if gotErr != tt.wantErr {
				t.Errorf("checkBooleanExpressionSyntax() error = %s, wantErr %s", gotErr, tt.wantErr)
			}
		})
	}
}

func Test_checkParenthesesBalance(t *testing.T) {
	type test struct {
		input string
		valid bool
	}
	tests := []test{
		{valid: true, input: "()"},
		{valid: true, input: "()()()"},
		{valid: false, input: "(("},
		{valid: false, input: "))"},
		{valid: false, input: "(((())"},
		{valid: true, input: "((((()))))"},
	}

	for i, tt := range tests {
		input := mustTokenize(tt.input)

		if err := checkParenthesesBalance(input); (err != nil) == tt.valid {
			t.Errorf("expected balance test %d to be successful", i+1)
		}
	}
}

func Test_infixToPostfix(t *testing.T) {
	type test struct {
		input    []token
		expected []token
	}
	tests := []test{
		{
			input: []token{
				{_type: "number_literal", strValue: "1"},
				{_type: "greater"},
				{_type: "number_literal", strValue: "2"},
			},
			expected: []token{
				{_type: "number_literal", strValue: "1"},
				{_type: "number_literal", strValue: "2"},
				{_type: "greater"},
			},
		},
		{
			input: []token{
				{_type: "number_literal", strValue: "1"},
				{_type: "greater_equal"},
				{_type: "number_literal", strValue: "2"},
				{_type: "or"},
				{_type: "left_parenthesis"},
				{_type: "number_literal", strValue: "3"},
				{_type: "greater"},
				{_type: "number_literal", strValue: "4"},
				{_type: "right_parenthesis"},
			},
			expected: []token{
				{_type: "number_literal", strValue: "1"},
				{_type: "number_literal", strValue: "2"},
				{_type: "greater_equal"},
				{_type: "number_literal", strValue: "3"},
				{_type: "number_literal", strValue: "4"},
				{_type: "greater"},
				{_type: "or"},
			},
		},
		{
			input: []token{
				{_type: "number_literal", strValue: "1"},
				{_type: "greater"},
				{_type: "number_literal", strValue: "2"},
				{_type: "or"},
				{_type: "number_literal", strValue: "3"},
				{_type: "less"},
				{_type: "number_literal", strValue: "4"},
			},
			expected: []token{
				{_type: "number_literal", strValue: "1"},
				{_type: "number_literal", strValue: "2"},
				{_type: "greater"},
				{_type: "number_literal", strValue: "3"},
				{_type: "number_literal", strValue: "4"},
				{_type: "less"},
				{_type: "or"},
			},
		},
		{
			input: []token{
				{_type: "left_parenthesis"},
				{_type: "number_literal", strValue: "1"},
				{_type: "greater"},
				{_type: "number_literal", strValue: "2"},
				{_type: "or"},
				{_type: "left_parenthesis"},
				{_type: "number_literal", strValue: "3"},
				{_type: "less"},
				{_type: "number_literal", strValue: "4"},
				{_type: "or"},
				{_type: "number_literal", strValue: "5"},
				{_type: "greater"},
				{_type: "number_literal", strValue: "6"},
				{_type: "right_parenthesis"},
				{_type: "and"},
				{_type: "left_parenthesis"},
				{_type: "number_literal", strValue: "7"},
				{_type: "greater"},
				{_type: "number_literal", strValue: "8"},
				{_type: "right_parenthesis"},
				{_type: "right_parenthesis"},
			},
			expected: []token{
				{_type: "number_literal", strValue: "1"},
				{_type: "number_literal", strValue: "2"},
				{_type: "greater"},
				{_type: "number_literal", strValue: "3"},
				{_type: "number_literal", strValue: "4"},
				{_type: "less"},
				{_type: "number_literal", strValue: "5"},
				{_type: "number_literal", strValue: "6"},
				{_type: "greater"},
				{_type: "or"},
				{_type: "number_literal", strValue: "7"},
				{_type: "number_literal", strValue: "8"},
				{_type: "greater"},
				{_type: "and"},
				{_type: "or"},
			},
		},
		{
			input: []token{
				{_type: "number_literal", strValue: "1"},
				{_type: "greater"},
				{_type: "number_literal", strValue: "2"},
				{_type: "or"},
				{_type: "number_literal", strValue: "3"},
				{_type: "less"},
				{_type: "number_literal", strValue: "4"},
				{_type: "or"},
				{_type: "number_literal", strValue: "5"},
				{_type: "greater"},
				{_type: "number_literal", strValue: "6"},
				{_type: "and"},
				{_type: "number_literal", strValue: "7"},
				{_type: "greater"},
				{_type: "number_literal", strValue: "8"},
			},
			expected: []token{
				{_type: "number_literal", strValue: "1"},
				{_type: "number_literal", strValue: "2"},
				{_type: "greater"},
				{_type: "number_literal", strValue: "3"},
				{_type: "number_literal", strValue: "4"},
				{_type: "less"},
				{_type: "or"},
				{_type: "number_literal", strValue: "5"},
				{_type: "number_literal", strValue: "6"},
				{_type: "greater"},
				{_type: "number_literal", strValue: "7"},
				{_type: "number_literal", strValue: "8"},
				{_type: "greater"},
				{_type: "and"},
				{_type: "or"},
			},
		},
		{
			input: []token{
				{_type: "left_parenthesis"},
				{_type: "number_literal", strValue: "1"},
				{_type: "greater"},
				{_type: "number_literal", strValue: "2"},
				{_type: "or"},
				{_type: "number_literal", strValue: "3"},
				{_type: "less"},
				{_type: "number_literal", strValue: "4"},
				{_type: "or"},
				{_type: "number_literal", strValue: "5"},
				{_type: "greater"},
				{_type: "number_literal", strValue: "6"},
				{_type: "right_parenthesis"},
				{_type: "and"},
				{_type: "number_literal", strValue: "7"},
				{_type: "greater"},
				{_type: "number_literal", strValue: "8"},
			},
			expected: []token{
				{_type: "number_literal", strValue: "1"},
				{_type: "number_literal", strValue: "2"},
				{_type: "greater"},
				{_type: "number_literal", strValue: "3"},
				{_type: "number_literal", strValue: "4"},
				{_type: "less"},
				{_type: "or"},
				{_type: "number_literal", strValue: "5"},
				{_type: "number_literal", strValue: "6"},
				{_type: "greater"},
				{_type: "or"},
				{_type: "number_literal", strValue: "7"},
				{_type: "number_literal", strValue: "8"},
				{_type: "greater"},
				{_type: "and"},
			},
		},
	}

	for i, tt := range tests {
		got := infixToPostfix(tt.input)
		if diff := cmp.Diff(got, tt.expected, cmp.AllowUnexported(token{})); diff != "" {
			t.Errorf("test %d failed: %s", i+1, diff)
		}
	}
}

func Test_infixToExpressionTree(t *testing.T) {
	type test struct {
		input    string
		expected *expression
	}
	tests := []test{
		{
			input: "1 > 2",
			expected: &expression{
				_type:    Operator,
				operator: "greater",
				left: &expression{
					_type:     Operand,
					valueType: "number_literal",
					goValue:   float64(1),
					strValue:  "1",
				},
				right: &expression{
					_type:     Operand,
					valueType: "number_literal",
					goValue:   float64(2),
					strValue:  "2",
				},
			},
		},
		{
			input: "a == b",
			expected: &expression{
				_type:    Operator,
				operator: "equal",
				left: &expression{
					_type:     Operand,
					valueType: "name",
					strValue:  "a",
				},
				right: &expression{
					_type:     Operand,
					valueType: "name",
					strValue:  "b",
				},
			},
		},
		{
			input: "(x > 0) and (y <= 10)",
			expected: &expression{
				_type:    Operator,
				operator: "and",
				left: &expression{
					_type:    Operator,
					operator: "greater",
					left: &expression{
						_type:     Operand,
						valueType: "name",
						strValue:  "x",
					},
					right: &expression{
						_type:     Operand,
						valueType: "number_literal",
						goValue:   float64(0),
						strValue:  "0",
					},
				},
				right: &expression{
					_type:    Operator,
					operator: "less_equal",
					left: &expression{
						_type:     Operand,
						valueType: "name",
						strValue:  "y",
					},
					right: &expression{
						_type:     Operand,
						valueType: "number_literal",
						goValue:   float64(10),
						strValue:  "10",
					},
				},
			},
		},
		{
			input: "(a > 0 and b <= 10) or (c == \"hello\" and d != 5)",
			expected: &expression{
				_type:    Operator,
				operator: "or",
				left: &expression{
					_type:    Operator,
					operator: "and",
					left: &expression{
						_type:    Operator,
						operator: "greater",
						left: &expression{
							_type:     Operand,
							valueType: "name",
							strValue:  "a",
						},
						right: &expression{
							_type:     Operand,
							valueType: "number_literal",
							goValue:   float64(0),
							strValue:  "0",
						},
					},
					right: &expression{
						_type:    Operator,
						operator: "less_equal",
						left: &expression{
							_type:     Operand,
							valueType: "name",
							strValue:  "b",
						},
						right: &expression{
							_type:     Operand,
							valueType: "number_literal",
							goValue:   float64(10),
							strValue:  "10",
						},
					},
				},
				right: &expression{
					_type:    Operator,
					operator: "and",
					left: &expression{
						_type:    Operator,
						operator: "equal",
						left: &expression{
							_type:     Operand,
							valueType: "name",
							strValue:  "c",
						},
						right: &expression{
							_type:     Operand,
							valueType: "string_literal",
							strValue:  "hello",
							goValue:   "hello",
						},
					},
					right: &expression{
						_type:    Operator,
						operator: "not_equal",
						left: &expression{
							_type:     Operand,
							valueType: "name",
							strValue:  "d",
						},
						right: &expression{
							_type:     Operand,
							valueType: "number_literal",
							goValue:   float64(5),
							strValue:  "5",
						},
					},
				},
			},
		},
	}

	for i, tt := range tests {
		input := mustTokenize(tt.input)
		got := infixToExpressionTree(input)
		if diff := cmp.Diff(got, tt.expected, cmp.AllowUnexported(expression{})); diff != "" {
			t.Errorf("test %d failed: %s", i+1, diff)
		}
	}
}

func Test_parser_mustTokenize(t *testing.T) {
	type test struct {
		input    string
		expected []token
	}
	tests := []test{
		{
			input: "SELECT , true FALSE \"string\" 123 123.321 () and OR == != > < >= <= foo ;",
			expected: []token{
				{_type: "clause", strValue: "select", line: 1, column: 1},
				{_type: "comma", strValue: ",", line: 1, column: 8},
				{_type: "boolean_literal", strValue: "true", goValue: true, line: 1, column: 10},
				{_type: "boolean_literal", strValue: "false", goValue: false, line: 1, column: 15},
				{_type: "string_literal", strValue: "string", goValue: "string", line: 1, column: 21},
				{_type: "number_literal", strValue: "123", goValue: float64(123), line: 1, column: 30},
				{_type: "number_literal", strValue: "123.321", goValue: float64(123.321), line: 1, column: 34},
				{_type: "left_parenthesis", strValue: "(", line: 1, column: 42},
				{_type: "right_parenthesis", strValue: ")", line: 1, column: 43},
				{_type: "and", strValue: "and", line: 1, column: 45},
				{_type: "or", strValue: "or", line: 1, column: 49},
				{_type: "equal", strValue: "==", line: 1, column: 52},
				{_type: "not_equal", strValue: "!=", line: 1, column: 55},
				{_type: "greater", strValue: ">", line: 1, column: 58},
				{_type: "less", strValue: "<", line: 1, column: 60},
				{_type: "greater_equal", strValue: ">=", line: 1, column: 62},
				{_type: "less_equal", strValue: "<=", line: 1, column: 65},
				{_type: "name", strValue: "foo", line: 1, column: 68},
				{_type: "end_of_statement", strValue: ";", line: 1, column: 72},
			},
		},
		{
			input: "; foo <= >= < > != == oR AnD () 123.321 123 \"string\" FaLse TrUe , sElEcT",
			expected: []token{
				{_type: "end_of_statement", strValue: ";", line: 1, column: 1},
				{_type: "name", strValue: "foo", line: 1, column: 3},
				{_type: "less_equal", strValue: "<=", line: 1, column: 7},
				{_type: "greater_equal", strValue: ">=", line: 1, column: 10},
				{_type: "less", strValue: "<", line: 1, column: 13},
				{_type: "greater", strValue: ">", line: 1, column: 15},
				{_type: "not_equal", strValue: "!=", line: 1, column: 17},
				{_type: "equal", strValue: "==", line: 1, column: 20},
				{_type: "or", strValue: "or", line: 1, column: 23},
				{_type: "and", strValue: "and", line: 1, column: 26},
				{_type: "left_parenthesis", strValue: "(", line: 1, column: 30},
				{_type: "right_parenthesis", strValue: ")", line: 1, column: 31},
				{_type: "number_literal", strValue: "123.321", goValue: float64(123.321), line: 1, column: 33},
				{_type: "number_literal", strValue: "123", goValue: float64(123), line: 1, column: 41},
				{_type: "string_literal", strValue: "string", goValue: "string", line: 1, column: 45},
				{_type: "boolean_literal", strValue: "false", goValue: false, line: 1, column: 54},
				{_type: "boolean_literal", strValue: "true", goValue: true, line: 1, column: 60},
				{_type: "comma", strValue: ",", line: 1, column: 65},
				{_type: "clause", strValue: "select", line: 1, column: 67},
			},
		},
	}

	for i, tt := range tests {
		got := mustTokenize(tt.input)
		if diff := cmp.Diff(got, tt.expected, cmp.AllowUnexported(token{})); diff != "" {
			t.Errorf("test %d failed: %s", i+1, diff)
		}
	}
}

func Test_parser_moveToNextToken(t *testing.T) {
	p := parser{t: &tokenizer{query: "; foo <= '"}}

	err := p.moveToNextToken()
	if err != nil {
		t.Error(err)
		return
	}

	if p.lookahead._type != "end_of_statement" {
		t.Errorf("expected ;, but got '%s'", p.lookahead._type)
		return
	}

	err = p.moveToNextToken()
	if err != nil {
		t.Error(err)
		return
	}

	if p.lookahead._type != "name" {
		t.Errorf("expected ;, but got '%s'", p.lookahead._type)
		return
	}

	err = p.moveToNextToken()
	if err != nil {
		t.Error(err)
		return
	}

	if p.lookahead._type != "less_equal" {
		t.Errorf("expected ;, but got '%s'", p.lookahead._type)
		return
	}

	err = p.moveToNextToken()
	if err == nil {
		t.Errorf("expected error, lookahead token is '%s'", p.lookahead._type)
		return
	}

	err = p.moveToNextToken()
	if err != nil {
		t.Error(err)
		return
	}

	if p.lookahead != tokenNoop {
		t.Errorf("expected noop, but got '%s'", p.lookahead._type)
	}
}

func Test_parser_consume(t *testing.T) {
	p := parser{t: &tokenizer{query: "; foo"}}

	err := p.moveToNextToken()
	if err != nil {
		t.Error(err)
		return
	}

	tk, err := p.consume()
	if err != nil {
		t.Error(err)
		return
	}

	if tk == p.lookahead {
		t.Errorf("last consumed token and lookahead should not be the same")
		return
	}

	if tk._type != "end_of_statement" {
		t.Errorf("expected consumed token to be ;, but got '%s'", tk._type)
		return
	}

	if p.lookahead._type != "name" {
		t.Errorf("expected lookahead token to be ;, but got '%s'", p.lookahead._type)
		return
	}

	tk, err = p.consume()
	if err != nil {
		t.Error(err)
		return
	}

	if tk == p.lookahead {
		t.Errorf("last consumed token and lookahead should not be the same")
		return
	}

	if tk._type != "name" {
		t.Errorf("expected consumed token to be name, but got '%s'", tk._type)
		return
	}

	if p.lookahead != tokenNoop {
		t.Errorf("expected lookahead token to be noop, but got '%s'", p.lookahead._type)
	}
}

func Test_parser_selectBody(t *testing.T) {
	type test struct {
		input       string
		expected    []*expression
		expectedErr string
	}
	tests := []test{
		{
			input: "foo, bar, 1 == 1, 2 <= 2, (1 != 2) and (true == false)",
			expected: []*expression{
				{
					_type:     "operand",
					valueType: "name",
					strValue:  "foo",
				},
				{
					_type:     "operand",
					valueType: "name",
					strValue:  "bar",
				},
				{
					_type:    "operator",
					operator: "equal",
					left: &expression{
						_type:     "operand",
						valueType: "number_literal",
						strValue:  "1",
						goValue:   float64(1),
					},
					right: &expression{
						_type:     "operand",
						valueType: "number_literal",
						strValue:  "1",
						goValue:   float64(1),
					},
				},
				{
					_type:    "operator",
					operator: "less_equal",
					left: &expression{
						_type:     "operand",
						valueType: "number_literal",
						strValue:  "2",
						goValue:   float64(2),
					},
					right: &expression{
						_type:     "operand",
						valueType: "number_literal",
						strValue:  "2",
						goValue:   float64(2),
					},
				},
				{
					_type:    "operator",
					operator: "and",
					left: &expression{
						_type:    "operator",
						operator: "not_equal",
						left: &expression{
							_type:     "operand",
							valueType: "number_literal",
							strValue:  "1",
							goValue:   float64(1),
						},
						right: &expression{
							_type:     "operand",
							valueType: "number_literal",
							strValue:  "2",
							goValue:   float64(2),
						},
					},
					right: &expression{
						_type:    "operator",
						operator: "equal",
						left: &expression{
							_type:     "operand",
							valueType: "boolean_literal",
							strValue:  "true",
							goValue:   true,
						},
						right: &expression{
							_type:     "operand",
							valueType: "boolean_literal",
							strValue:  "false",
							goValue:   false,
						},
					},
				},
			},
		},
		{
			input: "foo",
			expected: []*expression{
				{
					_type:     "operand",
					valueType: "name",
					strValue:  "foo",
				},
			},
		},
		{
			input: "1 == 1",
			expected: []*expression{
				{
					_type:    "operator",
					operator: "equal",
					left: &expression{
						_type:     "operand",
						valueType: "number_literal",
						strValue:  "1",
						goValue:   float64(1),
					},
					right: &expression{
						_type:     "operand",
						valueType: "number_literal",
						strValue:  "1",
						goValue:   float64(1),
					},
				},
			},
		},
		{
			input:       "1 == 1,",
			expectedErr: "unexpected comma at 1:7",
		},
		{
			input:       "",
			expectedErr: "invalid expression at 1:1",
		},
		{
			input:       "select",
			expectedErr: "invalid expression at 1:1",
		},
		{
			input:       " , , ",
			expectedErr: "invalid expression at 1:2",
		},
	}

	for i, tt := range tests {
		p := newParser(tt.input)

		got, err := p.selectBody()
		gotErr := ""
		if err != nil {
			gotErr = err.Error()
		}

		if tt.expectedErr != "" {
			if tt.expectedErr != gotErr {
				t.Errorf("test %d failed: expected err '%s', but got '%s'", i+1, tt.expectedErr, err.Error())
			}
			continue
		}

		if diff := cmp.Diff(got, tt.expected, cmp.AllowUnexported(token{}, expression{})); diff != "" {
			t.Errorf("test %d failed: %s", i+1, diff)
		}
	}
}

func Test_parser_definitionsBody(t *testing.T) {
	type test struct {
		input       string
		expected    []*newColumn
		expectedErr string
	}
	tests := []test{
		{
			input: "(foo int, bar string, baz bool)",
			expected: []*newColumn{
				{name: "foo", _type: Int32Type},
				{name: "bar", _type: StringType},
				{name: "baz", _type: BoolType},
			},
		},
		{
			input:       "()",
			expectedErr: "definitions cannot be empty near 1:2",
		},
		{
			input:       "(",
			expectedErr: "expected column name, but got '' at 1:2",
		},
		{
			input:       ")",
			expectedErr: "expected opening parenthesis, but got ')' at 1:1",
		},
		{
			input:       "bool)",
			expectedErr: "expected opening parenthesis, but got 'bool' at 1:1",
		},
		{
			input:       "(foo",
			expectedErr: "expected column type, but got '' at 1:5",
		},
		{
			input:       "(foo int",
			expectedErr: "expected comma, but got '' at 1:9",
		},
		{
			input: "(foo int)",
			expected: []*newColumn{
				{name: "foo", _type: Int32Type},
			},
		},
	}

	for i, tt := range tests {
		p := newParser(tt.input)

		got, err := p.definitionsBody()
		gotErr := ""
		if err != nil {
			gotErr = err.Error()
		}

		if tt.expectedErr != "" {
			if tt.expectedErr != gotErr {
				t.Errorf("test %d failed: expected err '%s', but got '%s'", i+1, tt.expectedErr, gotErr)
			}
			continue
		}

		if diff := cmp.Diff(got, tt.expected, cmp.AllowUnexported(newColumn{})); diff != "" {
			t.Errorf("test %d failed: %s", i+1, diff)
		}
	}
}

func Test_parser_name(t *testing.T) {
	type test struct {
		input       string
		expected    string
		expectedErr string
	}
	tests := []test{
		{
			input:    "foo",
			expected: "foo",
		},
		{
			input:       "123",
			expectedErr: "expected name, but got 'number_literal' at 1:1",
		},
		{
			input:       "true",
			expectedErr: "expected name, but got 'boolean_literal' at 1:1",
		},
		{
			input:       "bool",
			expectedErr: "expected name, but got 'data_type' at 1:1",
		},
	}

	for i, tt := range tests {
		p := newParser(tt.input)

		got, err := p.name()
		gotErr := ""
		if err != nil {
			gotErr = err.Error()
		}

		if tt.expectedErr != "" {
			if tt.expectedErr != gotErr {
				t.Errorf("test %d failed: expected err '%s', but got '%s'", i+1, tt.expectedErr, gotErr)
			}
			continue
		}

		if diff := cmp.Diff(got, tt.expected, cmp.AllowUnexported(newColumn{})); diff != "" {
			t.Errorf("test %d failed: %s", i+1, diff)
		}
	}
}

func Test_parser_valuesBody(t *testing.T) {
	type test struct {
		input       string
		expected    []string
		expectedErr string
	}
	tests := []test{
		{
			input:    `("foo", 123, 123.321, true, false)`,
			expected: []string{"foo", "123", "123.321", "true", "false"},
		},
		{
			input:    `("foo")`,
			expected: []string{"foo"},
		},
		{
			input:       `()`,
			expectedErr: "must provide values at 1:2",
		},
		{
			input:       `("foo"`,
			expectedErr: "expected comma, but got '' at 1:7",
		},
		{
			input:       `"foo"`,
			expectedErr: "expected opening parenthesis, but got 'foo' at 1:1",
		},
		{
			input:       `"foo")`,
			expectedErr: "expected opening parenthesis, but got 'foo' at 1:1",
		},
		{
			input:       `(`,
			expectedErr: "expected literal, but got ''",
		},
		{
			input:       `)`,
			expectedErr: "expected opening parenthesis, but got ')' at 1:1",
		},
	}

	for i, tt := range tests {
		p := newParser(tt.input)

		got, err := p.valuesBody()
		gotErr := ""
		if err != nil {
			gotErr = err.Error()
		}

		if gotErr != "" {
			if tt.expectedErr != gotErr {
				t.Errorf("test %d failed: expected err '%s', but got '%s'", i+1, tt.expectedErr, gotErr)
			}
			continue
		}

		if diff := cmp.Diff(got, tt.expected, cmp.AllowUnexported(newColumn{})); diff != "" {
			t.Errorf("test %d failed: %s", i+1, diff)
		}
	}
}

func Test_parser_whereBody(t *testing.T) {
	type test struct {
		input       string
		expected    *expression
		expectedErr string
	}
	tests := []test{
		{
			input: "a == 1",
			expected: &expression{
				_type:    Operator,
				operator: "equal",
				left: &expression{
					_type:     Operand,
					valueType: "name",
					strValue:  "a",
				},
				right: &expression{
					_type:     Operand,
					valueType: "number_literal",
					strValue:  "1",
					goValue:   float64(1),
				},
			},
		},
		{
			input:       "",
			expectedErr: "expected predicate after 'WHERE', but got nothing",
		},
		{
			input:       ")",
			expectedErr: "unexpected closing parenthesis at 1:1",
		},
	}

	for i, tt := range tests {
		p := newParser(tt.input)

		got, err := p.whereBody()
		gotErr := ""
		if err != nil {
			gotErr = err.Error()
		}

		if gotErr != "" {
			if tt.expectedErr != gotErr {
				t.Errorf("test %d failed: expected err '%s', but got '%s'", i+1, tt.expectedErr, gotErr)
			}
			continue
		}

		if diff := cmp.Diff(got, tt.expected, cmp.AllowUnexported(expression{})); diff != "" {
			t.Errorf("test %d failed: %s", i+1, diff)
		}
	}
}
