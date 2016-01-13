package scanner

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/wellington/sass/token"
)

func IsSymbol(r rune) bool {
	return strings.ContainsRune("(),;{}#:", r)
}

func isSpace(r rune) bool {
	return unicode.IsSpace(r)
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch >= 0x80 && unicode.IsLetter(ch)
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9' || ch >= 0x80 && unicode.IsDigit(ch)
}

const (
	symbols = `.*-_&>-:$%,+~[]=()`
)

func isAllowedRune(r rune) bool {
	return unicode.IsNumber(r) ||
		unicode.IsLetter(r) ||
		strings.ContainsRune(symbols, r)
}

// An ErrorHandler may be provided to Scanner.Init. If a syntax error is
// encountered and a handler was installed, the handler is called with a
// position and an error message. The position points to the beginning of
// the offending token.
//
type ErrorHandler func(pos token.Position, msg string)

type prefetch struct {
	pos token.Pos
	tok token.Token
	lit string
}

type Scanner struct {
	src    []byte
	ch     rune
	offset int

	// hack use a channel as a FIFO queue, this is probably a terrible idea
	queue chan prefetch

	mode Mode

	// Many things in Sass change on left or right side of colon
	// rhs will track which side of the colon we are in.
	rhs bool
	// Track whether we are inside function params. If so, treat everything
	// as whitespace delimited
	inParams bool

	file       *token.File
	dir        string
	err        ErrorHandler
	ErrorCount int
	rdOffset   int
	lineOffset int
}

// Mode controls scanner behavior
type Mode uint

const (
	ScanComments Mode = 1 << iota // return comments during Scan
)

func (s *Scanner) Init(file *token.File, src []byte, err ErrorHandler, mode Mode) {
	// Explicitly initialize all fields since a scanner may be reused.
	if file.Size() != len(src) {
		panic(fmt.Sprintf("file size (%d) does not match src len (%d)", file.Size(), len(src)))
	}
	s.file = file
	s.dir, _ = filepath.Split(file.Name())
	s.src = src
	s.err = err
	s.mode = mode
	// There should never more than 2 in the queue, but buffer to 10
	// just to be safe
	s.queue = make(chan prefetch, 10)
	s.rhs = false

	s.ch = ' '
	s.offset = 0
	s.rdOffset = 0
	s.lineOffset = 0
	s.ErrorCount = 0

	s.next()
	// if s.ch == bom {
	// 	s.next() // ignore BOM at file beginning
	// }

}

// rewind resets the scanner to a previous position
// DANGER: You should only use this if you know what you are doing.
//
// When rewinding newlines, it will cause incorrect line numbers
// to be reported.
func (s *Scanner) rewind(offs int) {
	s.rdOffset = offs
	s.next()
}

// backup one rune
func (s *Scanner) backup() {
	w := utf8.RuneLen(s.ch)
	s.rdOffset -= w

	// Copy of slice, this is expensive
	r, w := utf8.DecodeLastRune(s.src[:s.rdOffset])
	s.offset = s.rdOffset - w
	s.ch = r
}

func (s *Scanner) next() {

	if s.rdOffset < len(s.src) {
		s.offset = s.rdOffset

		if s.ch == '\n' {
			s.lineOffset = s.offset
			s.file.AddLine(s.offset)
		}
		r, w := rune(s.src[s.rdOffset]), 1
		switch {
		case r == 0:
			s.error(s.offset, "illegal character NUL")
		case r >= 0x80:
			// not ASCII
			r, w = utf8.DecodeRune(s.src[s.rdOffset:])
			if r == utf8.RuneError && w == 1 {
				s.error(s.offset, "illegal UTF-8 encoding")
				// } else if r == bom && s.offset > 0 {
				// 	s.error(s.offset, "illegal byte order mark")
			}
		}
		s.rdOffset += w
		s.ch = r
	} else {

		s.offset = len(s.src)
		if s.ch == '\n' {
			s.lineOffset = s.offset
			s.file.AddLine(s.offset)
		}

		s.ch = -1 // eof
	}

}

