package sql

import (
	"errors"
	"fmt"

	"github.com/jvitoroc/gobase/eval"
	"github.com/jvitoroc/gobase/schema"
)

type ClauseType string

const (
	Select ClauseType = "select"
	From   ClauseType = "from"
	Where  ClauseType = "where"

	CreateTable ClauseType = "create table"
	Definitions ClauseType = "definitions"

	InsertInto ClauseType = "insert into"
	Values     ClauseType = "values"
)

type Clause struct {
	Type ClauseType
	Body any
}

type Statement struct {
	Clauses []*Clause
}

type parser struct {
	t         *tokenizer
	lookahead token
}

func NewParser(q string) *parser {
	p := &parser{t: newTokenizer(q)}

	return p
}

func (p *parser) Parse() ([]*Statement, error) {
	err := p.moveToNextToken()
	if err != nil {
		return nil, err
	}

	return p.Statements()
}

func (p *parser) Statements() ([]*Statement, error) {
	sts := []*Statement{}

	for {
		s := &Statement{
			Clauses: []*Clause{},
		}

		if p.lookahead == tokenNoop {
			break
		}

		for {
			if p.lookahead == tokenNoop {
				return nil, errors.New("expected end of Statement, but got nothing")
			}

			if p.lookahead._type == endOfStatement {
				err := p.moveToNextToken()
				if err != nil {
					return nil, err
				}

				break
			}

			clause, err := p.clause()
			if err != nil {
				return nil, err
			}

			s.Clauses = append(s.Clauses, clause)
		}

		sts = append(sts, s)
	}

	return sts, nil
}

func (p *parser) clause() (*Clause, error) {
	if p.lookahead._type != clause {
		return nil, fmt.Errorf("expected clause keyword, but got '%s' at %d:%d", p.lookahead.strValue, p.validLine(), p.validColumn())
	}

	tk, err := p.consume()
	if err != nil {
		return nil, err
	}

	clauseType := ClauseType(tk.strValue)

	body, err := p.clauseBody(clauseType, tk)
	if err != nil {
		return nil, err
	}

	return &Clause{
		Type: clauseType,
		Body: body,
	}, nil
}

func (p *parser) clauseBody(_type ClauseType, tk token) (any, error) {
	switch _type {
	case Select:
		return p.selectBody()
	case From:
		return p.identifier()
	case Where:
		return p.whereBody()
	case CreateTable:
		return p.identifier()
	case Definitions:
		return p.definitionsBody()
	case InsertInto:
		return p.identifier()
	case Values:
		return p.valuesBody()
	}

	return nil, fmt.Errorf("clause '%s' not supported at %d:%d", _type, tk.line, tk.column)
}

