package parser

import (
	"testing"

	"github.com/wellington/sass/ast"
	"github.com/wellington/sass/token"
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

func TestVarScope_list2(t *testing.T) {
	f, err := ParseFile(token.NewFileSet(), "main.scss", `$zz : x,y;`, Trace)
	if err != nil {
		t.Fatal(err)
	}

	if e := "main.scss"; e != f.Name.Name {
		t.Fatalf("got: %s wanted: %s", f.Name, e)
	}

	vals := f.Decls[0].(*ast.GenDecl).Specs[0].(*ast.ValueSpec).Values

	if e := 2; len(vals) != e {
		t.Fatalf("got: %d wanted: %d", len(vals), e)
	}
}