type Item struct {
	Type  token.Token
	Pos   int
	Value string
}

func (s *Scanner) skipWhitespace() {
	for s.ch == ' ' || s.ch == '\t' || s.ch == '\n' || s.ch == '\r' {
		s.next()
	}
}

// Scan should differentiate between these cases
// selector[,#=.+>] {}
// :foo(ol) {}
// [hey = 'ho'], a > b {}
// c [hoo *= "ha" ] {}
// div,, , span, ,, {}
// a + b, c {}
// d e, f ~ g + h, > i {}
//
// reference parent selector: &
// function: @function h() {}
// return: @return function-exists();
// mixin: @mixin($var) {}
// call or conversion: abs(-5);
// $variable: $substitution
// rule: value;
// with #{} found anywhere in them
// directives: @import
// math 1 + 3 or (1 + 3)
// New strategy, scan until something important is encountered
func (s *Scanner) Scan() (pos token.Pos, tok token.Token, lit string) {
	defer func() {
		fmt.Printf("scan tok: %s lit: '%s'\n", tok, lit)
	}()
	// Check the queue, which may contain tokens that were fetched
	// in a previous scan while determing ambiguious tokens.
	select {
	case pre := <-s.queue:
		pos, tok, lit = pre.pos, pre.tok, pre.lit
		return
	default:
		// If the queue is empty, do nothing
	}

	s.skipWhitespace()
	pos = s.file.Pos(s.offset)
	offs := s.offset
	ch := s.ch
scanAgain:
	switch {
	case ch == '&':
		fallthrough
	case ch == '[':
		fallthrough
	case ch == '.':
		fallthrough
	case isLetter(ch):
		// Scan until encountering {};
		// selector: { termination
		// rule:  IDENT followed by : it must then be followed by ; or }
		// value: same as above but after the colon followed by ; or }
		pos, tok, lit = s.scanDelim(s.offset)
	case '0' <= ch && ch <= '9':
		// This can not be a selector
		tok, lit = s.scanNumber(false)
		utok, ulit := s.scanUnit()
		if utok != token.ILLEGAL {
			tok = utok
			lit = lit + ulit
		}
	}

	if tok != token.ILLEGAL {
		return
	}

	// move forward
	s.next()
	switch ch {
	case -1:
		// Text expects EOF to be empty string
		lit = ""
		tok = token.EOF
	case '$':
		lit = s.scanText(s.offset-1, 0, false)
		tok = token.VAR
	case '#':
		// # can be one of three things
		// color:    #fff[000]
		// selector: #a
		// interp:   #{}
		if s.ch == '{' {
			tok, lit = s.scanInterp(offs)

		}
		pos, tok, lit = s.scanDelim(offs)
	case ':':
		if isLetter(s.ch) {
			pos, tok, lit = s.scanDelim(offs)
		} else {
			// s.rhs = true
			tok = token.COLON
		}
	case '-':
		if isLetter(s.ch) {
			pos, tok, lit = s.scanDelim(offs)
		} else {
			tok = token.SUB
		}
	case '\'':
		lit = s.scanText(s.offset-1, '\'', true)
		tok = token.QSSTRING
	case '"':
		lit = s.scanText(s.offset-1, '"', true)
		tok = token.QSTRING
	case '.':
		if '0' <= s.ch && s.ch <= '9' {
			tok, lit = s.scanNumber(true)
			utok, ulit := s.scanUnit()
			if utok != token.ILLEGAL {
				tok = utok
				lit = lit + ulit
			}
		} else {
			tok = token.PERIOD
		}
	case '/':
		if s.ch == '/' || s.ch == '*' {
			comment := s.scanComment()
			if s.mode&ScanComments == 0 {
				goto scanAgain
			}
			tok = token.COMMENT
			lit = comment
		} else {
			tok = token.QUO
		}
	case '@':
		tok, lit = s.scanDirective()
	case '^':
		tok = token.XOR
	// case '#':
	// 	tok, lit = s.scanColor()
	case '&':
		tok = token.AND
	case '<':
		tok = s.switch2(token.LSS, token.LEQ)
	case '>':
		tok = s.switch2(token.GTR, token.GEQ)
	case '=':
		tok = s.switch2(token.ASSIGN, token.EQL)
	case '!':
		tok = s.switch2(token.NOT, token.NEQ)
	case ',':
		tok = token.COMMA
	case ';':
		// s.rhs = false
		tok = token.SEMICOLON
		lit = ";"
	case '(':
		s.inParams = true
		tok = token.LPAREN
	case ')':
		s.inParams = false
		tok = token.RPAREN
	case '[':
		tok = token.LBRACK
	case ']':
		tok = token.RBRACK
	case '{':
		tok = token.LBRACE
	case '}':
		tok = token.RBRACE
	case '%':
		tok = token.REM
	case '+':
		tok = token.ADD
	case '*':
		tok = token.MUL
	default:
		fmt.Printf("Illegal %q\n", ch)
	}

	// item = Item{Type: ItemILLEGAL, Value: string(ch)}
	return
}

