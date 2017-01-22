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
	Reader   io.RuneScanner
	Token    Token
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

type LexerError struct {
	Kind     int      // token kind
	Text     string   // message
	Position Position // Occur at
}

func (e *LexerError) Error() string {
	return fmt.Sprintf("lexer error: %s(%s) at line %d",
		tokenstring[e.Kind], e.Text, e.Position)
}

func (lx *Lexer) NewError(kind int, text string) *LexerError {
	return &LexerError{Kind: kind, Text: text, Position: lx.position}
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
	return fmt.Sprintf("%s:%s", tokenstring[t.Kind], t.Text)
}

func (lx *Lexer) ReadRune() (r rune, size int, err error) {
	r, size, err = lx.Reader.ReadRune()
	if IsNewline(r) {
		lx.position.row += 1
	}
	lx.position.total += 1
	return r, size, err
}

func (lx *Lexer) UnreadRune() error {
	err := lx.Reader.UnreadRune()
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
	defer lx.Reader.UnreadRune()
	r, _, err := lx.Reader.ReadRune()
	return r, err
}

// Read while digit
// err is EOF
//  TODO: implement scheme number tower
func (lx *Lexer) ReadNumber(sign rune) (Token, error) {
	is, size, err := lx.ReadWhile(unicode.IsDigit)
	// EOF
	if err != nil && size == 0 {
		return Token{Kind: EOF}, lx.NewError(EOF, "EOF")
	}
	if sign == '-' {
		is = "-" + is
	}

	r, err := lx.PeekRune()
	if err != nil || r != '.' {
		i, err := strconv.Atoi(is)
		if err != nil {
			return Token{Kind: Error}, lx.NewError(Number, "parse number")
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
		return Token{Kind: Error}, lx.NewError(Ident, "illegal identifier")
	}

	// Ignore EOF
	s, _, _ := lx.ReadWhile(IsIdentSubseq)
	s = string(initial) + s
	return Token{Kind: Ident, Text: s, Value: s}, nil
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
	s, size, _ := lx.ReadWhile(func(r rune) bool { return r == '.' })
	if size == 2 {
		return Token{Kind: Ident, Text: "...", Value: "..."}, nil
	} else if size == 0 {
		return Token{Kind: Dot, Text: ".", Value: "."}, nil
	} else {
		return Token{Kind: Error}, lx.NewError(Dot, "illegal dot before "+s)
	}
}

// Read # start token
func (lx *Lexer) ReadSharp() (Token, error) {
	var token Token
	r, _, err := lx.ReadRune()
	if err != nil { // # precede EOF
		return Token{Kind: Error}, lx.NewError(EOF, "nothing after #")
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
		token, err = Token{Kind: Error}, lx.NewError(Error, string(r)+"after #")
	}
	return token, err
}

// e.g. #\a
// not support char name (#\newline, #\space)
func (lx *Lexer) ReadChar() (Token, error) {
	r, _, err := lx.ReadRune()
	return Token{Kind: Char, Text: string(r), Value: r}, err
}

// Read , or ,@
func (lx *Lexer) ReadUnquote() (Token, error) {
	r, err := lx.PeekRune()
	if err != nil {
		return Token{Kind: Unquote, Text: "unquote", Value: "unquote"}, nil
	}
	if r != '@' {
		return Token{Kind: Unquote, Text: "unquote", Value: "unquote"}, nil
	}
	lx.ReadRune()
	return Token{Kind: UnquoteSplicing, Text: "unquote-splicing", Value: "unquote-splicing"}, err
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
			return Token{Kind: Error}, lx.NewError(EOF, "string not finished")
		case r == '"':
			return Token{Kind: String, Text: "\"" + string(rs) + "\"", Value: string(rs)}, nil
		case r == '\\':
			rr, _, _ := lx.ReadRune()
			if !(rr == '"' || rr == '\\') {
				return Token{Kind: Error}, lx.NewError(String, "illegal string element "+string(rr))
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
	return Token{Kind: Comment, Text: s, Value: s}, nil
}

// ReadToken return Token structure
func (lx *Lexer) ReadToken() (Token, error) {
	var err error
	if err = lx.SkipSpaces(); err != nil {
		lx.Token = Token{Kind: EOF, Position: lx.position} // not good... same the end of this function
		return lx.Token, lx.NewError(EOF, "skip spaces")
	}
	// Head of Token is its position
	headPos := lx.position

	// SkipSpaces guarantee rune existence
	r, _, _ := lx.ReadRune()

	switch {
	case unicode.IsDigit(r):
		lx.UnreadRune()
		lx.Token, err = lx.ReadNumber('+')
	case IsIdentInitial(r):
		lx.UnreadRune()
		lx.Token, err = lx.ReadIdent()
	case r == '.':
		lx.Token, err = lx.ReadDot()
	case r == '#':
		lx.Token, err = lx.ReadSharp()
	case r == '"':
		lx.Token, err = lx.ReadString()
	case r == '+' || r == '-':
		if nxt, _ := lx.PeekRune(); unicode.IsDigit(nxt) {
			lx.Token, err = lx.ReadNumber(r)
		} else {
			lx.Token = Token{Kind: Ident, Text: string(r)}
		}
	case r == ';':
		lx.Token, err = lx.ReadComment()
	case r == '(':
		lx.Token = Token{Kind: Open, Text: "("}
	case r == ')':
		lx.Token = Token{Kind: Close, Text: ")"}
	case r == '\'':
		lx.Token = Token{Kind: Quote, Text: "'", Value: "quote"}
	case r == '`':
		lx.Token = Token{Kind: QuasiQuote, Text: "`", Value: "quasiquote"}
	case r == ',':
		lx.Token, err = lx.ReadUnquote()
	default:
		lx.Token = Token{Kind: Error, Text: string(r)}
		err = lx.NewError(Error, "unknown token")
	}
	lx.Token.Position = headPos
	return lx.Token, err
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

// set Lexer's Reader
func (lx *Lexer) SetFile(name string) {
	fp, err := os.Open(name)
	if err != nil {
		panic(err)
	}
	lx.position = Position{filename: name}
	lx.Reader = bufio.NewReader(fp)
}

// set Lexer's Reader
func (lx *Lexer) SetString(s string) {
	lx.position = Position{filename: "<stdin>"}
	lx.Reader = strings.NewReader(s)
}
