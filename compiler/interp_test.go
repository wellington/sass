package compiler

import (
	"testing"

	"github.com/wellington/sass/parser"
	"github.com/wellington/sass/token"
)

func TestInterp(t *testing.T) {
	ctx := &Context{}
	ctx.Init()
	ctx.SetMode(parser.Trace)
	ctx.fset = token.NewFileSet()
	input := `div {
  hello: #{123+321};
}
`
	out, err := ctx.run("", input)
	if err != nil {
		t.Fatal(err)
	}

	e := `div {
  hello: 444; }
`
	if e != out {
		t.Fatalf("got:\n%s\nwanted:\n%s", out, e)
	}
}

func TestInterp_merge_front(t *testing.T) {
	ctx := &Context{}
	ctx.Init()
	ctx.SetMode(parser.Trace)
	ctx.fset = token.NewFileSet()
	input := `div {
  hello: before#{123+321};
}
`
	out, err := ctx.run("", input)
	if err != nil {
		t.Fatal(err)
	}

	e := `div {
  hello: before444; }
`
	if e != out {
		t.Fatalf("got:\n%s\nwanted:\n%s", out, e)
	}
}

func TestInterp_merge_back(t *testing.T) {
	ctx := &Context{}
	ctx.Init()
	ctx.SetMode(parser.Trace)
	ctx.fset = token.NewFileSet()
	input := `div {
  hello: #{123+321}after;
}
`
	out, err := ctx.run("", input)
	if err != nil {
		t.Fatal(err)
	}

	e := `div {
  hello: 444after; }
`
	if e != out {
		t.Fatalf("got:\n%s\nwanted:\n%s", out, e)
	}
}

func TestInterp_merge_both(t *testing.T) {
	ctx := &Context{}
	ctx.Init()
	ctx.SetMode(parser.Trace)
	ctx.fset = token.NewFileSet()
	input := `div {
  hello: before#{123+321}after;
}
`
	out, err := ctx.run("", input)
	if err != nil {
		t.Fatal(err)
	}

	e := `div {
  hello: before444after; }
`
	if e != out {
		t.Fatalf("got:\n%s\nwanted:\n%s", out, e)
	}
}