func isText(ch rune, whitespace bool) bool {

	switch {
	case
		isLetter(ch), isDigit(ch),
		ch == '.', ch == '/':
		return true
	case (ch == '\'' || ch == '"'):
		return true
	case whitespace && isSpace(ch):
		return true
	}
	return false
}

// Special parsing of tokens while inside params to account for different
// whitespace handling rules.
func (s *Scanner) scanParams() string {
	return ""
}

var colondelim = []byte(":")

// scanDelim looks through ambiguous text (selectors, rules, functions)
// and returns a properly parsed set.
//
// a#id { // 'a#id'
// { color: blue; } // 'color' ':' 'blue'
func (s *Scanner) scanDelim(offs int) (pos token.Pos, tok token.Token, lit string) {

	pos = s.file.Pos(offs)
	var loop int
	for !strings.ContainsRune(";{}()", s.ch) && s.ch != -1 {
		loop++
		if loop > 10 {
			fmt.Println("loop detected:", string(s.ch))
			return pos, token.ILLEGAL, string(s.src[offs:s.offset])
		}

		// runes not supported by scanText
		if isAllowedRune(s.ch) {
			s.next()
		}
		s.scanText(offs, 0, true)
	}

	sel := s.src[offs:s.offset]
	ch := s.ch
	switch ch {
	case '(':
		tok = token.FUNC
		// return pos, token.FUNC, string(bytes.TrimSpace(sel))
	case '{':
		return pos, token.SELECTOR,
			string(bytes.TrimSpace(sel))
	default:
		tok = token.VALUE
	}

	// Do we have a rule or a value?
	parts := bytes.Split(sel, colondelim)
	first := parts[0]
	l := len(first)
	tok = token.VALUE
	lit = string(sel)

	if len(parts) > 1 {
		tok = token.RULE
		lit = string(bytes.TrimSpace(first))
		s.queue <- prefetch{
			pos: s.file.Pos(offs + l),
			tok: token.COLON,
		}

		// Strip leading space to find start of value
		trim := bytes.TrimLeftFunc(parts[1], isSpace)

		s.queue <- prefetch{
			pos: s.file.Pos(offs + l + 1 + len(parts[1]) - len(trim)),
			tok: tok,
			lit: string(bytes.TrimSpace(parts[1])),
		}
	}

	return
}

// ScanText is responsible for gobbling non-whitespace characters
//
// This should validate variable naming http://stackoverflow.com/a/17194994
// a-zA-Z0-9_-
// Also these if escaped with \ !"#$%&'()*+,./:;<=>?@[]^{|}~
func (s *Scanner) scanText(offs int, end rune, whitespace bool) string {
	// offs := s.offset - 1 // catch first quote

	var ch rune
	for s.ch == '\\' || isText(s.ch, whitespace) ||
		s.ch == '#' ||
		// For now, just eat interpolations
		s.ch == end {
		ch = s.ch
		s.next()

		// For now, just eat interpolation
		if ch == '#' {
			// Verify introlation
			s.scanInterp(s.offset)
		}

		if ch == '\\' {
			if strings.ContainsRune(`!"#$%&'()*+,./:;<=>?@[]^{|}~`, s.ch) {
				s.next()
			} else {
				s.error(s.offset, "attempted to escape invalid character "+string(s.ch))
			}
		}

		if ch == end {
			break
		}
	}

	// eat the end character
	if end != 0 && ch != end {
		s.error(s.offset, "expected end of "+string(end))
	}

	ss := string(s.src[offs:s.offset])
	return ss
}

