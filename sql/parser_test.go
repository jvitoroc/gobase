package sql

import (
	"testing"

	"github.com/jvitoroc/gobase/eval"
	"github.com/jvitoroc/gobase/schema"

	"github.com/google/go-cmp/cmp"
)

func TestRidiculousSelect(t *testing.T) {
	p := NewParser(`SELeCT foo    , bar    FROM       jobs  where (foo == 
		 "  bbbbasdasd asd asd ") or (bar >= 1.0);`)
	q, err := p.Parse()
	if err != nil {
		t.Error(err)
	}

	diff := cmp.Diff(q, []*Statement{
		{
			Clauses: []*Clause{
				{
					Type: "select",
					Body: []*eval.Expression{
						{Type: "operand", Identifier: "foo"},
						{Type: "operand", Identifier: "bar"},
					},
				},
				{
					Type: "from",
					Body: "jobs",
				},
				{
					Type: "where",
					Body: &eval.Expression{
						Type:     "operator",
						Operator: "or",
						Left: &eval.Expression{
							Type:     "operator",
							Operator: "equal",
							Left:     &eval.Expression{Type: "operand", Identifier: "foo"},
							Right: &eval.Expression{
								Type:    "operand",
								GoValue: string("  bbbbasdasd asd asd "),
							},
						},
						Right: &eval.Expression{
							Type:     "operator",
							Operator: "greater_equal",
							Left:     &eval.Expression{Type: "operand", Identifier: "bar"},
							Right: &eval.Expression{
								Type:    "operand",
								GoValue: float64(1),
							},
						},
					},
				},
			},
		},
	}, cmp.AllowUnexported(token{}, eval.Expression{}))
	if diff != "" {
		t.Error(diff)
	}
}

func TestCreateTable(t *testing.T) {
	s, err := NewParser(`
		CREATE TABLE foo DEFINITIONS (
			foo bool,
			bar int,
			baz string
		);
	`).Parse()
	if err != nil {
		t.Error(err)
	}

	diff := cmp.Diff(s, []*Statement{
		{
			Clauses: []*Clause{
				{
					Type: "create table",
					Body: "foo",
				},
				{
					Type: "definitions",
					Body: []*schema.NewColumn{
						{Name: "foo", Type: schema.BoolType},
						{Name: "bar", Type: schema.Int32Type},
						{Name: "baz", Type: schema.StringType},
					},
				},
			},
		},
	}, cmp.AllowUnexported(token{}, eval.Expression{}, schema.NewColumn{}))
	if diff != "" {
		t.Error(diff)
	}
}

func TestInsertInto(t *testing.T) {
	s, err := NewParser(`
		INSERT INTO foo VALUES (true, 123, "foobarbaz");
	`).Parse()
	if err != nil {
		t.Error(err)
	}

	diff := cmp.Diff(s, []*Statement{
		{
			Clauses: []*Clause{
				{
					Type: "insert into",
					Body: "foo",
				},
				{
					Type: "values",
					Body: []string{
						"true",
						"123",
						"foobarbaz",
					},
				},
			},
		},
	}, cmp.AllowUnexported(token{}, eval.Expression{}))
	if diff != "" {
		t.Error(diff)
	}
}

