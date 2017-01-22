package lispy

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"
)

// Lexer
type Lexer struct {
	reader   io.RuneScanner
	token    Token
	position Position
}

// Token
type Token struct {
	Kind     int
	Text     string
	Position Position
	Value    interface{}
}

type Position struct {
	filename string
	row      int
	total    int
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

func (lx *Lexer) ReadRune() (r rune, size int, err error) {
	r, size, err = lx.reader.ReadRune()
	if IsNewline(r) {
		lx.position.row += 1
	}
	lx.position.total += 1
	return r, size, err
}

func (lx *Lexer) UnreadRune() error {
	err := lx.reader.UnreadRune()
	if err != nil {
		return err
	}
	r, err := lx.PeekRune()
	if IsNewline(r) {
		lx.position.row -= 1
	}
	lx.position.total -= 1
	return err
}

func IsNewline(r rune) bool {
	return r == '\n'
}

// SkipSpaces consume spaces
// error is io.EOF
func (lx *Lexer) SkipSpaces() error {
	for {
		r, _, err := lx.ReadRune()
		// EOF check
		if err != nil {
			return err
		}
		if !unicode.IsSpace(r) {
			lx.UnreadRune()
			return nil
		}
	}
}

// Read while pred(r) return true
// error is io.EOF
func (lx *Lexer) ReadWhile(pred func(rune) bool) (s string, size int, err error) {
	rs := make([]rune, 0)
	for {
		r, _, eof := lx.ReadRune()
		// EOF check
		if eof != nil {
			return string(rs), len(rs), eof
		}
		// pred fail
		if !pred(r) {
			lx.UnreadRune()
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
		i, err := strconv.Atoi(is)
		if err != nil {
			return Token{Kind: Error}, err
		}
		return Token{Kind: Number, Text: is, Value: i}, nil
	}
	lx.ReadRune() // consume dot
	fs, size, err := lx.ReadWhile(unicode.IsDigit)
	fs = is + "." + fs
	f, err := strconv.ParseFloat(fs, 64)
	return Token{Kind: Number, Text: fs, Value: f}, err
}

// Read Identifier
func (lx *Lexer) ReadIdent() (Token, error) {
	initial, _, err := lx.ReadRune()
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
	r, _, err := lx.ReadRune()
	if err != nil { // # precede EOF
		return Token{Kind: Error}, fmt.Errorf("lexer: # precede EOF")
	}
	switch r {
	case '(': // Vector open
		token = Token{Kind: OpenVec, Text: "#("}
	case 't':
		token = Token{Kind: Boolean, Text: "#t", Value: true}
	case 'f':
		token = Token{Kind: Boolean, Text: "#f", Value: false}
	case '\\': // Char
		token, err = lx.ReadChar()
	default:
		token, err = Token{Kind: Error}, fmt.Errorf("lexer: # precede %s", string(r))
	}
	return token, err
}

// e.g. #\a
func (lx *Lexer) ReadChar() (Token, error) {
	var token Token
	var err error
	// space will be skipped by readident
	if r, _, err := lx.ReadRune(); unicode.IsSpace(r) {
		token = Token{Kind: Char, Text: string(r), Value: r}
		return token, err
	}
	lx.UnreadRune() // recover from space-check.
	token, err = lx.ReadIdent()
	if len(token.Text) == 1 {
		token = Token{Kind: Char, Text: token.Text, Value: token.Text[0]}
	} else if token.Text == "newline" {
		token = Token{Kind: Char, Text: token.Text, Value: '\n'}
	} else if token.Text == "space" {
		token = Token{Kind: Char, Text: token.Text, Value: ' '}
	} else {
		token.Kind = Error
		err = fmt.Errorf("lexer: #precedes %s", token.Text)
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
	lx.ReadRune()
	return Token{Kind: UnquoteSplicing, Text: "unquote-splicing"}, err
}

// Read scheme string
//  first double quote has already been read.
func (lx *Lexer) ReadString() (Token, error) {
	rs := make([]rune, 0)
	for {
		r, _, eof := lx.ReadRune()
		// EOF check
		switch {
		case eof != nil:
			return Token{Kind: EOF}, eof
		case r == '"':
			return Token{Kind: String, Text: "\"" + string(rs) + "\"", Value: string(rs)}, nil
		case r == '\\':
			rr, _, _ := lx.ReadRune()
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
	var err error
	if err = lx.SkipSpaces(); err != nil {
		lx.token = Token{Kind: EOF, Position: lx.position} // not good... same the end of this function
		return lx.token, err
	}
	// Head of Token is its position
	headPos := lx.position

	// SkipSpaces guarantee rune existence
	r, _, _ := lx.ReadRune()

	switch {
	case unicode.IsDigit(r):
		lx.UnreadRune()
		lx.token, err = lx.ReadNumber('+')
	case IsIdentInitial(r):
		lx.UnreadRune()
		lx.token, err = lx.ReadIdent()
	case r == '.':
		lx.token, err = lx.ReadDot()
	case r == '#':
		lx.token, err = lx.ReadSharp()
	case r == '"':
		lx.token, err = lx.ReadString()
	case r == '+' || r == '-':
		if nxt, _ := lx.PeekRune(); unicode.IsDigit(nxt) {
			lx.token, err = lx.ReadNumber(r)
		} else {
			lx.token = Token{Kind: Ident, Text: string(r)}
		}
	case r == ';':
		lx.token, err = lx.ReadComment()
	case r == '(':
		lx.token = Token{Kind: Open, Text: "("}
	case r == ')':
		lx.token = Token{Kind: Close, Text: ")"}
	case r == '\'':
		lx.token = Token{Kind: Quote, Text: "'"}
	case r == '`':
		lx.token = Token{Kind: QuasiQuote, Text: "`"}
	case r == ',':
		lx.token, err = lx.ReadUnquote()
	default:
		lx.token = Token{Kind: Error, Text: string(r)}
		err = fmt.Errorf("lexer error: unknown token")
	}
	lx.token.Position = headPos
	return lx.token, err
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
	lx.position = Position{filename: name}
	lx.reader = bufio.NewReader(fp)
}

// set Lexer's reader
func (lx *Lexer) SetString(s string) {
	lx.position = Position{filename: "<stdin>"}
	lx.reader = strings.NewReader(s)
}

func (lx *Lexer) Token() Token {
	return lx.token
}