func (s *Scanner) scanRGB(pos int) (tok token.Token, lit string) {
	tok = token.COLOR
	offs := pos

	if s.ch != '(' {
		lit = string(s.src[offs:s.offset])
		s.error(offs, "invalid rgb (: "+lit)
	}

	// s.next()
	// ttok, num := s.scanNumber(false)
	// if ttok != token.INT {
	// 	s.error(s.offset, "invalid rgb int: "+num)
	// }

	// for i := 0; i < 2; i++ {
	// 	if s.ch != ',' {
	// 		s.error(s.offset, "invalid rgb ,: "+string(s.ch))
	// 	}
	// 	s.next()
	// 	s.skipWhitespace()
	// 	tok, num := s.scanNumber(false)
	// 	if tok != token.INT {
	// 		s.error(s.offset, "invalid rgb int: "+num)
	// 	}
	// }

	for s.ch != ')' && s.ch != ';' {
		s.next()
	}
	if s.ch == ';' {
		s.error(offs, "invalid rgb: "+string(s.src[offs:s.offset]))
	}
	s.next()

	lit = string(s.src[offs:s.offset])
	return
}

func (s *Scanner) scanInterp(offs int) (token.Token, string) {
	// Should only be called after '#' is detected
	if s.ch != '{' {
		return token.ILLEGAL, ""
	}
	for s.ch != '}' {
		s.next()
	}
	s.next()
	return token.INTERP, string(s.src[offs:s.offset])

}

func (s *Scanner) scanColor() (tok token.Token, lit string) {
	offs := s.offset - 1
	if s.ch == '{' {
		return s.scanInterp(offs)
	}
	for isLetter(s.ch) || isDigit(s.ch) {
		s.next()
	}
	lit = string(s.src[offs:s.offset])
	if len(lit) > 1 {
		return token.COLOR, lit
	}
	return token.ILLEGAL, lit
}

// ScanDirective matches Sass directives http://sass-lang.com/documentation/file.SASS_REFERENCE.html#directives
func (s *Scanner) scanDirective() (tok token.Token, lit string) {
	offs := s.offset - 1
	for isLetter(s.ch) || s.ch == '-' {
		s.next()
	}
	lit = string(s.src[offs:s.offset])
	switch lit {
	case "@import":
		tok = token.IMPORT
	case "@media":
		tok = token.MEDIA
	case "@mixin":
		tok = token.MIXIN
	case "@extend":
		tok = token.EXTEND
	case "@at-root":
		tok = token.ATROOT
	case "@include":
		tok = token.INCLUDE
	case "@debug":
		tok = token.DEBUG
	case "@warn":
		tok = token.WARN
	case "@error":
		tok = token.ERROR
	}

	return
}

// func (s *Scanner) scanRule() (token.Token, string) {
// 	offs := s.offset
// scanRule:
// 	for isLetter(s.ch) || isDigit(s.ch) || s.ch == '-' {
// 		s.next()
// 	}
// 	if s.ch == '#' {
// 		s.next()
// 		tok, lit := s.scanInterp(s.offset)
// 		if tok == token.ILLEGAL {
// 			// Interpolation failed, bailout
// 		}
// 	}
// 	ss := string(s.src[offs:s.offset])
// 	return ss
// }

func (s *Scanner) scanIdent() string {
	offs := s.offset
	for isLetter(s.ch) || isDigit(s.ch) {
		s.next()
	}
	ss := string(s.src[offs:s.offset])
	return ss
}

