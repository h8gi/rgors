package lispy

import (
	"fmt"
)

const (
	ASTSimple = -(iota + 1)
	ASTPair
	ASTNil
)

var aststring = map[int]string{
	ASTSimple: "Simple",
	ASTPair:   "Pair",
	ASTNil:    "()",
}

type AST struct {
	Kind  int
	Token Token
	Car   *AST
	Cdr   *AST
}

type Parser struct {
	Lexer
}

func (ast AST) String() string {
	switch ast.Kind {
	case ASTSimple:
		return fmt.Sprintf("<%s>", ast.Token)
	case ASTPair:
		return fmt.Sprintf("(%s . %s)", ast.Car, ast.Cdr)
	case ASTNil:
		return fmt.Sprintf("%s", aststring[ASTNil])
	default:
		return fmt.Sprintf("???")
	}
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

// Program is list of AST
// Top level
func (p *Parser) Program() ([]AST, error) {
	program := make([]AST, 0)
	for {
		if p.Token.Kind == EOF {
			return program, nil
		}
		child, err := p.Datum()
		program = append(program, child)
		if err != nil {
			return program, err
		}
	}
}

// one sexpression
func (p *Parser) Datum() (AST, error) {
	switch p.Token.Kind {
	case Boolean, Number, Char, String, Ident:
		return p.SimpleDatum()
	case Open:
		// return p.List()
		return p.Pair()
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

func (p *Parser) Pair() (AST, error) {
	// Consume open paren
	p.match(Open)
	// var pair = AST{Kind: ASTPair}
	// var err error

	var innerPair func() (AST, error)
	innerPair = func() (AST, error) {
		var car, cdr AST
		var pair = AST{Kind: ASTPair}
		var err error

		// read car
		switch p.Token.Kind {
		case EOF, Dot:
			p.match(Dot)
			return pair, fmt.Errorf("pair: illegal token, %+v", p.Token)
		case Close:
			return AST{Kind: ASTNil}, err
		default:
			car, err = p.Datum()
			pair.Car = &car
			if err != nil {
				return pair, err
			}
		}
		// read cdr
		switch p.Token.Kind {
		case Dot: // (car . cdr)
			err = p.match(Dot) // consume dot
			if err != nil {
				return pair, fmt.Errorf("pair: %s", err.Error())
			}
			cdr, err = p.Datum() // read cdr
			pair.Cdr = &cdr
			if err != nil {
				return pair, err
			}

			if err = p.match(Close); err != nil {
				err = fmt.Errorf("pair: %s", err.Error())
			}
			return pair, err
		default: // (a b ...)
			cdr, err = innerPair()
			pair.Cdr = &cdr
			return pair, err
		}
	}
	return innerPair()
}

// 'a `a ,a ,@a
func (p *Parser) Abbrev() (AST, error) {
	pair := AST{Kind: ASTPair}
	car := AST{Kind: ASTSimple, Token: Token{Kind: Ident, Text: tokenstring[p.Token.Kind]}}
	p.match(p.Token.Kind) // Consume abbrev car
	cdr, err := p.Datum()
	pair.Car = &car
	pair.Cdr = &AST{Kind: ASTPair, Car: &cdr, Cdr: &AST{Kind: ASTNil}}
	return pair, err
}

// utilities
func (p *Parser) ParseFile(name string) ([]AST, error) {
	p.SetFile(name)
	p.Start()
	return p.Program()
}

func (p *Parser) ParseString(s string) ([]AST, error) {
	p.SetString(s)
	p.Start()
	return p.Program()
}
