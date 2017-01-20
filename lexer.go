package lispy

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

// Lexer
type Lexer struct {
	reader io.RuneScanner
	token  Token
}

// Token
type Token struct {
	kind int
	text string
}

const (
	EOF = -(iota + 1)
	Error
	Comment
	Ident
	Boolean
	Number
	Char
	String
	Open
	Close
	OpenVec
	Quote
	QuasiQuote
	Unquote
	UnqoteAt
	Dot
)

var tokenstring = map[int]string{
	EOF:        "EOF",
	Comment:    "Comment",
	Error:      "Error",
	Ident:      "Ident",
	Number:     "Number",
	Char:       "Char",
	String:     "String",
	Open:       "Open",
	Close:      "Close",
	OpenVec:    "OpenVec",
	Quote:      "Quote",
	QuasiQuote: "QuasiQuote",
	Unquote:    "Unquote",
	UnqoteAt:   "UnqoteAt",
	Dot:        "Dot",
}

func (t Token) String() string {
	return fmt.Sprintf("%s: %s", tokenstring[t.kind], t.text)
}

// SkipSpaces consume spaces
// error is io.EOF
func (lx *Lexer) SkipSpaces() error {
	for {
		r, _, err := lx.reader.ReadRune()
		// EOF check
		if err != nil {
			return err
		}
		if !unicode.IsSpace(r) {
			lx.reader.UnreadRune()
			return nil
		}
	}
}

// Read while pred(r) return true
// error is io.EOF
func (lx *Lexer) ReadWhile(pred func(rune) bool) (s string, size int, err error) {
	rs := make([]rune, 0)
	for {
		r, _, eof := lx.reader.ReadRune()
		// EOF check
		if eof != nil {
			return string(rs), len(rs), eof
		}
		// pred fail
		if !pred(r) {
			lx.reader.UnreadRune()
			return string(rs), len(rs), nil
		}
		rs = append(rs, r)
	}
}

// Read while digit
// err is EOF
//  TODO: implement scheme number tower
func (lx *Lexer) ReadNumber() (Token, error) {
	s, size, err := lx.ReadWhile(unicode.IsDigit)
	// EOF
	if err != nil && size == 0 {
		return Token{kind: EOF}, err
	}

	return Token{kind: Number, text: s}, nil
}

// Read Identifier
func (lx *Lexer) ReadIdent() (Token, error) {
	head, _, err := lx.reader.ReadRune()
	if err != nil {
		return Token{kind: EOF}, err
	}
	if !IsIdentHead(head) {
		return Token{kind: Error}, fmt.Errorf("lexer: Illegal identifier %s", string(head))
	}

	// Ignore EOF
	s, _, _ := lx.ReadWhile(IsIdentSucc)
	s = string(head) + s
	return Token{kind: Ident, text: s}, nil
}

// Ident character
func IsIdentHead(r rune) bool {
	return strings.ContainsRune("!$%&*/:<=>?^_~", r) || unicode.IsLetter(r)
}
func IsIdentSucc(r rune) bool {
	return IsIdentHead(r) || unicode.IsDigit(r) || strings.ContainsRune("+-.@", r)
}

// error: EOF or Illegal dot
func (lx *Lexer) ReadDot() (Token, error) {
	s, size, err := lx.ReadWhile(func(r rune) bool { return r == '.' })
	if size == 2 {
		return Token{kind: Ident, text: "..."}, nil
	} else if size == 0 {
		return Token{kind: Dot, text: "."}, nil
	} else {
		err = fmt.Errorf("lexer: Illegal dot before %s", s)
		return Token{kind: Error}, err
	}
}

// Read # start token
func (lx *Lexer) ReadSharp() (Token, error) {
	var token Token
	r, _, err := lx.reader.ReadRune()
	if err != nil { // # precede EOF
		return Token{kind: Error}, fmt.Errorf("lexer: # precede EOF")
	}
	switch {
	case r == '(':
		token = Token{kind: OpenVec, text: "#("}
	case r == 't' || r == 'f':
		token = Token{kind: Boolean, text: string([]rune{'#', r})}
	case r == '\\':
		// space will be skipped by readident
		if r, _, err = lx.reader.ReadRune(); unicode.IsSpace(r) {
			token = Token{kind: Char, text: string(r)}
			break
		}
		lx.reader.UnreadRune() // recover from space-check.
		token, err = lx.ReadIdent()
		if len(token.text) == 1 || token.text == "newline" || token.text == "space" {
			token.kind = Char
		} else {
			token.kind = Error
			err = fmt.Errorf("lexer: #precedes %s", token.text)
		}
	default:
		token, err = Token{kind: Error}, fmt.Errorf("lexer: # precede %s", string(r))
	}
	return token, err
}

// Read scheme string
//  first double quote has already been read.
func (lx *Lexer) ReadString() (Token, error) {
	rs := make([]rune, 0)
	for {
		r, _, eof := lx.reader.ReadRune()
		// EOF check
		switch {
		case eof != nil:
			return Token{kind: EOF}, eof
		case r == '"':
			return Token{kind: String, text: string(rs)}, nil
		case r == '\\':
			rr, _, _ := lx.reader.ReadRune()
			if !(rr == '"' || rr == '\\') {
				return Token{kind: Error}, fmt.Errorf("lexer: Illegal string elm %s", string(rr))
			} else {
				rs = append(rs, r, rr)
			}
		default:
			rs = append(rs, r)
		}
	}
}

// Read comment
func (lx *Lexer) ReadComment() (Token, error) {
	s, _, _ := lx.ReadWhile(func(r rune) bool { return r != '\n' })
	return Token{kind: Comment, text: s}, nil
}

// ReadToken return Token structure
func (lx *Lexer) ReadToken() (Token, error) {

	var token Token
	var err error
	if err = lx.SkipSpaces(); err != nil {
		return Token{kind: EOF}, err
	}
	// SkipSpaces guarantee rune existence
	r, _, _ := lx.reader.ReadRune()

	switch {
	case unicode.IsDigit(r):
		lx.reader.UnreadRune()
		token, err = lx.ReadNumber()
	case IsIdentHead(r):
		lx.reader.UnreadRune()
		token, err = lx.ReadIdent()
	case r == '.':
		token, err = lx.ReadDot()
	case r == '#':
		token, err = lx.ReadSharp()
	case r == '"':
		token, err = lx.ReadString()
	case r == '+' || r == '-':
		token = Token{kind: Ident, text: string(r)}
	case r == ';':
		token, err = lx.ReadComment()
	case r == '(':
		token = Token{kind: Open, text: "("}
	case r == ')':
		token = Token{kind: Close, text: ")"}
	case r == '\'':
		token = Token{kind: Quote, text: "'"}
	case r == '`':
		token = Token{kind: QuasiQuote, text: "`"}
	default:
		token = Token{kind: Error, text: string(r)}
	}
	lx.token = token
	return token, err
}

// return token slice
func (lx *Lexer) ReadTokens() ([]Token, error) {
	var tokens []Token
	for {
		token, err := lx.ReadToken()
		if err != nil {
			return append(tokens, token), err
		} else if token.kind == Error {
			return append(tokens, token), err
		}
		tokens = append(tokens, token)
	}
}

// set Lexer's reader
func (lx *Lexer) SetFile(name string) {
	fp, err := os.Open(name)
	if err != nil {
		panic(err)
	}
	lx.reader = bufio.NewReader(fp)
}

// set Lexer's reader
func (lx *Lexer) SetString(s string) {
	lx.reader = strings.NewReader(s)
}

func (lx *Lexer) Token() Token {
	return lx.token
}
