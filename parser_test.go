package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

var (
	leftParenthesis  = token{_type: "left_parenthesis", value: ")"}
	rightParenthesis = token{_type: "right_parenthesis", value: "("}
	and              = token{_type: "and", value: "and"}
	or               = token{_type: "or", value: "or"}
	equal            = token{_type: "equal", value: "=="}
	not_equal        = token{_type: "not_equal", value: "!="}
)

func TestQuery(t *testing.T) {
	p := &parser{}
	q, err := p.Parse(`SELeCT foo     ,    bar FROM       jobs  where (foo ==  "  bbbbasdasd asd asd ") or (bar >= 1.0);`)
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
					{_type: "left_parenthesis", value: "("},
					{_type: "name", value: "foo"},
					{_type: "equal", value: "=="},
					{_type: "string_literal", value: "  bbbbasdasd asd asd "},
					{_type: "right_parenthesis", value: ")"},
					{_type: "or", value: "or"},
					{_type: "left_parenthesis", value: "("},
					{_type: "name", value: "bar"},
					{_type: "greater_equal", value: ">="},
					{_type: "decimal_literal", value: "1.0"},
					{_type: "right_parenthesis", value: ")"},
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
					{_type: "left_parenthesis", value: "("},
					{_type: "keyword", value: "select"},
				},
			},
			wantErr: "'select' is not valid as part of an expression",
		},
		{
			name: "invalid beginning of expression",
			args: args{
				tokens: []token{
					{_type: "and", value: "and"},
					{_type: "keyword", value: "select"},
				},
			},
			wantErr: "can't start expression with operator 'and'",
		},
		{
			name: "invalid token after operand",
			args: args{
				tokens: []token{
					{_type: "string_literal", value: "asdasda"},
					{_type: "decimal_literal", value: "121"},
				},
			},
			wantErr: "expected operator after 'asdasda'",
		},
		{
			name: "invalid token after operator",
			args: args{
				tokens: []token{
					{_type: "string_literal", value: "asdasda"},
					{_type: "equal", value: "=="},
					{_type: "not_equal", value: "!="},
					{_type: "string_literal", value: "asdasda"},
				},
			},
			wantErr: "expected operand after '=='",
		},
		{
			name: "empty parentheses",
			args: args{
				tokens: []token{
					{_type: "left_parenthesis", value: "("},
					{_type: "right_parenthesis", value: ")"},
				},
			},
			wantErr: "empty parentheses",
		},
		{
			name: "invalid token after left_parenthesis",
			args: args{
				tokens: []token{
					{_type: "left_parenthesis", value: "("},
					{_type: "and", value: "and"},
				},
			},
			wantErr: "a left parenthesis can't precede an operator",
		},
		{
			name: "ending expression with operator",
			args: args{
				tokens: []token{
					{_type: "right_parenthesis", value: ")"},
					{_type: "and", value: "and"},
				},
			},
			wantErr: "can't end expression with an operator 'and'",
		},
		{
			name: "happy path",
			args: args{
				tokens: []token{
					{_type: "left_parenthesis", value: "("},
					{_type: "name", value: "foo"},
					{_type: "equal", value: "=="},
					{_type: "string_literal", value: "\"bbbbasdasd asd asd \""},
					{_type: "right_parenthesis", value: ")"},
					{_type: "or", value: "or"},
					{_type: "left_parenthesis", value: "("},
					{_type: "name", value: "bar"},
					{_type: "greater_equal", value: ">="},
					{_type: "decimal_literal", value: "1.0"},
					{_type: "right_parenthesis", value: ")"},
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
		if tt.valid != checkParenthesesBalance(tt.tokens) {
			t.Errorf("expected balance test %d to be successful", i+1)
		}
	}
}

func Test_tokensToExpressionTree(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		want       *expression
	}{
		// {
		// 	name:       "test 1",
		// 	expression: "foo == \"aaa\"",
		// 	want: &expression{
		// 		_type: "equal",
		// 		value: "==",
		// 		left: &expression{
		// 			_type: "name",
		// 			value: "foo",
		// 		},
		// 		right: &expression{
		// 			_type: "string_literal",
		// 			value: "aaa",
		// 		},
		// 	},
		// },
		// {
		// 	name:       "test 2",
		// 	expression: "foo == \"aaa\" and bar != true",
		// 	want: &expression{
		// 		_type: "and",
		// 		value: "and",
		// 		left: &expression{
		// 			_type: "equal",
		// 			value: "==",
		// 			left: &expression{
		// 				_type: "name",
		// 				value: "foo",
		// 			},
		// 			right: &expression{
		// 				_type: "string_literal",
		// 				value: "aaa",
		// 			},
		// 		},
		// 		right: &expression{
		// 			_type: "not_equal",
		// 			value: "!=",
		// 			left: &expression{
		// 				_type: "name",
		// 				value: "bar",
		// 			},
		// 			right: &expression{
		// 				_type: "boolean_literal",
		// 				value: "true",
		// 			},
		// 		},
		// 	},
		// },
		{
			name:       "test 3",
			expression: "foo == \"aaa\" and ((bar != true and aaa >= 34.12) or foo == 123)",
			want: &expression{
				_type: "and",
				value: "and",
				left: &expression{
					_type: "equal",
					value: "==",
					left: &expression{
						_type: "name",
						value: "foo",
					},
					right: &expression{
						_type: "string_literal",
						value: "aaa",
					},
				},
				right: &expression{
					_type: "or",
					value: "or",
					left: &expression{
						_type: "and",
						value: "and",
						left: &expression{
							_type: "not_equal",
							value: "!=",
							left: &expression{
								_type: "name",
								value: "bar",
							},
							right: &expression{
								_type: "boolean_literal",
								value: "true",
							},
						},
						right: &expression{
							_type: "greater_equal",
							value: ">=",
							left: &expression{
								_type: "name",
								value: "aaa",
							},
							right: &expression{
								_type: "decimal_literal",
								value: "34.12",
							},
						},
					},
					right: &expression{
						_type: "equal",
						value: "==",
						left: &expression{
							_type: "name",
							value: "foo",
						},
						right: &expression{
							_type: "integer_literal",
							value: "123",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser{
				t: &tokenizer{
					query: tt.expression,
				},
			}
			tokens := p.mustTokenize()
			got := tokensToExpressionTree(tokens)
			if diff := cmp.Diff(got, tt.want, cmp.AllowUnexported(expression{})); diff != "" {
				t.Error(diff)
			}
		})
	}
}
