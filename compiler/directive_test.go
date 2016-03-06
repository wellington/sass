package compiler

import (
	"testing"

	"github.com/wellington/sass/parser"
	"github.com/wellington/sass/token"
)

func TestDirective_each_paran(t *testing.T) {
	ctx := &Context{}
	ctx.Init()
	ctx.SetMode(parser.Trace)
	ctx.fset = token.NewFileSet()
	input := `div {
  $v: 4;
  v: $v;
  @each $i in (1 2 3 4 5) {
   i: $i;
  }
}
`
	out, err := ctx.run("", input)
	if err != nil {
		t.Fatal(err)
	}

	e := `div {
  i: blah;
  i: 1;
  i: 2;
  i: 3;
  i: 4;
  i: 5; }
`
	if e != out {
		t.Fatalf("got:\n%s\nwanted:\n%s", out, e)
	}
}

func TestDirective_each(t *testing.T) {
	ctx := &Context{}
	ctx.Init()
	ctx.fset = token.NewFileSet()
	input := `div {
  @each $i in a b c {
   i: $i;
  }
}
`
	out, err := ctx.run("", input)
	if err != nil {
		t.Fatal(err)
	}

	e := ``
	if e != out {
		t.Fatalf("got:\n%s\nwanted:\n%s", out, e)
	}
}
