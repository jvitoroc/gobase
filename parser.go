package main

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
)

type tokenRegexps struct {
	name    string
	regexps []*regexp.Regexp
}

var (
	regexps = []*tokenRegexps{
		{
			name:    "keyword",
			regexps: []*regexp.Regexp{regexp.MustCompile(`(?i)^(SELECT|FROM|INSERT\s+INTO|WHERE)\b`)},
		},
		{
			name:    "comma",
			regexps: []*regexp.Regexp{regexp.MustCompile(`^,`)},
		},
		{
			name:    "boolean_literal",
			regexps: []*regexp.Regexp{regexp.MustCompile(`(?i)^(TRUE|FALSE)\b`)},
		},
		{
			name:    "string_literal",
			regexps: []*regexp.Regexp{regexp.MustCompile(`^"([^"]*)"`)},
		},
		{
			name:    "decimal_literal",
			regexps: []*regexp.Regexp{regexp.MustCompile(`^\d+\.?\d+`)},
		},
		{
			name:    "integer_literal",
			regexps: []*regexp.Regexp{regexp.MustCompile(`^\d+`)},
		},
		{
			name:    "left_parenthesis",
			regexps: []*regexp.Regexp{regexp.MustCompile(`^\(`)},
		},
		{
			name:    "right_parenthesis",
			regexps: []*regexp.Regexp{regexp.MustCompile(`^\)`)},
		},
		{
			name:    "and",
			regexps: []*regexp.Regexp{regexp.MustCompile(`(?i)^AND\b`)},
		},
		{
			name:    "or",
			regexps: []*regexp.Regexp{regexp.MustCompile(`(?i)^OR\b`)},
		},
		{
			name:    "equal",
			regexps: []*regexp.Regexp{regexp.MustCompile(`^==`)},
		},
		{
			name:    "not_equal",
			regexps: []*regexp.Regexp{regexp.MustCompile(`^!=`)},
		},
		{
			name:    "greater_equal",
			regexps: []*regexp.Regexp{regexp.MustCompile(`^>=`)},
		},
		{
			name:    "greater",
			regexps: []*regexp.Regexp{regexp.MustCompile(`^>`)},
		},
		{
			name:    "less_equal",
			regexps: []*regexp.Regexp{regexp.MustCompile(`^<=`)},
		},
		{
			name:    "less",
			regexps: []*regexp.Regexp{regexp.MustCompile(`^<`)},
		},
		{
			name:    "name",
			regexps: []*regexp.Regexp{regexp.MustCompile(`^\w*`)},
		},
		{
			name:    "whitespace",
			regexps: []*regexp.Regexp{regexp.MustCompile(`^\s*`)},
		},
		{
			name:    "end_of_statement",
			regexps: []*regexp.Regexp{regexp.MustCompile(`^;`)},
		},
		{
			name:    "invalid",
			regexps: []*regexp.Regexp{regexp.MustCompile(`^.*`)},
		},
	}
)

type expression struct {
	_type string
	value string
	left  *expression
	right *expression
}

type Part struct {
	Keyword string
	Body    any
}

type statement struct {
	Parts []*Part
}

type tokenizer struct {
	query  string
	cursor int
}

type token struct {
	_type string
	value string
}

func (t *tokenizer) getNextToken() *token {
	if t.cursor >= len(t.query) {
		return nil
	}

	s := t.query[t.cursor:]
	match := ""
	var tk *token

	for _, tr := range regexps {
		for _, r := range tr.regexps {
			match = r.FindString(s)
			if match != "" {
				tk = &token{
					_type: tr.name,
					value: match,
				}
				break
			}
		}
		if match != "" {
			break
		}
	}

	t.cursor += len(match)

	if tk == nil {
		return nil
	}

	if tk._type == "whitespace" {
		return t.getNextToken()
	}

	return tk
}

type parser struct {
	t         *tokenizer
	lookahead *token
}

func (p *parser) Parse(q string) (*statement, error) {
	p.t = &tokenizer{query: q}

	p.lookahead = p.t.getNextToken()

	return p.statement()
}

func (p *parser) statement() (*statement, error) {
	s := &statement{
		Parts: []*Part{},
	}
	for {
		if p.lookahead == nil {
			return nil, errors.New("expected end of statement, but got nothing")
		}

		if p.lookahead._type == "end_of_statement" {
			break
		}

		if p.lookahead._type != "keyword" {
			return nil, fmt.Errorf("expected keyword, but got '%s'", p.lookahead.value)
		}

		token := *p.lookahead

		body, err := p.partBody()
		if err != nil {
			return nil, err
		}

		s.Parts = append(s.Parts, &Part{
			Keyword: strings.ToLower(token.value),
			Body:    body,
		})
	}

	return s, nil
}

func (p *parser) partBody() (any, error) {
	switch strings.ToLower(p.lookahead.value) {
	case "select":
		return p.selectBody()
	case "from":
		return p.fromBody()
	case "where":
		return p.whereBody()
	}

	return nil, fmt.Errorf("expected a valid keyword, but got '%s'", p.lookahead.value)
}

