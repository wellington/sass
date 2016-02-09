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
	defer func() {
		fmt.Println("Exited on", input)
	}()

	var files []file
	// files := make([]file, len(inputs))
	for _, input = range inputs {
		if !strings.Contains(input, "13_") {
			//continue
		}
		// detailed commenting
		if strings.Contains(input, "06_") {
			continue
		}
		// bad math `> e`
		if strings.Contains(input, "09_") {
			continue
		}
		// ditto
		if strings.Contains(input, "10_") {
			continue
		}
		// back references
		if strings.Contains(input, "13_") {
			//continue
		}
		if strings.Contains(input, "14_") {
			continue
		}

		// parser skips
		if strings.Contains(input, "15_") {
			//continue
		}
		// Skip for built-in rules
		if strings.Contains(input, "16_") {
			continue
		}
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
		out, err := fileRun(f.input)
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
    color: blue;
	}
}
`
	out, err := ctx.run("", input)
	if err != nil {
		t.Fatal(err)
	}

	e := `a c, a d, a e {
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
