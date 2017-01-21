package lispy

import (
	"container/list"
	"fmt"
)

const (
	ASTSimple = -(iota + 1)
	ASTList
)

// https://en.wikipedia.org/wiki/Recursive_descent_parser
type AST struct {
	kind     int
	value    Token
	children []AST
}

type Parser struct {
	lx     *Lexer
	result []AST
}

func (p *Parser) isMatch(kind int) bool {
	if p.lx.token.kind == kind {
		return true
	}
	return false
}

func (p *Parser) next() error {
	tkn, err := p.lx.ReadToken()
	return err
}

func (p *Parser) push(ast AST) {
	p.result = append(p.result, ast)
}

func (list *AST) push(a AST) {
	list.children = append(list.children, a)
}

func (p *Parser) expect(kind int) error {
	if p.accept(kind) {
		return nil
	}
	return fmt.Errorf("expect: unexpected token %+v", p.lx.token)
}

func (p *Parser) Number() error {
	switch {
	case p.isMatch(Number):
		p.push(AST{kind: ASTSimple, value: p.lx.token})
		p.lx.ReadToken()
		return nil
	default:
		return fmt.Errorf("expected number...")
	}
}

func (p *Parser) Symbol() error {
	switch {
	case p.isMatch(Ident):
		p.push(AST{kind: ASTSimple, value: p.lx.token})
		p.lx.ReadToken()
		return nil
	default:
		return fmt.Errorf("expected symbol...")
	}
}

func (p *Parser) List() error {

	switch {
	case p.accept(Open):
		list = AST{kind: ASTList}
		for {
			if p.accept(Close) {
				p.push(list)
				return nil
			}
			err := p.Expr()
			if err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("expected open...")
	}
}

func (p *Parser) Expr() error {
	if err := p.Number(); err == nil {
		return nil
	}
	if err := p.Symbol(); err == nil {
		return nil
	}
	if err := p.List(); err == nil {
		return nil
	}
	return fmt.Errorf("expected Expr...")
}

func (p *Parser) Lispy() error {
	p.lx.ReadToken()
	for {
		if p.accept(EOF) {
			return nil
		}
		err := p.Expr()
		if err != nil {
			return err
		}
	}
}

func (p *Parser) SetLexer(lx *Lexer) {
	p.lx = lx
}
