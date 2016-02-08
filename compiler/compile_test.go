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
		if r := recover(); r != nil {
			log.Fatal("Recovered from", input)
		}
	}()

	var files []file
	// files := make([]file, len(inputs))
	for _, input = range inputs {
		if !strings.Contains(input, "13_") {
			continue
		}
		// detailed commenting
		if strings.Contains(input, "06_") {
			continue
		}
		// back references
		if strings.Contains(input, "13_") {
			continue
		}
		if strings.Contains(input, "14_") {
			continue
		}

		// parser skips
		if strings.Contains(input, "15_") {
			continue
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

func TestRun(t *testing.T) {
	files := findPaths()
	var f file
	defer func() {
		fmt.Println("exited on: ", f.input)
	}()
	for _, f = range files {
		fmt.Println("compiling", f.input)
		out, err := fileRun(f.input)
		sout := strings.Replace(out, "`", "", -1)
		if err != nil {
			log.Println("failed to compile", f.input, err)
		}

		if e := string(f.expect); e != sout {
			// t.Fatalf("got:\n%s", out)
			t.Fatalf("got:\n%q\nwanted:\n%q", out, e)
			// t.Fatalf("got:\n%s\nwanted:\n%s", out, e)
		}
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
		t.Errorf("got:\n%s\nwanted:\n%s", out, e)
	}

}

func TestSelector_ampersand(t *testing.T) {
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
		t.Errorf("got:\n%s\nwanted:\n%s", out, e)
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
		t.Fatal(err)
	}

	e := `div ~ b + div ~ b {
  color: red; }
`
	if e != out {
		t.Errorf("got:\n%s\nwanted:\n%s", out, e)
	}

}
