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
			regexps: []*regexp.Regexp{regexp.MustCompile(`(?i)^(SELECT|FROM|(INSERT\s+INTO)|WHERE|(CREATE\s+TABLE)|DEFINITIONS|VALUES)\b`)},
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
			name:    "number_literal",
			regexps: []*regexp.Regexp{regexp.MustCompile(`^\d+(\.\d+)?`)},
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

func (t *tokenizer) getNextToken() (*token, error) {
	if t.cursor >= len(t.query) {
		return &tokenNoop, nil
	}

	s := t.query[t.cursor:]
	match := ""
	var tk *token

	firstHalf := t.query[:t.cursor]

	column := strings.LastIndex(firstHalf, "\n") - len(firstHalf)
	if column > 0 {
		column = 1
	} else {
		column = column * -1
	}

	line := strings.Count(firstHalf, "\n") + 1

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

type parser struct {
	t         *tokenizer
	lookahead token
}

func (p *parser) parse(q string) ([]*statement, error) {
	p.t = &tokenizer{query: q}

	err := p.moveToNextToken()
	if err != nil {
		return nil, err
	}

	return p.statements()
}

func (p *parser) statements() ([]*statement, error) {
	sts := []*statement{}

	for {
		s := &statement{
			Parts: []*Part{},
		}

		if p.lookahead == tokenNoop {
			break
		}

		for {
			if p.lookahead == tokenNoop {
				return nil, errors.New("expected end of statement, but got nothing")
			}

			if p.lookahead._type == "end_of_statement" {
				_, _ = p.consume()
				break
			}

			if p.lookahead._type != "keyword" {
				return nil, fmt.Errorf("expected keyword, but got '%s' at %d:%d", p.lookahead.strValue, p.lookahead.line, p.lookahead.column)
			}

			currentPart := &Part{
				Keyword: p.lookahead.strValue,
			}

			body, err := p.partBody()
			if err != nil {
				return nil, err
			}

			currentPart.Body = body

			s.Parts = append(s.Parts, currentPart)
		}

		sts = append(sts, s)
	}

	return sts, nil
}

func (p *parser) moveToNextToken() error {
	tk, err := p.t.getNextToken()
	if err != nil {
		return err
	}

	p.lookahead = *tk
	return nil
}

func (p *parser) consume() (token, error) {
	t := p.lookahead

	err := p.moveToNextToken()
	if err != nil {
		return tokenNoop, err
	}

	return t, nil
}

func (p *parser) partBody() (any, error) {
	switch p.lookahead.strValue {
	case "select":
		return p.selectBody()
	case "from":
		return p.fromBody()
	case "where":
		return p.whereBody()
	case "create table":
		return p.createTable()
	case "definitions":
		return p.definitions()
	case "insert into":
		return p.insertInto()
	case "values":
		return p.values()
	}

	return nil, fmt.Errorf("expected a valid keyword, but got '%s' at %d:%d", p.lookahead.strValue, p.lookahead.line, p.lookahead.column)
}

func (p *parser) selectBody() (any, error) {
	body := make([]*expression, 0)
	for {
		err := p.moveToNextToken()
		if err != nil {
			return nil, err
		}

		tempTokens := make([]token, 0)
		for {
			if p.lookahead.isPredicateToken() {
				tempTokens = append(tempTokens, p.lookahead)
			} else {
				break
			}

			err = p.moveToNextToken()
			if err != nil {
				return nil, err
			}
		}

		body = append(body, infixToExpressionTree(tempTokens))

		if p.lookahead._type == "comma" {
			continue
		}

		break
	}

	return body, nil
}

func (p *parser) createTable() (any, error) {
	err := p.moveToNextToken()
	if err != nil {
		return nil, err
	}

	if p.lookahead._type != "name" {
		return nil, errors.New("expected name")
	}

	tk, err := p.consume()
	if err != nil {
		return nil, err
	}

	return tk.strValue, nil
}

func (p *parser) definitions() (any, error) {
	def := []*newColumn{}

	err := p.moveToNextToken()
	if err != nil {
		return nil, err
	}

	if !p.lookahead.isLeftParenthesis() {
		return nil, errors.New("expected opening parenthesis")
	}

	_, err = p.consume()
	if err != nil {
		return nil, err
	}

	for {
		if p.lookahead.isRightParenthesis() {
			break
		}

		c := &newColumn{}

		if p.lookahead._type == "name" {
			tk, err := p.consume()
			if err != nil {
				return nil, err
			}

			c.name = tk.strValue
		}

		if p.lookahead._type == "name" {
			tk, err := p.consume()
			if err != nil {
				return nil, err
			}

			c._type = columnType(tk.strValue)
		}

		def = append(def, c)

		if p.lookahead.isRightParenthesis() {
			_, err := p.consume()
			if err != nil {
				return nil, err
			}

			break
		}

		if p.lookahead._type != "comma" {
			return nil, fmt.Errorf("expected comma, but got '%s'", p.lookahead.strValue)
		}

		_, err := p.consume()
		if err != nil {
			return nil, err
		}
	}

	return def, nil
}

func (p *parser) values() (any, error) {
	values := []string{}

	err := p.moveToNextToken()
	if err != nil {
		return nil, err
	}

	if !p.lookahead.isLeftParenthesis() {
		return nil, errors.New("expected opening parenthesis")
	}

	_, err = p.consume()
	if err != nil {
		return nil, err
	}

	for {
		if p.lookahead.isRightParenthesis() {
			break
		}

		if !p.lookahead.isLiteral() {
			return nil, fmt.Errorf("expected literal, but got '%s'", p.lookahead.strValue)
		}

		tk, err := p.consume()
		if err != nil {
			return nil, err
		}

		values = append(values, tk.strValue)

		if p.lookahead.isRightParenthesis() {
			_, err := p.consume()
			if err != nil {
				return nil, err
			}

			break
		}

		if p.lookahead._type != "comma" {
			return nil, fmt.Errorf("expected comma, but got '%s'", p.lookahead.strValue)
		}

		_, err = p.consume()
		if err != nil {
			return nil, err
		}
	}

	return values, nil
}

func (p *parser) fromBody() (any, error) {
	err := p.moveToNextToken()
	if err != nil {
		return nil, err
	}

	if p.lookahead._type != "name" {
		return nil, fmt.Errorf("expected name, but got '%s' at %d:%d", p.lookahead.strValue, p.lookahead.line, p.lookahead.column)
	}

	value := p.lookahead.strValue
	err = p.moveToNextToken()
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (p *parser) insertInto() (any, error) {
	err := p.moveToNextToken()
	if err != nil {
		return nil, err
	}

	if p.lookahead._type != "name" {
		return nil, fmt.Errorf("expected name, but got '%s' at %d:%d", p.lookahead.strValue, p.lookahead.line, p.lookahead.column)
	}

	value := p.lookahead.strValue
	err = p.moveToNextToken()
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (p *parser) whereBody() (any, error) {
	body := make([]token, 0)
	for {
		err := p.moveToNextToken()
		if err != nil {
			return nil, err
		}

		if !p.lookahead.isPredicateToken() {
			break
		}

		body = append(body, p.lookahead)
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

	return infixToExpressionTree(body), nil
}

func infixToExpressionTree(tokens []token) *expression {
	return postfixToExpressionTree(infixToPostfix(tokens))
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

func postfixToExpressionTree(tokens []token) *expression {
	if len(tokens) == 0 {
		return nil
	}

	s := stack[*expression]{}

	for _, tk := range tokens {
		if tk.isOperand() {
			s.push(&expression{
				_type:     Operand,
				goValue:   tk.goValue,
				strValue:  tk.strValue,
				valueType: tk._type,
			})
		} else if tk.isOperator() {
			right := s.pop()
			left := s.pop()

			e := &expression{
				_type:    Operator,
				operator: tk._type,
				left:     left,
				right:    right,
			}

			s.push(e)
		}
	}

	return s.pop()
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

		if !t.isPredicateToken() {
			return fmt.Errorf("'%s' at %d:%d is not valid as part of an expression", t.strValue, t.line, t.column)
		}

		if i == 0 && t.isOperator() {
			return fmt.Errorf("can't start expression with operator '%s' at %d:%d", t.strValue, t.line, t.column)
		}

		if i == len(tokens)-1 && t.isOperator() {
			return fmt.Errorf("can't end expression with an operator '%s' at %d:%d", t.strValue, t.line, t.column)
		}

		if isPreviousOperand && t.isOperand() {
			return fmt.Errorf("expected operator after '%s' at %d:%d", previousToken.strValue, t.line, t.column)
		}

		if isPreviousOperator && t.isOperator() {
			return fmt.Errorf("expected operand after '%s' at %d:%d", previousToken.strValue, t.line, t.column)
		}

		previousToken = t
		isPreviousOperand = t.isOperand()
		isPreviousOperator = t.isOperator()
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
		err := p.moveToNextToken()
		if err != nil {
			panic(err)
		}

		if p.lookahead == tokenNoop {
			break
		}

		tokens = append(tokens, p.lookahead)
	}

	return tokens
}
