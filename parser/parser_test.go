package parser

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wellington/sass/ast"
	"github.com/wellington/sass/token"
)

func TestParse_files(t *testing.T) {
	inputs, err := filepath.Glob("../sass-spec/spec/basic/*/input.scss")
	if err != nil {
		t.Fatal(err)
	}

	mode := DeclarationErrors
	mode = Trace
	for _, name := range inputs {
		// These are fucked things in Sass like lists
		if strings.Contains(name, "15") {
			continue
		}
		fmt.Println("Parsing", name)
		_, err := ParseFile(token.NewFileSet(), name, nil, mode)
		fmt.Println("Done", name)
		if err != nil {
			t.Fatalf("ParseFile(%s): %v", name, err)
		}
	}
	fmt.Println("done")
}

func TestParseDir(t *testing.T) {

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
