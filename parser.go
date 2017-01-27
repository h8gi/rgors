package lispy

import (
	"fmt"
)

type Parser struct {
	Lexer
}

type UnclosedError struct {
	Text string
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
func (p *Parser) Program() ([]LObj, error) {
	program := make([]LObj, 0)
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
func (p *Parser) Datum() (LObj, error) {
	switch p.Token.Kind {
	case Boolean, Number, Char, String, Ident:
		return p.SimpleDatum()
	case Open:
		p.match(Open) // consume open
		return p.Pair()
	case OpenVec:
		p.match(OpenVec) // consume openvec
		return p.Vector()
	case Quote, QuasiQuote, Unquote, UnquoteSplicing:
		return p.Abbrev()
	case Comment:
		p.match(Comment)
		return p.Datum()
	case EOF:
		return LObj{}, fmt.Errorf("datum: illegal EOF")
	default:
		return LObj{}, fmt.Errorf("datum: illegal %+v", p.Token)
	}
}

func (p *Parser) SimpleDatum() (LObj, error) {
	var obj LObj
	switch p.Token.Kind {
	case Boolean:
		obj = LObj{Type: LispBoolean, Value: p.Token.Value}
	case Number:
		obj = LObj{Type: LispNumber, Value: p.Token.Value}
	case Char:
		obj = LObj{Type: LispChar, Value: p.Token.Value}
	case String:
		obj = LObj{Type: LispString, Value: p.Token.Value}
	default:
		obj = LObj{Type: LispSymbol, Value: p.Token.Value}

	}
	p.ReadToken()
	return obj, nil
}

func (p *Parser) Vector() (LObj, error) {
	var vec = make([]LObj, 0)
	for {
		switch p.Token.Kind {
		case Close:
			p.match(Close)
			return LObj{Type: LispVector, Value: vec}, nil
		case EOF:
			return LObj{}, &UnclosedError{Text: "vector"}
		default:
			elem, err := p.Datum()
			if err != nil {
				return elem, err
			}
			vec = append(vec, elem)
		}
	}
}

func (p *Parser) Pair() (LObj, error) {
	var car, cdr LObj
	var pair = LObj{Type: LispPair}
	var err error

	// read car
	switch p.Token.Kind {
	case Dot:
		p.match(Dot)
		return pair, fmt.Errorf("pair: illegal Dot")
	case EOF:
		return pair, &UnclosedError{Text: ")"}
	case Close:
		p.match(Close)
		return LObj{Type: LispNull}, err
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
		p.match(Dot) // consume dot
		if p.Token.Kind == EOF {
			return pair, &UnclosedError{Text: ")"}
		}
		cdr, err = p.Datum() // read cdr
		pair.Cdr = &cdr
		if err != nil {
			return pair, err
		}

		if err = p.match(Close); err != nil {
			if p.Token.Kind == EOF {
				err = &UnclosedError{Text: ")"}
			} else {
				err = fmt.Errorf("missing close paren \")\"")
			}
		}
		return pair, err
	default: // (a b ...)
		cdr, err = p.Pair()
		pair.Cdr = &cdr
		return pair, err
	}
}

// 'a `a ,a ,@a
func (p *Parser) Abbrev() (LObj, error) {
	pair := LObj{Type: LispPair}
	car := LObj{Type: LispSymbol, Value: p.Token.Value}
	p.match(p.Token.Kind) // Consume abbrev car
	cdr, err := p.Datum()
	pair.Car = &car
	pair.Cdr = &LObj{Type: LispPair, Car: &cdr, Cdr: &LObj{Type: LispNull}}
	return pair, err
}

// utilities
func (p *Parser) ParseFile(name string) ([]LObj, error) {
	p.SetFile(name)
	p.Start()
	return p.Program()
}

func (p *Parser) ParseString(s string) ([]LObj, error) {
	p.SetString(s)
	p.Start()
	return p.Program()
}
