package sql

import (
	"fmt"
	"regexp"
	"strings"
)

type tokenType string

const (
	clause           tokenType = "clause"
	dataType         tokenType = "data_type"
	comma            tokenType = "comma"
	booleanLiteral   tokenType = "boolean_literal"
	stringLiteral    tokenType = "string_literal"
	numberLiteral    tokenType = "number_literal"
	leftParenthesis  tokenType = "left_parenthesis"
	rightParenthesis tokenType = "right_parenthesis"
	and              tokenType = "and"
	or               tokenType = "or"
	equal            tokenType = "equal"
	notEqual         tokenType = "not_equal"
	greaterEqual     tokenType = "greater_equal"
	greater          tokenType = "greater"
	lessEqual        tokenType = "less_equal"
	less             tokenType = "less"
	identifier       tokenType = "identifier"
	whitespace       tokenType = "whitespace"
	endOfStatement   tokenType = "end_of_statement"
	invalid          tokenType = "invalid"
)

type tokenRegexps struct {
	name    tokenType
	regexps []*regexp.Regexp
}

var (
	regexps = []*tokenRegexps{
		{
			name:    clause,
			regexps: []*regexp.Regexp{regexp.MustCompile(`(?i)^(SELECT|FROM|(INSERT\s+INTO)|WHERE|(CREATE\s+TABLE)|DEFINITIONS|VALUES)\b`)},
		},
		{
			name:    dataType,
			regexps: []*regexp.Regexp{regexp.MustCompile(`(?i)^(int|string|bool)\b`)},
		},
		{
			name:    comma,
			regexps: []*regexp.Regexp{regexp.MustCompile(`^,`)},
		},
		{
			name:    booleanLiteral,
			regexps: []*regexp.Regexp{regexp.MustCompile(`(?i)^(TRUE|FALSE)\b`)},
		},
		{
			name:    stringLiteral,
			regexps: []*regexp.Regexp{regexp.MustCompile(`^"([^"]*)"`)},
		},
		{
			name:    numberLiteral,
			regexps: []*regexp.Regexp{regexp.MustCompile(`^\d+(\.\d+)?`)},
		},
		{
			name:    leftParenthesis,
			regexps: []*regexp.Regexp{regexp.MustCompile(`^\(`)},
		},
		{
			name:    rightParenthesis,
			regexps: []*regexp.Regexp{regexp.MustCompile(`^\)`)},
		},
		{
			name:    and,
			regexps: []*regexp.Regexp{regexp.MustCompile(`(?i)^AND\b`)},
		},
		{
			name:    or,
			regexps: []*regexp.Regexp{regexp.MustCompile(`(?i)^OR\b`)},
		},
		{
			name:    equal,
			regexps: []*regexp.Regexp{regexp.MustCompile(`^==`)},
		},
		{
			name:    notEqual,
			regexps: []*regexp.Regexp{regexp.MustCompile(`^!=`)},
		},
		{
			name:    greaterEqual,
			regexps: []*regexp.Regexp{regexp.MustCompile(`^>=`)},
		},
		{
			name:    greater,
			regexps: []*regexp.Regexp{regexp.MustCompile(`^>`)},
		},
		{
			name:    lessEqual,
			regexps: []*regexp.Regexp{regexp.MustCompile(`^<=`)},
		},
		{
			name:    less,
			regexps: []*regexp.Regexp{regexp.MustCompile(`^<`)},
		},
		{
			name:    identifier,
			regexps: []*regexp.Regexp{regexp.MustCompile(`^\w*`)},
		},
		{
			name:    whitespace,
			regexps: []*regexp.Regexp{regexp.MustCompile(`^\s*`)},
		},
		{
			name:    endOfStatement,
			regexps: []*regexp.Regexp{regexp.MustCompile(`^;`)},
		},
		{
			name:    invalid,
			regexps: []*regexp.Regexp{regexp.MustCompile(`^.*`)},
		},
	}
)

type tokenizer struct {
	query  string
	cursor int
	line   int
	column int
}

func newTokenizer(query string) *tokenizer {
	return &tokenizer{query: query, line: 1, column: 1}
}

func (t *tokenizer) getNextToken() (*token, error) {
	if t.cursor >= len(t.query) {
		return &tokenNoop, nil
	}

	s := t.query[t.cursor:]
	match := ""
	var tk *token

	line, column := t.getLineColumn(0)

	for _, tr := range regexps {
		for _, r := range tr.regexps {
			match = r.FindString(s)
			if match != "" {
				tk = &token{
					_type:    tr.name,
					strValue: match,

					line:   line,
					column: column,
				}
				break
			}
		}
		if match != "" {
			break
		}
	}

	t.cursor += len(match)
	t.line, t.column = t.getLineColumn(len(match))

	if tk == nil {
		return nil, fmt.Errorf("couldn't decipher token %d:%d", line, column)
	}

	if tk._type == "whitespace" {
		return t.getNextToken()
	}

	if tk._type == "string_literal" {
		tk.strValue = tk.strValue[1 : len(tk.strValue)-1]
	} else {
		tk.strValue = strings.ToLower(strings.Join(strings.Fields(tk.strValue), " "))
	}

	v, err := tk.convertToGoType()
	if err != nil {
		return nil, fmt.Errorf("invalid literal '%s' of type '%s' at %d:%d", tk.strValue, tk._type, tk.line, tk.column)
	}

	tk.goValue = v

	if tk._type == "invalid" {
		return nil, fmt.Errorf("error parsing '%s' at %d:%d", tk.strValue, line, column)
	}

	return tk, nil
}

func (t *tokenizer) getLineColumn(skip int) (int, int) {
	skipTotal := t.cursor + skip

	if skipTotal > len(t.query) {
		skipTotal = len(t.query)
	}

	firstHalf := t.query[:skipTotal]

	column := strings.LastIndex(firstHalf, "\n") - len(firstHalf)
	if column > 0 {
		column = 1
	} else {
		column = column * -1
	}

	line := strings.Count(firstHalf, "\n") + 1

	return line, column
}
