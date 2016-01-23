package compiler

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"testing"
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

	files := make([]file, len(inputs))
	for i, input := range inputs {
		exp, err := ioutil.ReadFile(strings.Replace(input,
			"input.scss", "expected_output.css", 1))
		if err != nil {
			log.Println("failed to read", input)
			continue
		}

		files[i] = file{
			input:  input,
			expect: exp,
		}
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
		fmt.Println("reading", f.input)
		out, err := fileRun(f.input)
		if err != nil {
			log.Println("failed to compile", f.input, err)
		}

		if e := string(f.expect); e != out {
			t.Fatalf("got:\n%q\nwanted:\n%q", out, e)
		}
	}

}