func (p *parser) selectBody() (any, error) {
	body := make([]*eval.Expression, 0)

	var lastComma token

	for {
		tempTokens := make([]token, 0)
		for {
			if !p.lookahead.isPredicateToken() {
				break
			}

			tempTokens = append(tempTokens, p.lookahead)

			err := p.moveToNextToken()
			if err != nil {
				return nil, err
			}
		}

		if len(tempTokens) == 0 && lastComma != tokenNoop {
			return nil, fmt.Errorf("unexpected comma at %d:%d", lastComma.line, lastComma.column)
		}

		if len(tempTokens) == 0 {
			return nil, fmt.Errorf("invalid eval.Expression at %d:%d", p.validLine(), p.validColumn())
		}

		expr, err := infixToExpressionTree(tempTokens)
		if err != nil {
			return nil, err
		}

		body = append(body, expr)

		if p.lookahead._type == comma {
			var err error
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
	def := []*schema.NewColumn{}

	if !p.lookahead.isLeftParenthesis() {
		return nil, fmt.Errorf("expected opening parenthesis, but got '%s' at %d:%d", p.lookahead.strValue, p.validLine(), p.validColumn())
	}

	_, err := p.consume()
	if err != nil {
		return nil, err
	}

	for {
		if p.lookahead.isRightParenthesis() {
			break
		}

		c := &schema.NewColumn{}

		if p.lookahead._type != identifier {
			return nil, fmt.Errorf("expected column name, but got '%s' at %d:%d", p.lookahead.strValue, p.validLine(), p.validColumn())
		}

		tk, err := p.consume()
		if err != nil {
			return nil, err
		}

		c.Name = tk.strValue

		if p.lookahead._type != dataType {
			return nil, fmt.Errorf("expected column type, but got '%s' at %d:%d", p.lookahead.strValue, p.validLine(), p.validColumn())
		}

		tk, err = p.consume()
		if err != nil {
			return nil, err
		}

		c.Type = schema.ColumnDataType(tk.strValue)

		def = append(def, c)

		if p.lookahead.isRightParenthesis() {
			_, err := p.consume()
			if err != nil {
				return nil, err
			}

			break
		}

		if p.lookahead._type != comma {
			return nil, fmt.Errorf("expected comma, but got '%s' at %d:%d", p.lookahead.strValue, p.validLine(), p.validColumn())
		}

		_, err = p.consume()
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

	if !p.lookahead.isLeftParenthesis() {
		return nil, fmt.Errorf("expected opening parenthesis, but got '%s' at %d:%d", p.lookahead.strValue, p.validLine(), p.validColumn())
	}

	err := p.moveToNextToken()
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

		if p.lookahead._type != comma {
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
		if !p.lookahead.isPredicateToken() {
			break
		}

		tk, err := p.consume()
		if err != nil {
			return nil, err
		}

		body = append(body, tk)
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

	return infixToExpressionTree(body)
}

func (p *parser) identifier() (any, error) {
	if p.lookahead._type != identifier {
		return nil, fmt.Errorf("expected identifier, but got '%s' at %d:%d", p.lookahead._type, p.validLine(), p.validColumn())
	}

	tk, err := p.consume()
	if err != nil {
		return nil, err
	}

	return tk.strValue, nil
}

func infixToExpressionTree(tokens []token) (*eval.Expression, error) {
	t, err := infixToPostfix(tokens)
	if err != nil {
		return nil, err
	}

	return postfixToExpressionTree(t)
}

func infixToPostfix(tokens []token) ([]token, error) {
	if len(tokens) == 0 {
		return nil, errors.New("no tokens given")
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
		} else if tk.isOperator() {
			for tki := s.pop(); tki != tokenNoop; tki = s.pop() {
				if tk.hasLowerOrSamePrecedenceThan(tki) && !tki.isLeftParenthesis() {
					postfix = append(postfix, tki)
					continue
				}
				s.push(tki)
				break
			}
			s.push(tk)
		} else {
			return nil, fmt.Errorf("token '%s' at %d:%d is invalid as part of an expression", tk.strValue, tk.line, tk.column)
		}
	}

	for tki := s.pop(); tki != tokenNoop; tki = s.pop() {
		if !tki.isParenthesis() {
			postfix = append(postfix, tki)
		}
	}

	return postfix, nil
}

func postfixToExpressionTree(tokens []token) (*eval.Expression, error) {
	if len(tokens) == 0 {
		return nil, errors.New("no tokens given")
	}

	s := stack[*eval.Expression]{}

	for _, tk := range tokens {
		if tk.isOperand() {
			expr := &eval.Expression{
				Type:    eval.Operand,
				GoValue: tk.goValue,
			}
			if tk._type == identifier {
				expr.Identifier = tk.strValue
			}
			s.push(expr)
		} else if tk.isOperator() {
			right := s.pop()
			left := s.pop()

			if !eval.IsOperator(string(tk._type)) {
				return nil, fmt.Errorf("token '%s' at %d:%d is not a valid operator", tk.strValue, tk.line, tk.column)
			}

			e := &eval.Expression{
				Type:     eval.Operator,
				Operator: eval.OperatorType(tk._type),
				Left:     left,
				Right:    right,
			}

			s.push(e)
		}
	}

	return s.pop(), nil
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
		return errors.New("empty eval.Expression")
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
			return fmt.Errorf("'%s' at %d:%d is not valid as part of an eval.Expression", t.strValue, t.line, t.column)
		}

		if i == 0 && t.isOperator() {
			return fmt.Errorf("can't start eval.Expression with operator '%s' at %d:%d", t.strValue, t.line, t.column)
		}

		if i == len(tokens)-1 && t.isOperator() {
			return fmt.Errorf("can't end eval.Expression with an operator '%s' at %d:%d", t.strValue, t.line, t.column)
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
