// Scanner for selectors. Selectors are special snowflakes and
// complex enough to require rescanning. The result can likely pass
// through the regular parser and be fine.
package selectors

import (
	"unicode"
	"unicode/utf8"

	"github.com/wellington/sass/scanner"

	"github.com/wellington/sass/token"
)

type Scanner struct {
	src    []byte
	ch     rune
	offset int

	file       *token.File
	err        scanner.ErrorHandler
	ErrorCount int
	rdOffset   int
	lineOffset int
}

func (s *Scanner) Init(file *token.File, src []byte, err scanner.ErrorHandler, mode scanner.Mode) {
	s.file = file
	s.src = src

	s.ch = ' '
	s.offset = 0
	s.rdOffset = 0
	s.lineOffset = 0
	s.ErrorCount = 0

	s.next()
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

func (s *Scanner) error(offs int, msg string) {
	if s.err != nil {
		s.err(s.file.Position(s.file.Pos(offs)), msg)
	}
	s.ErrorCount++
}

func (s *Scanner) skipWhitespace() {
	for s.ch == ' ' || s.ch == '\t' || s.ch == '\n' || s.ch == '\r' {
		s.next()
	}
}

// Types of Selectors https://developer.mozilla.org/en-US/docs/Web/Guide/CSS/Getting_started/Selectors
// attribute:   [disabled] [type='button'] [class~=key] [lang|=es]
//              [title*="example" i] a[href^="https://"] img[src$=".png"]
// pseudo:      :link, :visited, :first-child, :nth-last-child
// specificity: A E, A > E, E:first-child, B + E
// class/id   : .carrot, #first, strong
func (s *Scanner) Scan() (pos token.Pos, tok token.Token, lit string) {

	s.skipWhitespace()

	pos = s.file.Pos(s.offset)
	offs := s.offset

	switch ch := s.ch; {
	case isLetter(ch):
		tok = token.STRING
		for !unicode.IsSpace(s.ch) {
			s.next()
		}
		lit = string(s.src[offs:s.offset])
	default:
		s.next()
		switch ch {
		case -1:
			lit = ""
			tok = token.EOF
		case '&':
			tok = token.AND
		case '>':
			tok = token.GTR
		case '+':
			tok = token.ADD
		case ',':
			tok = token.COMMA
		case '[':
			tok = token.ATTRIBUTE
			for s.ch != ']' {
				s.next()
			}
			s.next()
			lit = string(s.src[offs:s.offset])
		case ':':
			tok = token.PSEUDO
			for s.ch != ',' && !unicode.IsSpace(s.ch) {
				s.next()
			}
		default:
			tok = token.ILLEGAL
			lit = string(ch)
		}
	}

	return
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch >= 0x80 && unicode.IsLetter(ch)
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
