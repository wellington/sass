package scanner

import (
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
	symbols = `/\.*-_`
)

func isAllowedRune(r rune) bool {
	return unicode.IsNumber(r) ||
		unicode.IsLetter(r) ||
		strings.ContainsRune(symbols, r)
}

var eof = rune(0)

// An ErrorHandler may be provided to Scanner.Init. If a syntax error is
// encountered and a handler was installed, the handler is called with a
// position and an error message. The position points to the beginning of
// the offending token.
//
type ErrorHandler func(pos token.Position, msg string)

type Scanner struct {
	src    []byte
	ch     rune
	offset int

	mode Mode

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
	fmt.Println("source", string(src))
	// Explicitly initialize all fields since a scanner may be reused.
	if file.Size() != len(src) {
		panic(fmt.Sprintf("file size (%d) does not match src len (%d)", file.Size(), len(src)))
	}
	s.file = file
	s.dir, _ = filepath.Split(file.Name())
	s.src = src
	s.err = err
	s.mode = mode

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

func (s *Scanner) Scan() (pos token.Pos, tok token.Token, lit string) {
scanAgain:
	s.skipWhitespace()

	pos = s.file.Pos(s.offset)
	ch := s.ch
	switch {
	case ch == '$' || isLetter(ch):
		lit = s.scanIdent()
		// Do some string analysis to determine token
		tok = token.IDENT
	case '0' <= ch && ch <= '9':
		tok, lit = s.scanNumber(false)
	}

	if tok != token.ILLEGAL {
		return
	}

	// move forward
	s.next()
	switch ch {
	case -1:
		tok = token.EOF
	case '.':
		if '0' <= s.ch && s.ch <= '9' {
			tok, lit = s.scanNumber(true)
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
	case '^':
		tok = token.XOR
	case '#':
		tok = token.NUMBER
	case '&':
		tok = token.AND

	case '<':
		tok = s.switch2(token.LSS, token.LEQ)
	case '>':
		tok = s.switch2(token.GTR, token.GEQ)
	case '=':
		tok = s.switch2(token.ASSIGN, token.EQL)

	case '@':
		tok = token.AT
	case '$':
		tok = token.DOLLAR
	case '!':
		tok = s.switch2(token.NOT, token.NEQ)
	case ':':
		tok = token.COLON
	case ',':
		tok = token.COMMA
	case ';':
		tok = token.SEMICOLON
		lit = ";"
	case '"':
		tok = token.QUOTE
	case '(':
		tok = token.LPAREN
	case ')':
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
	case '-':
		tok = token.SUB
	case '*':
		tok = token.MUL
	default:
		fmt.Printf("Illegal %q\n", ch)
	}

	// item = Item{Type: ItemILLEGAL, Value: string(ch)}
	return
}

func (s *Scanner) scanIdent() string {
	offs := s.offset
	for s.ch == '$' || isLetter(s.ch) || isDigit(s.ch) {
		s.next()
	}
	ss := string(s.src[offs:s.offset])
	fmt.Println("ident", ss)
	return ss
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
	// Track scanning errors
	// if s.err != nil {
	// 	s.err(s.file.Position(s.file.Pos(offs)), msg)
	// }
	// s.ErrorCount++
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
