package compiler

import (
	"testing"

	"github.com/wellington/sass/token"
)

func TestDecl_if(t *testing.T) {
	ctx := &Context{}
	ctx.Init()
	ctx.fset = token.NewFileSet()
	input := `
@if 1 == 1 {
  div { hi: there; }
}
`
	// ctx.SetMode(parser.Trace)
	out, err := ctx.run("", input)
	if err != nil {
		t.Fatal(err)
	}

	e := `div {
  hi: there; }
`
	if e != out {
		t.Fatalf("got:\n%s\nwanted:\n%s", out, e)
	}
}
