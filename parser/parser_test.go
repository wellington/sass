package parser

import (
	"testing"

	"github.com/wellington/sass/ast"
	"github.com/wellington/sass/token"
)

var validFiles = []string{
	"../sass-spec/spec/basic/00_empty/input.scss",
	"../sass-spec/spec/basic/01_simple_css/input.scss",
	"../sass-spec/spec/basic/02_simple_nesting/input.scss",
	"../sass-spec/spec/basic/03_simple_variable/input.scss",
	"../sass-spec/spec/basic/04_basic_variables/input.scss",
	"../sass-spec/spec/basic/05_empty_levels/input.scss",
	"../sass-spec/spec/basic/06_nesting_and_comments/input.scss",
	"../sass-spec/spec/basic/07_nested_simple_selector_groups/input.scss",
}

func TestParse_files(t *testing.T) {
	mode := DeclarationErrors
	mode = AllErrors + Trace
	for _, name := range validFiles {
		_, err := ParseFile(token.NewFileSet(), name, nil, mode)
		if err != nil {
			t.Fatalf("ParseFile(%s): %v", name, err)
		}
	}
}

func TestParseDir(t *testing.T) {
	// paths := "../sass-spec/spec/basic/00_empty"
}

func TestVarScope_list2(t *testing.T) {
	t.Skip("Parser will have to split rhs lists")
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

func TestVarScope_quotes(t *testing.T) {
	f, err := ParseFile(token.NewFileSet(), "main.scss", `$zz : word;`, 0)
	if err != nil {
		t.Fatal(err)
	}

	vals := f.Decls[0].(*ast.GenDecl).Specs[0].(*ast.ValueSpec).Values

	if e := 1; len(vals) != e {
		for _, v := range vals {
			t.Logf("%s % #v\n", v.(*ast.BasicLit).Kind, v)
		}
		t.Fatalf("got: %d wanted: %d", len(vals), e)
	}

	lit := vals[0].(*ast.BasicLit)
	if e := token.QSSTRING; e != lit.Kind {
		// t.Fatalf("got: %s wanted: %s", lit.Kind, e)
	}

	f, err = ParseFile(token.NewFileSet(), "main.scss", `$zz : "word";`, 0)
	if err != nil {
		t.Fatal(err)
	}

	vals = f.Decls[0].(*ast.GenDecl).Specs[0].(*ast.ValueSpec).Values
	if e := 1; len(vals) != e {
		t.Fatalf("got: %d wanted: %d", len(vals), e)
	}

	lit = vals[0].(*ast.BasicLit)
	if e := token.QSTRING; e != lit.Kind {
		t.Fatalf("got: %s wanted: %s", lit.Kind, e)
	}
}