func TestMultipleStatements(t *testing.T) {
	s, err := NewParser(`
		CREATE TABLE foo DEFINITIONS (
			foo bool,
			bar int,
			baz string
		);

		INSERT INTO foo VALUES (true, 123, "foobarbaz");

		SELECT foo, bar, baz FROM foo WHERE foo != false AND bar > 100;
	`).Parse()
	if err != nil {
		t.Error(err)
	}

	diff := cmp.Diff(s, []*Statement{
		{
			Clauses: []*Clause{
				{
					Type: "create table",
					Body: "foo",
				},
				{
					Type: "definitions",
					Body: []*schema.NewColumn{
						{Name: "foo", Type: schema.BoolType},
						{Name: "bar", Type: schema.Int32Type},
						{Name: "baz", Type: schema.StringType},
					},
				},
			},
		},
		{
			Clauses: []*Clause{
				{
					Type: "insert into",
					Body: "foo",
				},
				{
					Type: "values",
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
					Type: "select",
					Body: []*eval.Expression{
						{Type: eval.Operand, Identifier: "foo"},
						{Type: eval.Operand, Identifier: "bar"},
						{Type: eval.Operand, Identifier: "baz"},
					},
				},
				{
					Type: "from",
					Body: "foo",
				},
				{
					Type: "where",
					Body: &eval.Expression{
						Type:     eval.Operator,
						Operator: "and",
						Left: &eval.Expression{
							Type:     eval.Operator,
							Operator: "not_equal",
							Left: &eval.Expression{
								Type:       eval.Operand,
								Identifier: "foo",
							},
							Right: &eval.Expression{
								Type:    eval.Operand,
								GoValue: false,
							},
						},
						Right: &eval.Expression{
							Type:     eval.Operator,
							Operator: "greater",
							Left: &eval.Expression{
								Type:       eval.Operand,
								Identifier: "bar",
							},
							Right: &eval.Expression{
								Type:    eval.Operand,
								GoValue: float64(100),
							},
						},
					},
				},
			},
		},
	}, cmp.AllowUnexported(token{}, eval.Expression{}, schema.NewColumn{}))
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
			wantErr: "empty eval.Expression",
		},
		{
			name: "invalid token",
			args: args{
				input: "(select",
			},
			wantErr: "'select' at 1:2 is not valid as part of an eval.Expression",
		},
		{
			name: "invalid beginning of eval.Expression",
			args: args{
				input: "and select",
			},
			wantErr: "can't start eval.Expression with operator 'and' at 1:1",
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
			name: "ending eval.Expression with operator",
			args: args{
				input: ") and",
			},
			wantErr: "can't end eval.Expression with an operator 'and' at 1:3",
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
		input       []token
		expected    []token
		expectedErr string
	}
	tests := []test{
		{
			input:       nil,
			expectedErr: "no tokens given",
		},
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
		got, err := infixToPostfix(tt.input)
		gotErr := ""
		if err != nil {
			gotErr = err.Error()
		}

		if gotErr != "" {
			if tt.expectedErr != gotErr {
				t.Errorf("test %d failed: expected err '%s', but got '%s'", i+1, tt.expectedErr, err.Error())
			}
			continue
		}

		if diff := cmp.Diff(got, tt.expected, cmp.AllowUnexported(token{})); diff != "" {
			t.Errorf("test %d failed: %s", i+1, diff)
		}
	}
}

