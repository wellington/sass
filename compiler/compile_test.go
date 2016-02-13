package compiler

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wellington/sass/token"
)

type file struct {
	input  string // path to Sass input.scss
	expect []byte // path to expected_output.css
}

func findPaths() []file {
	inputs, err := filepath.Glob("../sass-spec/spec/basic/*/input.scss")
	if err != nil {
		log.Fatal(err)
	}

	var input string
	var files []file
	// files := make([]file, len(inputs))
	for _, input = range inputs {
		if !strings.Contains(input, "22_") {
			//continue
		}
		// detailed commenting
		if strings.Contains(input, "06_") {
			continue
		}

		// skip insane list math
		if strings.Contains(input, "15_") {
			continue
		}
		// Skip for built-in rules
		if strings.Contains(input, "24_") {
			continue
		}

		exp, err := ioutil.ReadFile(strings.Replace(input,
			"input.scss", "expected_output.css", 1))
		if err != nil {
			log.Println("failed to read", input)
			continue
		}

		files = append(files, file{
			input:  input,
			expect: exp,
		})
	}
	return files
}

func TestCompile_files(t *testing.T) {
	// It will be a long time before these are all supported, so let's just
	// short these for now.
	if testing.Short() {
		t.Skip("Skip robust testing so true errors can better be diagnosed")
	}

	files := findPaths()
	var f file
	defer func() {
		fmt.Println("exited on: ", f.input)
	}()
	for _, f = range files {
		fmt.Printf(`
=================================
compiling: %s\n
=================================
`, f.input)
		out, err := Run(f.input)
		sout := strings.Replace(out, "`", "", -1)
		if err != nil {
			log.Println("failed to compile", f.input, err)
		}

		if e := string(f.expect); e != sout {
			//t.Fatalf("got:\n%s", out)
			t.Fatalf("got:\n%q\nwanted:\n%q", out, e)
			// t.Fatalf("got:\n%s\nwanted:\n%s", out, e)
		}
		fmt.Printf(`
=================================
compiled: %s\n
=================================
`, f.input)
	}

}

func TestSelector_nesting(t *testing.T) {
	ctx := &Context{}
	ctx.Init()
	ctx.fset = token.NewFileSet()
	input := `a {
d { color: red; }
}
`
	out, err := ctx.run("", input)
	if err != nil {
		t.Fatal(err)
	}

	e := `a d {
  color: red; }
`
	if e != out {
		t.Errorf("got:\n%s\nwanted:\n%s", out, e)
	}
}

func TestSelector_inplace_nesting(t *testing.T) {
	ctx := &Context{}
	ctx.Init()
	ctx.fset = token.NewFileSet()
	input := `hey, ho {
  foo &.goo {
    color: blue;
  }
}
`
	out, err := ctx.run("", input)
	if err != nil {
		t.Fatal(err)
	}

	e := `foo hey.goo, foo ho.goo {
  color: blue; }
`
	if e != out {
		t.Fatalf("got:\n%s\nwanted:\n%s", out, e)
	}
}

func TestSelector_deep_nesting(t *testing.T) {
	ctx := &Context{}
	ctx.Init()
	ctx.fset = token.NewFileSet()
	input := `a {
	c, d, e {
	  f, g, h {
      m, n, o {
        color: blue;
      }
    }
	}
}`
	out, err := ctx.run("", input)
	if err != nil {
		t.Fatal(err)
	}

	e := `a c f m, a c f n, a c f o, a c g m, a c g n, a c g o, a c h m, a c h n, a c h o, a d f m, a d f n, a d f o, a d g m, a d g n, a d g o, a d h m, a d h n, a d h o, a e f m, a e f n, a e f o, a e g m, a e g n, a e g o, a e h m, a e h n, a e h o {
  color: blue; }
`
	if e != out {
		t.Fatalf("got:\n%s\nwanted:\n%s", out, e)
	}
}

func TestSelector_nesting_implicit_unary(t *testing.T) {

	ctx := &Context{}
	ctx.Init()
	ctx.fset = token.NewFileSet()
	input := `a {
  > e {
    color: blue;
  }
}
`
	out, err := ctx.run("", input)
	if err != nil {
		t.Fatal(err)
	}

	e := `a > e {
  color: blue; }
`
	if e != out {
		t.Fatalf("got:\n%s\nwanted:\n%s", out, e)
	}
}

func TestSelector_nesting_unary(t *testing.T) {

	// This is bizarre, may never support this odd syntax
	ctx := &Context{}
	ctx.Init()
	ctx.fset = token.NewFileSet()
	input := `a {
  & > e {
    color: blue;
  }
}
`
	out, err := ctx.run("", input)
	if err != nil {
		t.Fatal(err)
	}

	e := `a > e {
  color: blue; }
`
	if e != out {
		t.Fatalf("got:\n%s\nwanted:\n%s", out, e)
	}
}

func TestSelector_nesting_parent_group(t *testing.T) {

	ctx := &Context{}
	ctx.Init()
	ctx.fset = token.NewFileSet()
	input := `a, b {
d { color: red; }
}
`
	out, err := ctx.run("", input)
	if err != nil {
		t.Fatal(err)
	}

	e := `a d, b d {
  color: red; }
`
	if e != out {
		t.Fatalf("got:\n%s\nwanted:\n%s", out, e)
	}
}

func TestSelector_nesting_child_group(t *testing.T) {

	ctx := &Context{}
	ctx.Init()
	ctx.fset = token.NewFileSet()
	input := `a {
b, c { color: red; }
}
`
	out, err := ctx.run("", input)
	if err != nil {
		t.Fatal(err)
	}

	e := `a b, a c {
  color: red; }
`
	if e != out {
		t.Fatalf("got:\n%s\nwanted:\n%s", out, e)
	}
}

func TestSelector_many_nests(t *testing.T) {
	ctx := &Context{}
	ctx.Init()
	ctx.fset = token.NewFileSet()
	input := `a, b {
c, d { color: red; }
}
`
	out, err := ctx.run("", input)
	if err != nil {
		t.Fatal(err)
	}

	e := `a c, a d, b c, b d {
  color: red; }
`
	if e != out {
		t.Fatalf("got:\n%s\nwanted:\n%s", out, e)
	}
}

func TestSelector_combinators(t *testing.T) {
	ctx := &Context{}
	ctx.Init()
	ctx.fset = token.NewFileSet()
	input := `a + b ~ c { color: red; }
`
	out, err := ctx.run("", input)
	if err != nil {
		t.Fatal(err)
	}

	e := `a + b ~ c {
  color: red; }
`
	if e != out {
		t.Fatalf("got:\n%s\nwanted:\n%s", out, e)
	}

}

func TestSelector_singleampersand(t *testing.T) {
	ctx := &Context{}
	ctx.Init()
	ctx.fset = token.NewFileSet()
	input := `div {
& { color: red; }
}
`
	out, err := ctx.run("", input)
	if err != nil {
		t.Fatal(err)
	}

	e := `div {
  color: red; }
`
	if e != out {
		t.Fatalf("got:\n%s\nwanted:\n%s", out, e)
	}

}

func TestSelector_comboampersand(t *testing.T) {
	ctx := &Context{}
	ctx.Init()
	ctx.fset = token.NewFileSet()
	input := `div ~ b {
& + & { color: red; }
}
`
	out, err := ctx.run("", input)
	if err != nil {
		t.Fatal("compilation fail", err)
	}

	e := `div ~ b + div ~ b {
  color: red; }
`
	if e != out {
		t.Fatalf("got:\n%s\nwanted:\n%s", out, e)
	}

}
