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
		t.Errorf("got:\n%s\nwanted:\n%s", out, e)
	}
}

func TestTypeOf(t *testing.T) {
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