func Test_infixToExpressionTree(t *testing.T) {
	type test struct {
		input       string
		expected    *eval.Expression
		expectedErr string
	}
	tests := []test{
		{
			input:       "",
			expectedErr: "no tokens given",
		},
		{
			input:       "1 select 1",
			expectedErr: "token 'select' at 1:3 is invalid as part of an expression",
		},
		{
			input: "1 > 2",
			expected: &eval.Expression{
				Type:     eval.Operator,
				Operator: "greater",
				Left: &eval.Expression{
					Type:    eval.Operand,
					GoValue: float64(1),
				},
				Right: &eval.Expression{
					Type:    eval.Operand,
					GoValue: float64(2),
				},
			},
		},
		{
			input: "a == b",
			expected: &eval.Expression{
				Type:     eval.Operator,
				Operator: "equal",
				Left: &eval.Expression{
					Type: eval.Operand,

					Identifier: "a",
				},
				Right: &eval.Expression{
					Type: eval.Operand,

					Identifier: "b",
				},
			},
		},
		{
			input: "(x > 0) and (y <= 10)",
			expected: &eval.Expression{
				Type:     eval.Operator,
				Operator: "and",
				Left: &eval.Expression{
					Type:     eval.Operator,
					Operator: "greater",
					Left: &eval.Expression{
						Type:       eval.Operand,
						Identifier: "x",
					},
					Right: &eval.Expression{
						Type:    eval.Operand,
						GoValue: float64(0),
					},
				},
				Right: &eval.Expression{
					Type:     eval.Operator,
					Operator: "less_equal",
					Left: &eval.Expression{
						Type:       eval.Operand,
						Identifier: "y",
					},
					Right: &eval.Expression{
						Type:    eval.Operand,
						GoValue: float64(10),
					},
				},
			},
		},
		{
			input: "(a > 0 and b <= 10) or (c == \"hello\" and d != 5)",
			expected: &eval.Expression{
				Type:     eval.Operator,
				Operator: "or",
				Left: &eval.Expression{
					Type:     eval.Operator,
					Operator: "and",
					Left: &eval.Expression{
						Type:     eval.Operator,
						Operator: "greater",
						Left: &eval.Expression{
							Type:       eval.Operand,
							Identifier: "a",
						},
						Right: &eval.Expression{
							Type:    eval.Operand,
							GoValue: float64(0),
						},
					},
					Right: &eval.Expression{
						Type:     eval.Operator,
						Operator: "less_equal",
						Left: &eval.Expression{
							Type:       eval.Operand,
							Identifier: "b",
						},
						Right: &eval.Expression{
							Type:    eval.Operand,
							GoValue: float64(10),
						},
					},
				},
				Right: &eval.Expression{
					Type:     eval.Operator,
					Operator: "and",
					Left: &eval.Expression{
						Type:     eval.Operator,
						Operator: "equal",
						Left: &eval.Expression{
							Type:       eval.Operand,
							Identifier: "c",
						},
						Right: &eval.Expression{
							Type:    eval.Operand,
							GoValue: "hello",
						},
					},
					Right: &eval.Expression{
						Type:     eval.Operator,
						Operator: "not_equal",
						Left: &eval.Expression{
							Type: eval.Operand,

							Identifier: "d",
						},
						Right: &eval.Expression{
							Type:    eval.Operand,
							GoValue: float64(5),
						},
					},
				},
			},
		},
	}

	for i, tt := range tests {
		input := mustTokenize(tt.input)

		got, err := infixToExpressionTree(input)
		gotErr := ""
		if err != nil {
			gotErr = err.Error()
		}

		if gotErr != "" {
			if tt.expectedErr != gotErr {
				t.Errorf("test %d failed: expected err '%s', but got '%s'", i+1, tt.expectedErr, err.Error())
			}
			continue
		}

		if diff := cmp.Diff(got, tt.expected, cmp.AllowUnexported(eval.Expression{})); diff != "" {
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
				{_type: "identifier", strValue: "foo", line: 1, column: 68},
				{_type: "end_of_statement", strValue: ";", line: 1, column: 72},
			},
		},
		{
			input: "; foo <= >= < > != == oR AnD () 123.321 123 \"string\" FaLse TrUe , sElEcT",
			expected: []token{
				{_type: "end_of_statement", strValue: ";", line: 1, column: 1},
				{_type: "identifier", strValue: "foo", line: 1, column: 3},
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

	if p.lookahead._type != "identifier" {
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

	if p.lookahead._type != "identifier" {
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

	if tk._type != "identifier" {
		t.Errorf("expected consumed token to be identifier, but got '%s'", tk._type)
		return
	}

	if p.lookahead != tokenNoop {
		t.Errorf("expected lookahead token to be noop, but got '%s'", p.lookahead._type)
	}
}

func Test_parser_selectBody(t *testing.T) {
	type test struct {
		input       string
		expected    []*eval.Expression
		expectedErr string
	}
	tests := []test{
		{
			input: "foo, bar, 1 == 1, 2 <= 2, (1 != 2) and (true == false)",
			expected: []*eval.Expression{
				{
					Type: "operand",

					Identifier: "foo",
				},
				{
					Type: "operand",

					Identifier: "bar",
				},
				{
					Type:     "operator",
					Operator: "equal",
					Left: &eval.Expression{
						Type:    "operand",
						GoValue: float64(1),
					},
					Right: &eval.Expression{
						Type:    "operand",
						GoValue: float64(1),
					},
				},
				{
					Type:     "operator",
					Operator: "less_equal",
					Left: &eval.Expression{
						Type:    "operand",
						GoValue: float64(2),
					},
					Right: &eval.Expression{
						Type:    "operand",
						GoValue: float64(2),
					},
				},
				{
					Type:     "operator",
					Operator: "and",
					Left: &eval.Expression{
						Type:     "operator",
						Operator: "not_equal",
						Left: &eval.Expression{
							Type:    "operand",
							GoValue: float64(1),
						},
						Right: &eval.Expression{
							Type:    "operand",
							GoValue: float64(2),
						},
					},
					Right: &eval.Expression{
						Type:     "operator",
						Operator: "equal",
						Left: &eval.Expression{
							Type:    "operand",
							GoValue: true,
						},
						Right: &eval.Expression{
							Type:    "operand",
							GoValue: false,
						},
					},
				},
			},
		},
		{
			input: "foo",
			expected: []*eval.Expression{
				{
					Type:       "operand",
					Identifier: "foo",
				},
			},
		},
		{
			input: "1 == 1",
			expected: []*eval.Expression{
				{
					Type:     "operator",
					Operator: "equal",
					Left: &eval.Expression{
						Type:    "operand",
						GoValue: float64(1),
					},
					Right: &eval.Expression{
						Type:    "operand",
						GoValue: float64(1),
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
			expectedErr: "invalid eval.Expression at 1:1",
		},
		{
			input:       "select",
			expectedErr: "invalid eval.Expression at 1:1",
		},
		{
			input:       " , , ",
			expectedErr: "invalid eval.Expression at 1:2",
		},
	}

	for i, tt := range tests {
		p := NewParser(tt.input)

		err := p.moveToNextToken()
		if err != nil {
			t.Error(err)
			return
		}

		got, err := p.selectBody()
		gotErr := ""
		if err != nil {
			gotErr = err.Error()
		}

		if gotErr != "" {
			if tt.expectedErr != gotErr {
				t.Errorf("test %d failed: expected err '%s', but got '%s'", i+1, tt.expectedErr, err.Error())
			}
			continue
		}

		if diff := cmp.Diff(got, tt.expected, cmp.AllowUnexported(token{}, eval.Expression{})); diff != "" {
			t.Errorf("test %d failed: %s", i+1, diff)
		}
	}
}

func Test_parser_definitionsBody(t *testing.T) {
	type test struct {
		input       string
		expected    []*schema.NewColumn
		expectedErr string
	}
	tests := []test{
		{
			input: "(foo int, bar string, baz bool)",
			expected: []*schema.NewColumn{
				{Name: "foo", Type: schema.Int32Type},
				{Name: "bar", Type: schema.StringType},
				{Name: "baz", Type: schema.BoolType},
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
			expected: []*schema.NewColumn{
				{Name: "foo", Type: schema.Int32Type},
			},
		},
	}

	for i, tt := range tests {
		p := NewParser(tt.input)

		err := p.moveToNextToken()
		if err != nil {
			t.Error(err)
			return
		}

		got, err := p.definitionsBody()
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

		if diff := cmp.Diff(got, tt.expected, cmp.AllowUnexported(schema.NewColumn{})); diff != "" {
			t.Errorf("test %d failed: %s", i+1, diff)
		}
	}
}

func Test_parser_identifier(t *testing.T) {
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
			expectedErr: "expected identifier, but got 'number_literal' at 1:1",
		},
		{
			input:       "true",
			expectedErr: "expected identifier, but got 'boolean_literal' at 1:1",
		},
		{
			input:       "bool",
			expectedErr: "expected identifier, but got 'data_type' at 1:1",
		},
	}

	for i, tt := range tests {
		p := NewParser(tt.input)

		err := p.moveToNextToken()
		if err != nil {
			t.Error(err)
			return
		}

		got, err := p.identifier()
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

		if diff := cmp.Diff(got, tt.expected, cmp.AllowUnexported(schema.NewColumn{})); diff != "" {
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
		p := NewParser(tt.input)

		err := p.moveToNextToken()
		if err != nil {
			t.Error(err)
			return
		}

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

		if diff := cmp.Diff(got, tt.expected, cmp.AllowUnexported(schema.NewColumn{})); diff != "" {
			t.Errorf("test %d failed: %s", i+1, diff)
		}
	}
}

func Test_parser_whereBody(t *testing.T) {
	type test struct {
		input       string
		expected    *eval.Expression
		expectedErr string
	}
	tests := []test{
		{
			input: "a == 1",
			expected: &eval.Expression{
				Type:     eval.Operator,
				Operator: "equal",
				Left: &eval.Expression{
					Type:       eval.Operand,
					Identifier: "a",
				},
				Right: &eval.Expression{
					Type:    eval.Operand,
					GoValue: float64(1),
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
		p := NewParser(tt.input)

		err := p.moveToNextToken()
		if err != nil {
			t.Error(err)
			return
		}

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

		if diff := cmp.Diff(got, tt.expected, cmp.AllowUnexported(eval.Expression{})); diff != "" {
			t.Errorf("test %d failed: %s", i+1, diff)
		}
	}
}