func (p *parser) selectBody() (any, error) {
	body := make([]string, 0)
	for {
		p.lookahead = p.t.getNextToken()
		if p.lookahead == nil {
			return nil, errors.New("expected name, but got nothing")
		}

		if p.lookahead._type == "name" {
			body = append(body, p.lookahead.value)
		} else {
			return nil, fmt.Errorf("expected name, but got '%s'", p.lookahead._type)
		}

		p.lookahead = p.t.getNextToken()
		if p.lookahead == nil {
			break
		}

		if p.lookahead._type == "comma" {
			continue
		}

		break
	}

	return body, nil
}

func (p *parser) fromBody() (any, error) {
	p.lookahead = p.t.getNextToken()
	if p.lookahead == nil {
		return nil, errors.New("expected name, but got nothing")
	}

	if p.lookahead._type == "name" {
		value := p.lookahead.value
		p.lookahead = p.t.getNextToken()
		return value, nil
	}

	return nil, fmt.Errorf("expected name, but got '%s'", p.lookahead.value)
}

var allowedPredicateTokens = []string{
	"name",

	"string_literal",
	"integer_literal",
	"decimal_literal",

	"left_parenthesis",
	"right_parenthesis",

	"and",
	"or",

	"equal",
	"not_equal",
	"greater",
	"greater_equal",
	"less",
	"less_equal",
}

func (p *parser) whereBody() (any, error) {
	body := make([]token, 0)
	for {
		p.lookahead = p.t.getNextToken()
		if p.lookahead == nil {
			break
		}

		if !slices.Contains(allowedPredicateTokens, p.lookahead._type) {
			break
		}

		body = append(body, *p.lookahead)
	}

	if len(body) == 0 {
		return nil, errors.New("expected predicate after WHERE, but got nothing")
	}

	if !checkParenthesesBalance(body) {
		return nil, errors.New("expected closing of parenthesis") // fix me: this message is too generic
	}

	if err := checkBooleanExpressionSyntax(body); err != nil {
		return nil, err
	}

	return body, nil
}

func checkParenthesesBalance(tokens []token) bool {
	unclosedParentheses := 0
	for _, t := range tokens {
		if t._type == "left_parenthesis" {
			unclosedParentheses += 1
		} else if t._type == "right_parenthesis" {
			unclosedParentheses -= 1
		}
		if unclosedParentheses < 0 {
			return false
		}
	}

	return unclosedParentheses == 0
}

var (
	operands  = []string{"name", "decimal_literal", "integer_literal", "string_literal", "boolean_literal"}
	operators = []string{"and", "or", "equal", "not_equal", "greater_equal", "greater", "less", "less_equal"}
)

func checkBooleanExpressionSyntax(tokens []token) error {
	if len(tokens) == 0 {
		return errors.New("empty expression")
	}

	if err := checkParenthesesSyntax(tokens); err != nil {
		return err
	}

	var previousToken token
	isPreviousOperand := false
	isPreviousOperator := false

	for i, t := range tokens {
		if t._type == "left_parenthesis" || t._type == "right_parenthesis" {
			continue
		}

		isOperand := slices.Contains(operands, t._type)
		isOperator := slices.Contains(operators, t._type)
		if !isOperand && !isOperator {
			return fmt.Errorf("'%s' is not valid as part of an expression", t.value)
		}

		if i == 0 && isOperator {
			return fmt.Errorf("can't start expression with operator '%s'", t.value)
		}

		if i == len(tokens)-1 && isOperator {
			return fmt.Errorf("can't end expression with an operator '%s'", t.value)
		}

		if isPreviousOperand && isOperand {
			return fmt.Errorf("expected operator after '%s'", previousToken.value)
		}

		if isPreviousOperator && isOperator {
			return fmt.Errorf("expected operand after '%s'", previousToken.value)
		}

		previousToken = t
		isPreviousOperand = isOperand
		isPreviousOperator = isOperator
	}

	return nil
}

func checkParenthesesSyntax(tokens []token) error {
	isPreviousLeftParenthesis := false
	isPreviousRightParenthesis := false

	for i, t := range tokens {
		isOperand := slices.Contains(operands, t._type)
		isOperator := slices.Contains(operators, t._type)
		isLeftParenthesis := t._type == "left_parenthesis"
		isRightParenthesis := t._type == "right_parenthesis"

		if i > 0 {
			if isPreviousLeftParenthesis && isOperator {
				return errors.New("a left parenthesis can't precede an operator")
			}

			if isPreviousRightParenthesis && isOperand {
				return errors.New("a right parenthesis can't precede an operand")
			}

			if isPreviousLeftParenthesis && isRightParenthesis {
				return errors.New("empty parentheses")
			}
		}

		isPreviousLeftParenthesis = isLeftParenthesis
		isPreviousRightParenthesis = isRightParenthesis
	}

	return nil
}
