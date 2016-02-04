package selectors

import (
	"testing"

	"github.com/wellington/sass/scanner"
	"github.com/wellington/sass/token"
)

type elt struct {
	tok token.Token
	lit string
}

const whitespace = "   \n\n\n"

var elts = []elt{
	{token.STRING, "a#id"},
	{token.COMMA, ","},
	{token.ATTRIBUTE, "[disabled]"},
	{token.STRING, "A"},
	{token.GTR, ">"},
	{token.STRING, "E"},
	{token.ADD, "+"},
	{token.STRING, "F"},
}

var source = func(tokens []elt) []byte {
	var src []byte
	for _, t := range tokens {
		src = append(src, t.lit...)
		src = append(src, whitespace...)
	}

	return src
}

var fset = token.NewFileSet()

func newlineCount(s string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			n++
		}
	}
	return n
}

func checkPos(t *testing.T, lit string, p token.Pos, expected token.Position) {
	pos := fset.Position(p)
	if pos.Filename != expected.Filename {
		t.Errorf("bad filename for %q: got %s, expected %s", lit, pos.Filename, expected.Filename)
	}
	if pos.Offset != expected.Offset {
		t.Errorf("bad position for %q: got %d, expected %d", lit, pos.Offset, expected.Offset)
	}
	if pos.Line != expected.Line {
		t.Errorf("bad line for %q: got %d, expected %d", lit, pos.Line, expected.Line)
	}
	if pos.Column != expected.Column {
		t.Errorf("bad column for %q: got %d, expected %d", lit, pos.Column, expected.Column)
	}
}

func TestScan(t *testing.T) {
	testScan(t, elts)
}

func testScan(t *testing.T, tokens []elt) {
	whitespaceLinecount := newlineCount(whitespace)

	// error handler
	eh := func(_ token.Position, msg string) {
		t.Errorf("error handler called (msg = %s)", msg)
	}

	src := source(tokens)

	var s Scanner
	s.Init(fset.AddFile("", fset.Base(), len(src)), src, eh, scanner.ScanComments)

	epos := token.Position{
		Filename: "",
		Offset:   0,
		Line:     1,
		Column:   1,
	}

	index := 0
	for {
		pos, tok, lit := s.Scan()
		if tok == token.EOF {
			epos.Line = newlineCount(string(src))
			epos.Column = 2
		}
		checkPos(t, lit, pos, epos)

		// check token
		e := elt{token.EOF, ""}
		if index < len(tokens) {
			e = tokens[index]
			index++
		}
		if tok != e.tok {
			t.Errorf("bad token for %q: got %s, expected %s", lit, tok, e.tok)
		}

		// check literal
		elit := ""
		switch e.tok {
		case token.COMMENT:
			// no CRs in comments
			elit = string(stripCR([]byte(e.lit)))
			//-style comment literal doesn't contain newline
			if elit[1] == '/' {
				elit = elit[0 : len(elit)-1]
			}
		case token.IDENT:
			elit = e.lit
		case token.SEMICOLON:
			elit = ";"
		default:
			if e.tok.IsLiteral() {
				// no CRs in raw string literals
				elit = e.lit
				if elit[0] == '`' {
					elit = string(stripCR([]byte(elit)))
				}
			} else if e.tok.IsKeyword() || e.tok.IsCSSNum() {
				elit = e.lit
			}
		}
		if lit != elit {
			t.Errorf("bad literal for %q: got %q, expected %q",
				lit, lit, elit)
		}

		if tok == token.EOF {
			break
		}

		// update position
		epos.Offset += len(e.lit) + len(whitespace)
		epos.Line += newlineCount(e.lit) + whitespaceLinecount

	}
}
