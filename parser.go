package main

import (
	"errors"
	"fmt"
	"regexp"
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
			regexps: []*regexp.Regexp{regexp.MustCompile(`^\d+\.\d+`)},
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
				firstHalf := t.query[:t.cursor]

				column := strings.LastIndex(firstHalf, "\n") - len(firstHalf)
				if column > 0 {
					column = 1
				} else {
					column = column * -1
				}

				tk = &token{
					_type:    tr.name,
					valueStr: match,

					line:   strings.Count(firstHalf, "\n") + 1,
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

	if tk == nil {
		return nil
	}

	if tk._type == "whitespace" {
		return t.getNextToken()
	}

	if tk._type == "string_literal" {
		tk.valueStr = tk.valueStr[1 : len(tk.valueStr)-1]
	} else {
		tk.valueStr = strings.ToLower(tk.valueStr)
	}

	return tk
}

type parser struct {
	t         *tokenizer
	lookahead *token
}

func (p *parser) parse(q string) (*statement, error) {
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
			return nil, fmt.Errorf("expected keyword, but got '%s' at %d:%d", p.lookahead.valueStr, p.lookahead.line, p.lookahead.column)
		}

		token := *p.lookahead

		body, err := p.partBody()
		if err != nil {
			return nil, err
		}

		s.Parts = append(s.Parts, &Part{
			Keyword: token.valueStr,
			Body:    body,
		})
	}

	return s, nil
}

func (p *parser) partBody() (any, error) {
	switch p.lookahead.valueStr {
	case "select":
		return p.selectBody()
	case "from":
		return p.fromBody()
	case "where":
		return p.whereBody()
	}

	return nil, fmt.Errorf("expected a valid keyword, but got '%s' at %d:%d", p.lookahead.valueStr, p.lookahead.line, p.lookahead.column)
}

func (p *parser) selectBody() (any, error) {
	body := make([]string, 0)
	for {
		p.lookahead = p.t.getNextToken()
		if p.lookahead == nil {
			return nil, errors.New("expected name, but got nothing")
		}

		if p.lookahead._type == "name" {
			body = append(body, p.lookahead.valueStr)
		} else {
			return nil, fmt.Errorf("expected name, but got '%s' at %d:%d", p.lookahead._type, p.lookahead.line, p.lookahead.column)
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
		value := p.lookahead.valueStr
		p.lookahead = p.t.getNextToken()
		return value, nil
	}

	return nil, fmt.Errorf("expected name, but got '%s' at %d:%d", p.lookahead.valueStr, p.lookahead.line, p.lookahead.column)
}

func (p *parser) whereBody() (any, error) {
	body := make([]token, 0)
	for {
		p.lookahead = p.t.getNextToken()
		if p.lookahead == nil {
			break
		}

		if !p.lookahead.isPredicateToken() {
			break
		}

		body = append(body, *p.lookahead)
	}

	if len(body) == 0 {
		return nil, errors.New("expected predicate after WHERE, but got nothing")
	}

	if err := checkParenthesesBalance(body); err != nil {
		return nil, err
	}

	if err := checkBooleanExpressionSyntax(body); err != nil {
		return nil, err
	}

	return infixToPostfix(body), nil
}

func infixToPostfix(tokens []token) []token {
	if len(tokens) == 0 {
		return tokens
	}

	s := stack[token]{}
	postfix := make([]token, 0, len(tokens))

	for _, tk := range tokens {
		if tk.isLeftParenthesis() {
			s.push(tk)
		} else if tk.isRightParenthesis() {
			for tki := s.pop(); tki != tokenNoop; tki = s.pop() {
				if tki.isLeftParenthesis() {
					break
				}
				postfix = append(postfix, tki)
			}
		} else if tk.isOperand() {
			postfix = append(postfix, tk)
		} else {
			for tki := s.pop(); tki != tokenNoop; tki = s.pop() {
				if tk.hasLowerOrSamePrecedenceThan(tki) && !tki.isLeftParenthesis() {
					postfix = append(postfix, tki)
					continue
				}
				s.push(tki)
				break
			}
			s.push(tk)
		}
	}

	for tki := s.pop(); tki != tokenNoop; tki = s.pop() {
		if !tki.isParenthesis() {
			postfix = append(postfix, tki)
		}
	}

	return postfix
}

func checkParenthesesBalance(tokens []token) error {
	unclosedParentheses := stack[token]{}
	for _, t := range tokens {
		if t.isLeftParenthesis() {
			unclosedParentheses.push(t)
		} else if t.isRightParenthesis() {
			tk := unclosedParentheses.pop()
			if tk == tokenNoop {
				return fmt.Errorf("unexpected closing parenthesis at %d:%d", t.line, t.column)
			}
		}
	}

	if len(unclosedParentheses) > 0 {
		tk := unclosedParentheses.pop()
		return fmt.Errorf("opening parenthesis at %d:%d, but missing its closing parenthesis", tk.line, tk.column)
	}

	return nil
}

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
		if t.isParenthesis() {
			continue
		}

		isOperand := t.isOperand()
		isOperator := t.isComparisonOperator() || t.isLogicalOperator()
		if !isOperand && !isOperator {
			return fmt.Errorf("'%s' at %d:%d is not valid as part of an expression", t.valueStr, t.line, t.column)
		}

		if i == 0 && isOperator {
			return fmt.Errorf("can't start expression with operator '%s' at %d:%d", t.valueStr, t.line, t.column)
		}

		if i == len(tokens)-1 && isOperator {
			return fmt.Errorf("can't end expression with an operator '%s' at %d:%d", t.valueStr, t.line, t.column)
		}

		if isPreviousOperand && isOperand {
			return fmt.Errorf("expected operator after '%s' at %d:%d", previousToken.valueStr, t.line, t.column)
		}

		if isPreviousOperator && isOperator {
			return fmt.Errorf("expected operand after '%s' at %d:%d", previousToken.valueStr, t.line, t.column)
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
		isOperand := t.isOperand()
		isOperator := t.isOperator()
		isLeftParenthesis := t.isLeftParenthesis()
		isRightParenthesis := t.isRightParenthesis()

		if i > 0 {
			if isPreviousLeftParenthesis && isOperator {
				return fmt.Errorf("an operator is not allowed to be positioned at %d:%d after a opening parenthesis", t.line, t.column)
			}

			if isPreviousRightParenthesis && isOperand {
				return fmt.Errorf("an operand is not allowed to be positioned at %d:%d after a closing parenthesis", t.line, t.column)
			}

			if isPreviousLeftParenthesis && isRightParenthesis {
				return fmt.Errorf("empty parentheses at %d:%d", t.line, t.column)
			}
		}

		isPreviousLeftParenthesis = isLeftParenthesis
		isPreviousRightParenthesis = isRightParenthesis
	}

	return nil
}

func (p *parser) mustTokenize() []token {
	tokens := make([]token, 0)
	for {
		p.lookahead = p.t.getNextToken()
		if p.lookahead == nil {
			break
		}

		tokens = append(tokens, *p.lookahead)
	}

	return tokens
}