func (s *Scanner) scanUnit() (token.Token, string) {
	offs := s.offset
	switch s.ch {
	case 'p':
		// pt px
		s.next()
		if s.ch == 'x' {
			s.next()
			return token.UPX, string(s.src[offs:s.offset])
		} else if s.ch == 't' {
			s.next()
			return token.UPT, string(s.src[offs:s.offset])
		}
	case '%':
		s.next()
		return token.UPCT, "%"
	}
	return token.ILLEGAL, ""
}

func (s *Scanner) scanNumber(seenDecimalPoint bool) (token.Token, string) {
	// digitVal(s.ch) < 10
	offs := s.offset
	tok := token.INT

	if seenDecimalPoint {
		offs--
		tok = token.FLOAT
		s.scanMantissa(10)
		goto exponent
	}

	if s.ch == '0' {
		// int or float
		offs := s.offset
		s.next()
		if s.ch == 'x' || s.ch == 'X' {
			// hexadecimal int
			s.next()
			s.scanMantissa(16)
			if s.offset-offs <= 2 {
				// only scanned "0x" or "0X"
				s.error(offs, "illegal hexadecimal number")
			}
		} else {
			// octal int or float
			seenDecimalDigit := false
			s.scanMantissa(8)
			if s.ch == '8' || s.ch == '9' {
				// illegal octal int or float
				seenDecimalDigit = true
				s.scanMantissa(10)
			}
			if s.ch == '.' || s.ch == 'e' || s.ch == 'E' || s.ch == 'i' {
				goto fraction
			}
			// octal int
			if seenDecimalDigit {
				s.error(offs, "illegal octal number")
			}
		}
		goto exit
	}

	// decimal int or float
	s.scanMantissa(10)

fraction:
	if s.ch == '.' {
		tok = token.FLOAT
		s.next()
		s.scanMantissa(10)
	}

exponent:
	if s.ch == 'e' || s.ch == 'E' {
		tok = token.FLOAT
		s.next()
		if s.ch == '-' || s.ch == '+' {
			s.next()
		}
		s.scanMantissa(10)
	}

	if s.ch == 'i' {
		tok = token.ILLEGAL
		s.next()
	}

exit:
	return tok, string(s.src[offs:s.offset])

}

func (s *Scanner) scanMantissa(base int) {
	for digitVal(s.ch) < base {
		s.next()
	}
}

func digitVal(ch rune) int {
	switch {
	case '0' <= ch && ch <= '9':
		return int(ch - '0')
	case 'a' <= ch && ch <= 'f':
		return int(ch - 'a' + 10)
	case 'A' <= ch && ch <= 'F':
		return int(ch - 'A' + 10)
	}
	return 16 // larger than any legal digit val
}

func (s *Scanner) scanComment() string {
	// initial '/' already consumed; s.ch == '/' || s.ch == '*'
	offs := s.offset - 1 // position of initial '/'
	hasCR := false

	if s.ch == '/' {
		//-style comment
		s.next()
		for s.ch != '\n' && s.ch >= 0 {
			if s.ch == '\r' {
				hasCR = true
			}
			s.next()
		}
		goto exit
	}

	/*-style comment */
	s.next()
	for s.ch >= 0 {
		ch := s.ch
		if ch == '\r' {
			hasCR = true
		}
		s.next()
		if ch == '*' && s.ch == '/' {
			s.next()
			goto exit
		}
	}
	s.error(offs, "comment not terminated")

exit:
	lit := s.src[offs:s.offset]
	if hasCR {
		lit = stripCR(lit)
	}

	return string(lit)
}

func (s *Scanner) error(offs int, msg string) {
	if s.err != nil {
		s.err(s.file.Position(s.file.Pos(offs)), msg)
	}
	s.ErrorCount++
}

func stripCR(b []byte) []byte {
	c := make([]byte, len(b))
	i := 0
	for _, ch := range b {
		if ch != '\r' {
			c[i] = ch
			i++
		}
	}
	return c[:i]
}

func (s *Scanner) switch2(tok0, tok1 token.Token) token.Token {
	if s.ch == '=' {
		s.next()
		return tok1
	}
	return tok0
}
