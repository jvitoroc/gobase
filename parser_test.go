package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

var (
	leftParenthesis  = token{_type: "left_parenthesis", strValue: ")"}
	rightParenthesis = token{_type: "right_parenthesis", strValue: "("}
)

func TestRidiculousSelect(t *testing.T) {
	p := &parser{}
	q, err := p.parse(`SELeCT foo    , bar    FROM       jobs  where (foo == 
		 "  bbbbasdasd asd asd ") or (bar >= 1.0);`)
	if err != nil {
		t.Error(err)
	}
	diff := cmp.Diff(q, []*statement{
		{
			Parts: []*Part{
				{
					Keyword: "select",
					Body: []*expression{
						{_type: "operand", valueType: "name", strValue: "foo"},
						{_type: "operand", valueType: "name", strValue: "bar"},
					},
				},
				{
					Keyword: "from",
					Body:    "jobs",
				},
				{
					Keyword: "where",
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
	p := &parser{}
	q, err := p.parse(`
		CREATE TABLE foo DEFINITIONS (
			foo bool,
			bar int,
			baz string
		);
	`)
	if err != nil {
		t.Error(err)
	}
	diff := cmp.Diff(q, []*statement{
		{
			Parts: []*Part{
				{
					Keyword: "create table",
					Body:    "foo",
				},
				{
					Keyword: "definitions",
					Body: []*Column{
						{Name: "foo", Type: BoolType},
						{Name: "bar", Type: Int32Type},
						{Name: "baz", Type: StringType},
					},
				},
			},
		},
	}, cmp.AllowUnexported(token{}, expression{}))
	if diff != "" {
		t.Error(diff)
	}
}

func TestInsertInto(t *testing.T) {
	p := &parser{}
	q, err := p.parse(`
		INSERT INTO foo VALUES (true, 123, "foobarbaz");
	`)
	if err != nil {
		t.Error(err)
	}
	diff := cmp.Diff(q, []*statement{
		{
			Parts: []*Part{
				{
					Keyword: "insert into",
					Body:    "foo",
				},
				{
					Keyword: "values",
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
	p := &parser{}
	q, err := p.parse(`
		CREATE TABLE foo DEFINITIONS (
			foo bool,
			bar int,
			baz string
		);

		INSERT INTO foo VALUES (true, 123, "foobarbaz");

		SELECT foo, bar, baz FROM foo WHERE foo != false AND bar > 100;
	`)
	if err != nil {
		t.Error(err)
	}
	diff := cmp.Diff(q, []*statement{
		{
			Parts: []*Part{
				{
					Keyword: "create table",
					Body:    "foo",
				},
				{
					Keyword: "definitions",
					Body: []*Column{
						{Name: "foo", Type: BoolType},
						{Name: "bar", Type: Int32Type},
						{Name: "baz", Type: StringType},
					},
				},
			},
		},
		{
			Parts: []*Part{
				{
					Keyword: "insert into",
					Body:    "foo",
				},
				{
					Keyword: "values",
					Body: []string{
						"true",
						"123",
						"foobarbaz",
					},
				},
			},
		},
		{
			Parts: []*Part{
				{
					Keyword: "select",
					Body: []*expression{
						{_type: Operand, valueType: "name", strValue: "foo"},
						{_type: Operand, valueType: "name", strValue: "bar"},
						{_type: Operand, valueType: "name", strValue: "baz"},
					},
				},
				{
					Keyword: "from",
					Body:    "foo",
				},
				{
					Keyword: "where",
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
	}, cmp.AllowUnexported(token{}, expression{}))
	if diff != "" {
		t.Error(diff)
	}
}

func Test_checkBooleanExpressionSyntax(t *testing.T) {
	type args struct {
		tokens []token
	}
	tests := []struct {
		name    string
		args    args
		wantErr string
	}{
		{
			name: "nil",
			args: args{
				tokens: nil,
			},
			wantErr: "empty expression",
		},
		{
			name: "empty body",
			args: args{
				tokens: []token{},
			},
			wantErr: "empty expression",
		},
		{
			name: "invalid token",
			args: args{
				tokens: []token{
					{_type: "left_parenthesis", strValue: "("},
					{_type: "keyword", strValue: "select"},
				},
			},
			wantErr: "'select' at 0:0 is not valid as part of an expression",
		},
		{
			name: "invalid beginning of expression",
			args: args{
				tokens: []token{
					{_type: "and", strValue: "and"},
					{_type: "keyword", strValue: "select"},
				},
			},
			wantErr: "can't start expression with operator 'and' at 0:0",
		},
		{
			name: "invalid token after operand",
			args: args{
				tokens: []token{
					{_type: "string_literal", strValue: "asdasda"},
					{_type: "number_literal", strValue: "121"},
				},
			},
			wantErr: "expected operator after 'asdasda' at 0:0",
		},
		{
			name: "invalid token after operator",
			args: args{
				tokens: []token{
					{_type: "string_literal", strValue: "asdasda"},
					{_type: "equal", strValue: "=="},
					{_type: "not_equal", strValue: "!="},
					{_type: "string_literal", strValue: "asdasda"},
				},
			},
			wantErr: "expected operand after '==' at 0:0",
		},
		{
			name: "empty parentheses",
			args: args{
				tokens: []token{
					{_type: "left_parenthesis", strValue: "("},
					{_type: "right_parenthesis", strValue: ")"},
				},
			},
			wantErr: "empty parentheses at 0:0",
		},
		{
			name: "invalid token after left_parenthesis",
			args: args{
				tokens: []token{
					{_type: "left_parenthesis", strValue: "("},
					{_type: "and", strValue: "and"},
				},
			},
			wantErr: "an operator is not allowed to be positioned at 0:0 after a opening parenthesis",
		},
		{
			name: "ending expression with operator",
			args: args{
				tokens: []token{
					{_type: "right_parenthesis", strValue: ")"},
					{_type: "and", strValue: "and"},
				},
			},
			wantErr: "can't end expression with an operator 'and' at 0:0",
		},
		{
			name: "happy path",
			args: args{
				tokens: []token{
					{_type: "left_parenthesis", strValue: "("},
					{_type: "name", strValue: "foo"},
					{_type: "equal", strValue: "=="},
					{_type: "string_literal", strValue: "\"bbbbasdasd asd asd \""},
					{_type: "right_parenthesis", strValue: ")"},
					{_type: "or", strValue: "or"},
					{_type: "left_parenthesis", strValue: "("},
					{_type: "name", strValue: "bar"},
					{_type: "greater_equal", strValue: ">="},
					{_type: "number_literal", strValue: "1.0"},
					{_type: "right_parenthesis", strValue: ")"},
				},
			},
			wantErr: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkBooleanExpressionSyntax(tt.args.tokens)
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
		tokens []token
		valid  bool
	}
	tests := []test{
		{valid: true, tokens: []token{leftParenthesis, rightParenthesis}},
		{valid: true, tokens: []token{leftParenthesis, rightParenthesis, leftParenthesis, rightParenthesis, leftParenthesis, rightParenthesis}},
		{valid: false, tokens: []token{leftParenthesis, leftParenthesis}},
		{valid: false, tokens: []token{rightParenthesis, rightParenthesis}},
		{valid: false, tokens: []token{leftParenthesis, leftParenthesis, leftParenthesis, leftParenthesis, rightParenthesis, rightParenthesis}},
		{valid: true, tokens: []token{leftParenthesis, leftParenthesis, leftParenthesis, leftParenthesis, rightParenthesis, rightParenthesis, rightParenthesis, rightParenthesis}},
	}

	for i, tt := range tests {
		if err := checkParenthesesBalance(tt.tokens); (err != nil) == tt.valid {
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
