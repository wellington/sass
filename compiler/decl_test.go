package compiler

import (
	"testing"

	"github.com/wellington/sass/token"
)

func TestDecl_if(t *testing.T) {
	ctx := &Context{}
	ctx.Init()
	ctx.fset = token.NewFileSet()
	input := `$x: 1 2;
@if 1 + 1 == length($x) {
  div { hi: there;
  a: type-of(nth($x, 1));
  }
}
`
	// ctx.SetMode(parser.Trace)
	out, err := ctx.run("", input)
	if err != nil {
		t.Fatal(err)
	}

	e := `div {
  hi: there;
  a: number; }
`
	if e != out {
		t.Fatalf("got:\n%s\nwanted:\n%s", out, e)
	}
}
