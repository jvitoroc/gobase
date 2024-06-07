package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

var (
	leftParenthesis  = token{_type: "left_parenthesis", valueStr: ")"}
	rightParenthesis = token{_type: "right_parenthesis", valueStr: "("}
)

func Test_parse(t *testing.T) {
	p := &parser{}
	q, err := p.parse(`SELeCT foo     , bar    FROM       jobs  where (foo == 
		 "  bbbbasdasd asd asd ") or (bar >= 1.0);`)
	if err != nil {
		t.Error(err)
	}
	diff := cmp.Diff(q, &statement{
		Parts: []*Part{
			{
				Keyword: "select",
				Body:    []string{"foo", "bar"},
			},
			{
				Keyword: "from",
				Body:    "jobs",
			},
			{
				Keyword: "where",
				Body: []token{
					{_type: "name", valueStr: "foo", line: 1, column: 49},
					{_type: "string_literal", valueStr: "  bbbbasdasd asd asd ", line: 2, column: 4},
					{_type: "equal", valueStr: "==", line: 1, column: 53},
					{_type: "name", valueStr: "bar", line: 2, column: 33},
					{_type: "decimal_literal", valueStr: "1.0", line: 2, column: 40},
					{_type: "greater_equal", valueStr: ">=", line: 2, column: 37},
					{_type: "or", valueStr: "or", line: 2, column: 29},
				},
			},
		},
	}, cmp.AllowUnexported(token{}))
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
					{_type: "left_parenthesis", valueStr: "("},
					{_type: "keyword", valueStr: "select"},
				},
			},
			wantErr: "'select' is not valid as part of an expression",
		},
		{
			name: "invalid beginning of expression",
			args: args{
				tokens: []token{
					{_type: "and", valueStr: "and"},
					{_type: "keyword", valueStr: "select"},
				},
			},
			wantErr: "can't start expression with operator 'and'",
		},
		{
			name: "invalid token after operand",
			args: args{
				tokens: []token{
					{_type: "string_literal", valueStr: "asdasda"},
					{_type: "decimal_literal", valueStr: "121"},
				},
			},
			wantErr: "expected operator after 'asdasda'",
		},
		{
			name: "invalid token after operator",
			args: args{
				tokens: []token{
					{_type: "string_literal", valueStr: "asdasda"},
					{_type: "equal", valueStr: "=="},
					{_type: "not_equal", valueStr: "!="},
					{_type: "string_literal", valueStr: "asdasda"},
				},
			},
			wantErr: "expected operand after '=='",
		},
		{
			name: "empty parentheses",
			args: args{
				tokens: []token{
					{_type: "left_parenthesis", valueStr: "("},
					{_type: "right_parenthesis", valueStr: ")"},
				},
			},
			wantErr: "empty parentheses",
		},
		{
			name: "invalid token after left_parenthesis",
			args: args{
				tokens: []token{
					{_type: "left_parenthesis", valueStr: "("},
					{_type: "and", valueStr: "and"},
				},
			},
			wantErr: "a left parenthesis can't precede an operator",
		},
		{
			name: "ending expression with operator",
			args: args{
				tokens: []token{
					{_type: "right_parenthesis", valueStr: ")"},
					{_type: "and", valueStr: "and"},
				},
			},
			wantErr: "can't end expression with an operator 'and'",
		},
		{
			name: "happy path",
			args: args{
				tokens: []token{
					{_type: "left_parenthesis", valueStr: "("},
					{_type: "name", valueStr: "foo"},
					{_type: "equal", valueStr: "=="},
					{_type: "string_literal", valueStr: "\"bbbbasdasd asd asd \""},
					{_type: "right_parenthesis", valueStr: ")"},
					{_type: "or", valueStr: "or"},
					{_type: "left_parenthesis", valueStr: "("},
					{_type: "name", valueStr: "bar"},
					{_type: "greater_equal", valueStr: ">="},
					{_type: "decimal_literal", valueStr: "1.0"},
					{_type: "right_parenthesis", valueStr: ")"},
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

// func Test_tokensToExpressionTree(t *testing.T) {
// 	tests := []struct {
// 		name       string
// 		expression string
// 		want       *expression
// 	}{
// 		{
// 			name:       "test 1",
// 			expression: "foo == \"aaa\"",
// 			want: &expression{
// 				_type: "equal",
// 				value: "==",
// 				left: &expression{
// 					_type: "name",
// 					value: "foo",
// 				},
// 				right: &expression{
// 					_type: "string_literal",
// 					value: "aaa",
// 				},
// 			},
// 		},
// 		{
// 			name:       "test 2",
// 			expression: "foo == \"aaa\" and bar != true",
// 			want: &expression{
// 				_type: "and",
// 				value: "and",
// 				left: &expression{
// 					_type: "equal",
// 					value: "==",
// 					left: &expression{
// 						_type: "name",
// 						value: "foo",
// 					},
// 					right: &expression{
// 						_type: "string_literal",
// 						value: "aaa",
// 					},
// 				},
// 				right: &expression{
// 					_type: "not_equal",
// 					value: "!=",
// 					left: &expression{
// 						_type: "name",
// 						value: "bar",
// 					},
// 					right: &expression{
// 						_type: "boolean_literal",
// 						value: "true",
// 					},
// 				},
// 			},
// 		},
// 		{
// 			name:       "test 3",
// 			expression: "(foo == \"aaa\") and ((bar != true and (aaa >= 34.12 OR foo == false)) or foo == 123)",
// 			want: &expression{
// 				_type: "and",
// 				value: "and",
// 				left: &expression{
// 					_type: "equal",
// 					value: "==",
// 					left: &expression{
// 						_type: "name",
// 						value: "foo",
// 					},
// 					right: &expression{
// 						_type: "string_literal",
// 						value: "aaa",
// 					},
// 				},
// 				right: &expression{
// 					_type: "or",
// 					value: "or",
// 					left: &expression{
// 						_type: "and",
// 						value: "and",
// 						left: &expression{
// 							_type: "not_equal",
// 							value: "!=",
// 							left: &expression{
// 								_type: "name",
// 								value: "bar",
// 							},
// 							right: &expression{
// 								_type: "boolean_literal",
// 								value: "true",
// 							},
// 						},
// 						right: &expression{
// 							_type: "or",
// 							value: "or",
// 							left: &expression{
// 								_type: "greater_equal",
// 								value: ">=",
// 								left: &expression{
// 									_type: "name",
// 									value: "aaa",
// 								},
// 								right: &expression{
// 									_type: "decimal_literal",
// 									value: "34.12",
// 								},
// 							},
// 							right: &expression{
// 								_type: "equal",
// 								value: "==",
// 								left: &expression{
// 									_type: "name",
// 									value: "foo",
// 								},
// 								right: &expression{
// 									_type: "boolean_literal",
// 									value: "false",
// 								},
// 							},
// 						},
// 					},
// 					right: &expression{
// 						_type: "equal",
// 						value: "==",
// 						left: &expression{
// 							_type: "name",
// 							value: "foo",
// 						},
// 						right: &expression{
// 							_type: "integer_literal",
// 							value: "123",
// 						},
// 					},
// 				},
// 			},
// 		},
// 		{
// 			name:       "test 4",
// 			expression: "foo == \"aaa\" and ((bar != true and aaa >= 34.12) or foo == 123)",
// 			want: &expression{
// 				_type: "and",
// 				value: "and",
// 				left: &expression{
// 					_type: "equal",
// 					value: "==",
// 					left: &expression{
// 						_type: "name",
// 						value: "foo",
// 					},
// 					right: &expression{
// 						_type: "string_literal",
// 						value: "aaa",
// 					},
// 				},
// 				right: &expression{
// 					_type: "or",
// 					value: "or",
// 					left: &expression{
// 						_type: "and",
// 						value: "and",
// 						left: &expression{
// 							_type: "not_equal",
// 							value: "!=",
// 							left: &expression{
// 								_type: "name",
// 								value: "bar",
// 							},
// 							right: &expression{
// 								_type: "boolean_literal",
// 								value: "true",
// 							},
// 						},
// 						right: &expression{
// 							_type: "greater_equal",
// 							value: ">=",
// 							left: &expression{
// 								_type: "name",
// 								value: "aaa",
// 							},
// 							right: &expression{
// 								_type: "decimal_literal",
// 								value: "34.12",
// 							},
// 						},
// 					},
// 					right: &expression{
// 						_type: "equal",
// 						value: "==",
// 						left: &expression{
// 							_type: "name",
// 							value: "foo",
// 						},
// 						right: &expression{
// 							_type: "integer_literal",
// 							value: "123",
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			p := parser{
// 				t: &tokenizer{
// 					query: tt.expression,
// 				},
// 			}
// 			tokens := p.mustTokenize()
// 			got := tokensToExpressionTree(tokens)
// 			if diff := cmp.Diff(got, tt.want, cmp.AllowUnexported(expression{})); diff != "" {
// 				t.Error(diff)
// 			}
// 		})
// 	}
// }

func Test_infixToPostfix(t *testing.T) {
	type test struct {
		input    []token
		expected []token
	}
	tests := []test{
		{
			input: []token{
				{_type: "integer_literal", valueStr: "1"},
				{_type: "greater"},
				{_type: "integer_literal", valueStr: "2"},
			},
			expected: []token{
				{_type: "integer_literal", valueStr: "1"},
				{_type: "integer_literal", valueStr: "2"},
				{_type: "greater"},
			},
		},
		{
			input: []token{
				{_type: "integer_literal", valueStr: "1"},
				{_type: "greater_equal"},
				{_type: "integer_literal", valueStr: "2"},
				{_type: "or"},
				{_type: "left_parenthesis"},
				{_type: "integer_literal", valueStr: "3"},
				{_type: "greater"},
				{_type: "integer_literal", valueStr: "4"},
				{_type: "right_parenthesis"},
			},
			expected: []token{
				{_type: "integer_literal", valueStr: "1"},
				{_type: "integer_literal", valueStr: "2"},
				{_type: "greater_equal"},
				{_type: "integer_literal", valueStr: "3"},
				{_type: "integer_literal", valueStr: "4"},
				{_type: "greater"},
				{_type: "or"},
			},
		},
		{
			input: []token{
				{_type: "integer_literal", valueStr: "1"},
				{_type: "greater"},
				{_type: "integer_literal", valueStr: "2"},
				{_type: "or"},
				{_type: "integer_literal", valueStr: "3"},
				{_type: "less"},
				{_type: "integer_literal", valueStr: "4"},
			},
			expected: []token{
				{_type: "integer_literal", valueStr: "1"},
				{_type: "integer_literal", valueStr: "2"},
				{_type: "greater"},
				{_type: "integer_literal", valueStr: "3"},
				{_type: "integer_literal", valueStr: "4"},
				{_type: "less"},
				{_type: "or"},
			},
		},
		{
			input: []token{
				{_type: "left_parenthesis"},
				{_type: "integer_literal", valueStr: "1"},
				{_type: "greater"},
				{_type: "integer_literal", valueStr: "2"},
				{_type: "or"},
				{_type: "left_parenthesis"},
				{_type: "integer_literal", valueStr: "3"},
				{_type: "less"},
				{_type: "integer_literal", valueStr: "4"},
				{_type: "or"},
				{_type: "integer_literal", valueStr: "5"},
				{_type: "greater"},
				{_type: "integer_literal", valueStr: "6"},
				{_type: "right_parenthesis"},
				{_type: "and"},
				{_type: "left_parenthesis"},
				{_type: "integer_literal", valueStr: "7"},
				{_type: "greater"},
				{_type: "integer_literal", valueStr: "8"},
				{_type: "right_parenthesis"},
				{_type: "right_parenthesis"},
			},
			expected: []token{
				{_type: "integer_literal", valueStr: "1"},
				{_type: "integer_literal", valueStr: "2"},
				{_type: "greater"},
				{_type: "integer_literal", valueStr: "3"},
				{_type: "integer_literal", valueStr: "4"},
				{_type: "less"},
				{_type: "integer_literal", valueStr: "5"},
				{_type: "integer_literal", valueStr: "6"},
				{_type: "greater"},
				{_type: "or"},
				{_type: "integer_literal", valueStr: "7"},
				{_type: "integer_literal", valueStr: "8"},
				{_type: "greater"},
				{_type: "and"},
				{_type: "or"},
			},
		},
		{
			input: []token{
				{_type: "integer_literal", valueStr: "1"},
				{_type: "greater"},
				{_type: "integer_literal", valueStr: "2"},
				{_type: "or"},
				{_type: "integer_literal", valueStr: "3"},
				{_type: "less"},
				{_type: "integer_literal", valueStr: "4"},
				{_type: "or"},
				{_type: "integer_literal", valueStr: "5"},
				{_type: "greater"},
				{_type: "integer_literal", valueStr: "6"},
				{_type: "and"},
				{_type: "integer_literal", valueStr: "7"},
				{_type: "greater"},
				{_type: "integer_literal", valueStr: "8"},
			},
			expected: []token{
				{_type: "integer_literal", valueStr: "1"},
				{_type: "integer_literal", valueStr: "2"},
				{_type: "greater"},
				{_type: "integer_literal", valueStr: "3"},
				{_type: "integer_literal", valueStr: "4"},
				{_type: "less"},
				{_type: "or"},
				{_type: "integer_literal", valueStr: "5"},
				{_type: "integer_literal", valueStr: "6"},
				{_type: "greater"},
				{_type: "integer_literal", valueStr: "7"},
				{_type: "integer_literal", valueStr: "8"},
				{_type: "greater"},
				{_type: "and"},
				{_type: "or"},
			},
		},
		{
			input: []token{
				{_type: "left_parenthesis"},
				{_type: "integer_literal", valueStr: "1"},
				{_type: "greater"},
				{_type: "integer_literal", valueStr: "2"},
				{_type: "or"},
				{_type: "integer_literal", valueStr: "3"},
				{_type: "less"},
				{_type: "integer_literal", valueStr: "4"},
				{_type: "or"},
				{_type: "integer_literal", valueStr: "5"},
				{_type: "greater"},
				{_type: "integer_literal", valueStr: "6"},
				{_type: "right_parenthesis"},
				{_type: "and"},
				{_type: "integer_literal", valueStr: "7"},
				{_type: "greater"},
				{_type: "integer_literal", valueStr: "8"},
			},
			expected: []token{
				{_type: "integer_literal", valueStr: "1"},
				{_type: "integer_literal", valueStr: "2"},
				{_type: "greater"},
				{_type: "integer_literal", valueStr: "3"},
				{_type: "integer_literal", valueStr: "4"},
				{_type: "less"},
				{_type: "or"},
				{_type: "integer_literal", valueStr: "5"},
				{_type: "integer_literal", valueStr: "6"},
				{_type: "greater"},
				{_type: "or"},
				{_type: "integer_literal", valueStr: "7"},
				{_type: "integer_literal", valueStr: "8"},
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
