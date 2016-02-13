package compiler

import (
	"testing"

	"github.com/wellington/sass/token"
)

func runParse(t *testing.T, in string, e string) {
	ctx := &Context{}
	ctx.Init()
	ctx.fset = token.NewFileSet()

	out, err := ctx.run("", in)
	if err != nil {
		t.Fatal(err)
	}
	if e != out {
		t.Errorf("got:\n%q\nwanted:\n%q", out, e)
	}
}

func TestBuiltin_inspect(t *testing.T) {
	in := `$x: 1;
hey, ho {
 a: inspect(1);
 b: inspect(a);
 c: inspect(#000);
 d: inspect("a");
 e: inspect('a');
 f: inspect($x);
}
`
	e := `hey, ho {
  a: 1;
  b: a;
  c: #000;
  d: "a";
  e: 'a';
  f: 1; }
`
	runParse(t, in, e)
}

func TestBuiltin_typeof(t *testing.T) {
	in := `$x: 1;
    hey, ho {
		a: type-of(1);
		b: type-of(a);
		c: type-of(#000);
		d: type-of("a");
		e: type-of('a');
        f: type-of($x);
	}`

	e := `hey, ho {
  a: number;
  b: string;
  c: color;
  d: string;
  e: string;
  f: number; }
`
	runParse(t, in, e)

}
