package main

import (
	"errors"
	"fmt"
)

type Clause struct {
	Name string
	Body any
}

type statement struct {
	Clauses []*Clause
}

type parser struct {
	t         *tokenizer
	lookahead token
}

func newParser(q string) *parser {
	p := &parser{t: newTokenizer(q)}

	return p
}

func (p *parser) parse() ([]*statement, error) {
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
			Clauses: []*Clause{},
		}

		if p.lookahead == tokenNoop {
			break
		}

		for {
			if p.lookahead == tokenNoop {
				return nil, errors.New("expected end of statement, but got nothing")
			}

			if p.lookahead._type == "end_of_statement" {
				err := p.moveToNextToken()
				if err != nil {
					return nil, err
				}

				break
			}

			if p.lookahead._type != "clause" {
				return nil, fmt.Errorf("expected keyword, but got '%s' at %d:%d", p.lookahead.strValue, p.validLine(), p.validColumn())
			}

			currentPart := &Clause{
				Name: p.lookahead.strValue,
			}

			body, err := p.partBody()
			if err != nil {
				return nil, err
			}

			currentPart.Body = body

			s.Clauses = append(s.Clauses, currentPart)
		}

		sts = append(sts, s)
	}

	return sts, nil
}

func (p *parser) partBody() (any, error) {
	switch p.lookahead.strValue {
	case "select":
		return p.selectBody()
	case "from":
		return p.name()
	case "where":
		return p.whereBody()
	case "create table":
		return p.name()
	case "definitions":
		return p.definitionsBody()
	case "insert into":
		return p.name()
	case "values":
		return p.valuesBody()
	}

	return nil, fmt.Errorf("expected a valid keyword, but got '%s' at %d:%d", p.lookahead.strValue, p.validLine(), p.validColumn())
}

func (p *parser) selectBody() (any, error) {
	body := make([]*expression, 0)

	err := p.moveToNextToken()
	if err != nil {
		return nil, err
	}

	var lastComma token

	for {
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

		if len(tempTokens) == 0 && lastComma != tokenNoop {
			return nil, fmt.Errorf("unexpected comma at %d:%d", lastComma.line, lastComma.column)
		}

		if len(tempTokens) == 0 {
			return nil, fmt.Errorf("invalid expression at %d:%d", p.validLine(), p.validColumn())
		}

		body = append(body, infixToExpressionTree(tempTokens))

		if p.lookahead._type == "comma" {
			lastComma, err = p.consume()
			if err != nil {
				return nil, err
			}

			continue
		}

		break
	}

	return body, nil
}

func (p *parser) definitionsBody() (any, error) {
	def := []*newColumn{}

	err := p.moveToNextToken()
	if err != nil {
		return nil, err
	}

	if !p.lookahead.isLeftParenthesis() {
		return nil, fmt.Errorf("expected opening parenthesis, but got '%s' at %d:%d", p.lookahead.strValue, p.validLine(), p.validColumn())
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
		} else {
			return nil, fmt.Errorf("expected column name, but got '%s' at %d:%d", p.lookahead.strValue, p.validLine(), p.validColumn())
		}

		if p.lookahead._type == "data_type" {
			tk, err := p.consume()
			if err != nil {
				return nil, err
			}

			c._type = columnType(tk.strValue)
		} else {
			return nil, fmt.Errorf("expected column type, but got '%s' at %d:%d", p.lookahead.strValue, p.validLine(), p.validColumn())
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
			return nil, fmt.Errorf("expected comma, but got '%s' at %d:%d", p.lookahead.strValue, p.validLine(), p.validColumn())
		}

		_, err := p.consume()
		if err != nil {
			return nil, err
		}
	}

	if len(def) == 0 {
		return nil, fmt.Errorf("definitions cannot be empty near %d:%d", p.validLine(), p.validColumn())
	}

	return def, nil
}

func (p *parser) valuesBody() (any, error) {
	values := []string{}

	err := p.moveToNextToken()
	if err != nil {
		return nil, err
	}

	if !p.lookahead.isLeftParenthesis() {
		return nil, fmt.Errorf("expected opening parenthesis, but got '%s' at %d:%d", p.lookahead.strValue, p.validLine(), p.validColumn())
	}

	err = p.moveToNextToken()
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
			err := p.moveToNextToken()
			if err != nil {
				return nil, err
			}

			break
		}

		if p.lookahead._type != "comma" {
			return nil, fmt.Errorf("expected comma, but got '%s' at %d:%d", p.lookahead.strValue, p.validLine(), p.validColumn())
		}

		err = p.moveToNextToken()
		if err != nil {
			return nil, err
		}
	}

	if len(values) == 0 {
		return nil, fmt.Errorf("must provide values at %d:%d", p.validLine(), p.validColumn())
	}

	return values, nil
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
		return nil, errors.New("expected predicate after 'WHERE', but got nothing")
	}

	if err := checkParenthesesBalance(body); err != nil {
		return nil, err
	}

	if err := checkBooleanExpressionSyntax(body); err != nil {
		return nil, err
	}

	return infixToExpressionTree(body), nil
}

func (p *parser) name() (any, error) {
	err := p.moveToNextToken()
	if err != nil {
		return nil, err
	}

	if p.lookahead._type != "name" {
		return nil, fmt.Errorf("expected name, but got '%s' at %d:%d", p.lookahead._type, p.validLine(), p.validColumn())
	}

	tk, err := p.consume()
	if err != nil {
		return nil, err
	}

	return tk.strValue, nil
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
				return fmt.Errorf("an operator is not allowed to be positioned at %d:%d after an opening parenthesis", t.line, t.column)
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

func (p *parser) validLine() int {
	if p.lookahead.line > 0 {
		return p.lookahead.line
	}

	return p.t.line
}

func (p *parser) validColumn() int {
	if p.lookahead.column > 0 {
		return p.lookahead.column
	}

	return p.t.column
}

func mustTokenize(input string) []token {
	p := parser{t: newTokenizer(input)}
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
