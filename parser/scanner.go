package parser

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	goscanner "go/scanner"
	gotoken "go/token"

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

type Scanner struct {
	src    []byte
	ch     rune
	offset int

	file       *gotoken.File
	dir        string
	err        goscanner.ErrorHandler
	ErrorCount int
	rdOffset   int
	lineOffset int
}

func (s *Scanner) Init(file *gotoken.File, src []byte, err goscanner.ErrorHandler) {

	// Explicitly initialize all fields since a scanner may be reused.
	if file.Size() != len(src) {
		panic(fmt.Sprintf("file size (%d) does not match src len (%d)", file.Size(), len(src)))
	}
	s.file = file
	s.dir, _ = filepath.Split(file.Name())
	s.src = src
	s.err = err
	// s.mode = mode

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
	offs := s.offset
	for s.ch == ' ' || s.ch == '\n' || s.ch == '\r' {
		s.next()
	}
	fmt.Printf("skipping`%s`", string(s.src[offs:s.offset]))
}

func (s *Scanner) Scan() (pos gotoken.Pos, tok token.Token, lit string) {

	s.skipWhitespace()

	pos = s.file.Pos(s.offset)
	ch := s.ch
	switch {
	case isLetter(ch):
		fmt.Println("letter")
		lit = s.scanIdent()
		// Do some string analysis to determine token
		tok = token.IDENT
		return
	}

	fmt.Println("1", string(s.ch))
	// move forward
	s.next()
	fmt.Println("2", string(s.ch))
	switch ch {
	case eof:
		tok = token.EOF
		return
	case '/':
		fmt.Println("find", s.ch)
		if s.ch == '/' || s.ch == '*' {
			fmt.Println("scan comment")
			comment := s.scanComment()
			tok = token.CMT
			lit = comment
			return
		}
	}
	fmt.Println("find", string(s.ch))

	// item = Item{Type: ItemILLEGAL, Value: string(ch)}
	return
}

func (s *Scanner) scanIdent() string {
	offs := s.offset
	for isLetter(s.ch) || isDigit(s.ch) {
		s.next()
	}
	ss := string(s.src[offs:s.offset])
	fmt.Println(ss)
	return ss
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
