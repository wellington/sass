package parser

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"unicode"
)

type ItemType int

const NotFound = 01

// Special item types.
const (
	ItemEOF ItemType = iota
	ItemILLEGAL
	ItemSpace
	ItemError
)

func IsSymbol(r rune) bool {
	return strings.ContainsRune("(),;{}#:", r)
}

func isSpace(r rune) bool {
	return unicode.IsSpace(r)
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
	r      *bufio.Reader
	ch     rune
	offset int
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

func (s *Scanner) read() {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		s.ch = eof
		return
	}
	s.offset += 1
	s.ch = ch
}

func (s *Scanner) unread() {
	_ = s.r.UnreadRune()
	s.offset -= 1
}

type Item struct {
	Type  ItemType
	Pos   int
	Value string
}

func (s *Scanner) skipWhitespace() {
	for s.ch == ' ' || s.ch == '\n' || s.ch == '\r' {
		s.read()
	}
}

func (s *Scanner) Scan() (item Item) {

	s.skipWhitespace()

	ch := s.ch
	switch {
	case isAllowedRune(ch):
		lit := s.scanIdent()
		// Do some string analysis to determine token
		item = Item{Type: ItemError, Value: lit}
		return
	}

	// move forward
	s.read()
	switch ch {
	case eof:
		item = Item{Type: ItemEOF}
	case '/':
		if s.ch = '/' || s.ch == '*' {

		}
	}

	// item = Item{Type: ItemILLEGAL, Value: string(ch)}
}

func (s *Scanner) scanIdent() string {
	return ""
}

// scanWhitespace consumes the current rune and all contiguous whitespace.
func (s *Scanner) scanWhitespace() Item {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	s.read()
	buf.WriteRune(s.ch)

	// Read every subsequent whitespace character into the buffer.
	// Non-whitespace characters and EOF will cause the loop to exit.
	for {
		s.read()
		if s.ch == eof {
			break
		} else if !isSpace(s.ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(s.ch)
		}
	}

	return Item{
		Type:  ItemSpace,
		Value: buf.String(),
	}
}
