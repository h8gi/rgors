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
	Kind int
	Text string
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
	UnquoteSplicing
	Dot
)

var tokenstring = map[int]string{
	EOF:             "EOF",
	Comment:         "Comment",
	Error:           "Error",
	Ident:           "Ident",
	Number:          "Number",
	Char:            "Char",
	String:          "String",
	Open:            "Open",
	Close:           "Close",
	OpenVec:         "OpenVec",
	Quote:           "quote",
	QuasiQuote:      "quasiquote",
	Unquote:         "unquote",
	UnquoteSplicing: "unqote-splicing",
	Dot:             "Dot",
}

func (t Token) String() string {
	return fmt.Sprintf("%s: %s", tokenstring[t.Kind], t.Text)
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

// PeekRune
func (lx *Lexer) PeekRune() (rune, error) {
	defer lx.reader.UnreadRune()
	r, _, err := lx.reader.ReadRune()
	return r, err
}

// Read while digit
// err is EOF
//  TODO: implement scheme number tower
func (lx *Lexer) ReadNumber(sign rune) (Token, error) {
	is, size, err := lx.ReadWhile(unicode.IsDigit)
	// EOF
	if err != nil && size == 0 {
		return Token{Kind: EOF}, err
	}
	if sign == '-' {
		is = "-" + is
	}

	r, err := lx.PeekRune()
	if err != nil || r != '.' {
		return Token{Kind: Number, Text: is}, nil
	}
	lx.reader.ReadRune() // consume dot
	fs, size, err := lx.ReadWhile(unicode.IsDigit)
	return Token{Kind: Number, Text: is + "." + fs}, nil
}

// Read Identifier
func (lx *Lexer) ReadIdent() (Token, error) {
	initial, _, err := lx.reader.ReadRune()
	if err != nil {
		return Token{Kind: EOF}, err
	}
	if !IsIdentInitial(initial) {
		return Token{Kind: Error}, fmt.Errorf("lexer: Illegal identifier %s", string(initial))
	}

	// Ignore EOF
	s, _, _ := lx.ReadWhile(IsIdentSubseq)
	s = string(initial) + s
	return Token{Kind: Ident, Text: s}, nil
}

// Ident character
func IsIdentInitial(r rune) bool {
	return strings.ContainsRune("!$%&*/:<=>?^_~", r) || unicode.IsLetter(r)
}
func IsIdentSubseq(r rune) bool {
	return IsIdentInitial(r) || unicode.IsDigit(r) || strings.ContainsRune("+-.@", r)
}

// error: EOF or Illegal dot
func (lx *Lexer) ReadDot() (Token, error) {
	s, size, err := lx.ReadWhile(func(r rune) bool { return r == '.' })
	if size == 2 {
		return Token{Kind: Ident, Text: "..."}, nil
	} else if size == 0 {
		return Token{Kind: Dot, Text: "."}, nil
	} else {
		err = fmt.Errorf("lexer: Illegal dot before %s", s)
		return Token{Kind: Error}, err
	}
}

// Read # start token
func (lx *Lexer) ReadSharp() (Token, error) {
	var token Token
	r, _, err := lx.reader.ReadRune()
	if err != nil { // # precede EOF
		return Token{Kind: Error}, fmt.Errorf("lexer: # precede EOF")
	}
	switch {
	case r == '(':
		token = Token{Kind: OpenVec, Text: "#("}
	case r == 't' || r == 'f':
		token = Token{Kind: Boolean, Text: string([]rune{'#', r})}
	case r == '\\':
		// space will be skipped by readident
		if r, _, err = lx.reader.ReadRune(); unicode.IsSpace(r) {
			token = Token{Kind: Char, Text: string(r)}
			break
		}
		lx.reader.UnreadRune() // recover from space-check.
		token, err = lx.ReadIdent()
		if len(token.Text) == 1 || token.Text == "newline" || token.Text == "space" {
			token.Kind = Char
		} else {
			token.Kind = Error
			err = fmt.Errorf("lexer: #precedes %s", token.Text)
		}
	default:
		token, err = Token{Kind: Error}, fmt.Errorf("lexer: # precede %s", string(r))
	}
	return token, err
}

// Read , or ,@
func (lx *Lexer) ReadUnquote() (Token, error) {
	r, err := lx.PeekRune()
	if err != nil {
		return Token{Kind: Unquote, Text: "unquote"}, nil
	}
	if r != '@' {
		return Token{Kind: Unquote, Text: "unquote"}, nil
	}
	lx.reader.ReadRune()
	return Token{Kind: UnquoteSplicing, Text: "unquote-splicing"}, err
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
			return Token{Kind: EOF}, eof
		case r == '"':
			return Token{Kind: String, Text: string(rs)}, nil
		case r == '\\':
			rr, _, _ := lx.reader.ReadRune()
			if !(rr == '"' || rr == '\\') {
				return Token{Kind: Error}, fmt.Errorf("lexer: Illegal string elm %s", string(rr))
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
	return Token{Kind: Comment, Text: s}, nil
}

// ReadToken return Token structure
func (lx *Lexer) ReadToken() (Token, error) {
	var token Token
	var err error
	if err = lx.SkipSpaces(); err != nil {
		token = Token{Kind: EOF}
		lx.token = token // not good... same the end of this function
		return token, err
	}
	// SkipSpaces guarantee rune existence
	r, _, _ := lx.reader.ReadRune()

	switch {
	case unicode.IsDigit(r):
		lx.reader.UnreadRune()
		token, err = lx.ReadNumber('+')
	case IsIdentInitial(r):
		lx.reader.UnreadRune()
		token, err = lx.ReadIdent()
	case r == '.':
		token, err = lx.ReadDot()
	case r == '#':
		token, err = lx.ReadSharp()
	case r == '"':
		token, err = lx.ReadString()
	case r == '+' || r == '-':
		if nxt, _ := lx.PeekRune(); unicode.IsDigit(nxt) {
			token, err = lx.ReadNumber(r)
		} else {
			token = Token{Kind: Ident, Text: string(r)}
		}
	case r == ';':
		token, err = lx.ReadComment()
	case r == '(':
		token = Token{Kind: Open, Text: "("}
	case r == ')':
		token = Token{Kind: Close, Text: ")"}
	case r == '\'':
		token = Token{Kind: Quote, Text: "'"}
	case r == '`':
		token = Token{Kind: QuasiQuote, Text: "`"}
	case r == ',':
		token, err = lx.ReadUnquote()
	default:
		token = Token{Kind: Error, Text: string(r)}
		err = fmt.Errorf("lexer error: unknown token")
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
		} else if token.Kind == Error {
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
