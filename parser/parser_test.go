package parser

import (
	"fmt"
	"go/token"
	"testing"
)

var validFiles = []string{
	"../sass-spec/spec/basic/00_empty/input.scss",
	"../sass-spec/spec/sass/imported/imported.sass",
	"../sass-spec/spec/libsass/bourbon/lib/_bourbon.scss",
}

func TestParse(t *testing.T) {
	for _, name := range validFiles {
		_, err := ParseFile(token.NewFileSet(), name, nil, DeclarationErrors)
		if err != nil {
			t.Fatalf("ParseFile(%s): %v", name, err)
		}
	}
}

func TestParseDir(t *testing.T) {
	// paths := "../sass-spec/spec/basic/00_empty"
}

func TestVarScope(t *testing.T) {
	f, err := ParseFile(token.NewFileSet(), "", `$z = x;`, 0)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("% #v\n", f.Name)
	fmt.Printf("% #v\n", f.Scope)
	fmt.Printf("% #v\n", f)
}
