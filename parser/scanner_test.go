package parser

import (
	"fmt"
	gotoken "go/token"
	"testing"

	"github.com/wellington/sass/token"
)

type elt struct {
	tok token.Token
	lit string
	// class int
}

const whitespace = "  \t  \n\n\n" // to separate tokens

var tokens = [...]elt{
	{token.CMT, "/* a comment */"},
	{token.CMT, "//"},
}

var source = func() []byte {
	var src []byte
	for _, t := range tokens {
		src = append(src, t.lit...)
		src = append(src, whitespace...)
	}
	return src
}()

var fset = gotoken.NewFileSet()

func TestScan(t *testing.T) {

	// error handler
	eh := func(_ gotoken.Position, msg string) {
		t.Errorf("error handler called (msg = %s)", msg)
	}

	var s Scanner
	s.Init(fset.AddFile("", fset.Base(), len(source)), source, eh)

	epos := gotoken.Position{
		Filename: "",
		Offset:   0,
		Line:     1,
		Column:   1,
	}
	_ = epos
	for {
		pos, tok, lit := s.Scan()
		fmt.Println(pos, tok, lit)
	}
}
