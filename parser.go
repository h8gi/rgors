package lispy

import (
	"fmt"
)

const (
	ASTSimple = -(iota + 1)
	ASTList
)

var aststring = map[int]string{
	ASTSimple: "Simple",
	ASTList:   "List",
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
		return fmt.Sprintf("<List:%s>", ast.Children)
	}

}

func (ast *AST) push(child AST) {
	ast.Children = append(ast.Children, child)
}

func (lx *Lexer) match(kind int) error {
	if lx.Token.Kind == kind {
		lx.ReadToken()
		return nil
	} else {
		return fmt.Errorf("unmatch: %+v, %+v", lx.Token, kind)
	}
}

func (lx *Lexer) Datum() (AST, error) {
	switch lx.Token.Kind {
	case Boolean, Number, Char, String, Ident:
		return lx.SimpleDatum()
	case Open:
		return lx.List()
	case Quote, QuasiQuote, Unquote, UnquoteSplicing:
		return lx.Abbrev()
	case EOF:
		return AST{}, fmt.Errorf("datum: illegal EOF")
	default:
		return AST{}, fmt.Errorf("datum: illegal %+v", lx.Token)
	}
}

func (lx *Lexer) SimpleDatum() (AST, error) {
	defer lx.ReadToken()
	token := lx.Token
	return AST{Kind: ASTSimple, Token: token}, nil
}

// lispy List includes dot list
func (lx *Lexer) List() (AST, error) {
	// Consume open paren
	lx.match(Open)
	list := AST{Kind: ASTList, Children: make([]AST, 0)}
	for {
		switch lx.Token.Kind {
		case Close:
			return list, lx.match(Close)
		case EOF:
			return list, fmt.Errorf("list: illegal EOF")
		case Dot: // list should be (<datum>+ . <datum>)
			dot := AST{Kind: ASTSimple, Token: lx.Token}
			list.push(dot)
			lx.match(Dot) // consume dot
			if len(list.Children) < 1 {
				return list, fmt.Errorf("list: illegal Dot")
			}
			lastchild, err := lx.Datum()
			list.push(lastchild)
			if err != nil {
				return list, fmt.Errorf("list: illegal datum after dot, %+v", list)
			}
			// Should be closed
			if err := lx.match(Close); err != nil {
				return list, fmt.Errorf("list: illegal datum after dot, %+v", list)
			}
			return list, nil
		default:
			child, err := lx.Datum()
			list.push(child)
			if err != nil {
				return list, err
			}
		}
	}
}

// 'a `a ,a ,@a
func (lx *Lexer) Abbrev() (AST, error) {
	head := AST{Kind: ASTSimple, Token: Token{Kind: Ident, Text: tokenstring[lx.Token.Kind]}}
	lx.match(lx.Token.Kind) // Consume abbrev head
	datum, err := lx.Datum()
	children := []AST{head, datum}
	return AST{Kind: ASTList, Children: children}, err
}
