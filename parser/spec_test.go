package parser

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wellington/sass/token"
)

func TestSpec_files(t *testing.T) {
	if testing.Short() {
		t.Skip("Skip robust testing to better indicate errors")
	}

	inputs, err := filepath.Glob("../sass-spec/spec/basic/*/input.scss")
	if err != nil {
		t.Fatal(err)
	}

	mode := DeclarationErrors
	mode = 0 //Trace | ParseComments
	var name string
	defer func() { fmt.Println("exit parsing", name) }()
	for _, name = range inputs {

		if !strings.Contains(name, "23_") {
			continue
		}
		if strings.Contains(name, "06_") {
			continue
		}
		// These are fucked things in Sass like lists
		if strings.Contains(name, "15_") {
			continue
		}
		if strings.Contains(name, "16_") {
			continue
		}
		// namespaces are wtf
		if strings.Contains(name, "24_") {
			continue
		}
		fmt.Println("Parsing", name)
		_, err := ParseFile(token.NewFileSet(), name, nil, mode)
		if err != nil {
			t.Fatalf("ParseFile(%s): %v", name, err)
		}
		fmt.Println("Parsed", name)
	}
}
