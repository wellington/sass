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
	mode = Trace | ParseComments
	var name string
	defer func() { fmt.Println("exit parsing", name) }()
	for _, name = range inputs {

		if !strings.Contains(name, "17_") {
			// continue
		}
		// These are fucked things in Sass like lists
		if strings.Contains(name, "15_") {
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

func TestParseDir(t *testing.T) {

}

func TestPrint(t *testing.T) {

	fset := token.NewFileSet()
	f, err := ParseFile(fset, "../sass-spec/spec/basic/19_full_mixin_craziness/input.scss", nil, Trace)
	_ = f
	if err != nil {
		t.Fatal(err)
	}

	ast.Print(fset, nil)
}

func TestParse_quotes(t *testing.T) {
	f, err := ParseFile(token.NewFileSet(), "main.scss", `$zz : word;`, 0)
	if err != nil {
		t.Fatal(err)
	}

	vals := f.Decls[0].(*ast.GenDecl).Specs[0].(*ast.ValueSpec).Values
	if e := 1; len(vals) != e {
		for _, v := range vals {
			fmt.Printf("%s % #v\n", v.(*ast.Ident), v)
		}
		t.Fatalf("got: %d wanted: %d", len(vals), e)
	}

	_, ok := vals[0].(*ast.Value)
	if !ok {
		t.Fatal("IDENT not found")
	}

	f, err = ParseFile(token.NewFileSet(), "main.scss", `$zz : "word";`, 0)
	if err != nil {
		t.Fatal(err)
	}

	vals = f.Decls[0].(*ast.GenDecl).Specs[0].(*ast.ValueSpec).Values
	if e := 1; len(vals) != e {
		t.Fatalf("got: %d wanted: %d", len(vals), e)
	}

	lit := vals[0].(*ast.Value)
	if e := token.QSTRING; e != lit.Kind {
		t.Fatalf("got: %s wanted: %s", lit.Kind, e)
	}
}
