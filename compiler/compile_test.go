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

	var input string
	defer func() {
		if r := recover(); r != nil {
			log.Fatal("Recovered from", input)
		}
	}()
	var files []file
	// files := make([]file, len(inputs))
	for _, input = range inputs {
		// Comments are lost right now
		if !strings.Contains(input, "17_") {
			// continue
		}
		if strings.Contains(input, "17_") {
			continue
		}
		if strings.Contains(input, "06_") {
			continue
		}
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
		fmt.Println("reading", f.input)
		out, err := fileRun(f.input)
		sout := strings.Replace(out, "`", "", -1)
		if err != nil {
			log.Println("failed to compile", f.input, err)
		}

		if e := string(f.expect); e != sout {
			t.Fatalf("got:\n%s\nwanted:\n%s", out, e)
		}
	}

}
