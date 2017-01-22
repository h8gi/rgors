package lispy

import (
	"fmt"
)

const (
	ASTSimple = -(iota + 1)
	ASTList
	ASTProgram
)

var aststring = map[int]string{
	ASTSimple:  "Simple",
	ASTList:    "List",
	ASTProgram: "Program",
}

type AST struct {
	Kind     int
	Token    Token
	Children []AST
}

type Parser struct {
	Lexer
}

func (ast AST) String() string {
	if ast.Kind == ASTSimple {
		return fmt.Sprintf("<%s>", ast.Token)
	} else {
		return fmt.Sprintf("<%s:%s>", aststring[ast.Kind], ast.Children)
	}
}

func (ast *AST) push(child AST) {
	ast.Children = append(ast.Children, child)
}

func (p *Parser) match(kind int) error {
	if p.Token.Kind == kind {
		p.ReadToken()
		return nil
	} else {
		return fmt.Errorf("unmatch: %+v, %+v", p.Token, kind)
	}
}

func (p *Parser) Start() error {
	_, err := p.ReadToken()
	return err
}

func (p *Parser) Program() (AST, error) {
	program := AST{Kind: ASTProgram, Children: make([]AST, 0)}
	for {
		if p.Token.Kind == EOF {
			return program, nil
		}
		child, err := p.Datum()
		program.push(child)
		if err != nil {
			return program, err
		}
	}
}

func (p *Parser) Datum() (AST, error) {
	switch p.Token.Kind {
	case Boolean, Number, Char, String, Ident:
		return p.SimpleDatum()
	case Open:
		return p.List()
	case Quote, QuasiQuote, Unquote, UnquoteSplicing:
		return p.Abbrev()
	case EOF:
		return AST{}, fmt.Errorf("datum: illegal EOF")
	default:
		return AST{}, fmt.Errorf("datum: illegal %+v", p.Token)
	}
}

func (p *Parser) SimpleDatum() (AST, error) {
	defer p.ReadToken()
	token := p.Token
	return AST{Kind: ASTSimple, Token: token}, nil
}

// lispy List includes dot list
func (p *Parser) List() (AST, error) {
	// Consume open paren
	p.match(Open)
	list := AST{Kind: ASTList, Children: make([]AST, 0)}
	for {
		switch p.Token.Kind {
		case Close:
			return list, p.match(Close)
		case EOF:
			return list, fmt.Errorf("list: illegal EOF")
		case Dot: // list should be (<datum>+ . <datum>)
			dot := AST{Kind: ASTSimple, Token: p.Token}
			list.push(dot)
			p.match(Dot) // consume dot
			if len(list.Children) < 1 {
				return list, fmt.Errorf("list: illegal Dot")
			}
			lastchild, err := p.Datum()
			list.push(lastchild)
			if err != nil {
				return list, fmt.Errorf("list: illegal datum after dot, %+v", list)
			}
			// Should be closed
			if err := p.match(Close); err != nil {
				return list, fmt.Errorf("list: illegal datum after dot, %+v", list)
			}
			return list, nil
		default:
			child, err := p.Datum()
			list.push(child)
			if err != nil {
				return list, err
			}
		}
	}
}

// 'a `a ,a ,@a
func (p *Parser) Abbrev() (AST, error) {
	head := AST{Kind: ASTSimple, Token: Token{Kind: Ident, Text: tokenstring[p.Token.Kind]}}
	p.match(p.Token.Kind) // Consume abbrev head
	datum, err := p.Datum()
	children := []AST{head, datum}
	return AST{Kind: ASTList, Children: children}, err
}

// utilities
func (p *Parser) ParseFile(name string) (AST, error) {
	p.SetFile(name)
	p.Start()
	return p.Program()
}

func (p *Parser) ParseString(s string) (AST, error) {
	p.SetString(s)
	p.Start()
	return p.Program()
}
