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
	mode = 0 //Trace | ParseComments
	var name string
	defer func() { fmt.Println("exit parsing", name) }()
	for _, name = range inputs {

		if !strings.Contains(name, "13_") {
			continue
		}
		if strings.Contains(name, "06_") {
			continue
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

func testString(t *testing.T, in string, mode Mode) (*ast.File, *token.FileSet) {
	fset := token.NewFileSet()
	f, err := ParseFile(fset, "testfile", in, mode)
	if err != nil {
		t.Fatal(err)
	}
	return f, fset

}

func TestSelMath(t *testing.T) {
	// Selectors act like boolean math
	in := `
div ~ span { }`
	f, fset := testString(t, in, 0)
	_ = fset
	sel, ok := f.Decls[0].(*ast.SelDecl)
	if !ok {
		t.Fatal("SelDecl expected")
	}

	bexpr, ok := sel.Sel.(*ast.BinaryExpr)
	if !ok {
		t.Fatal("BinaryExpr expected")
	}

	lit, ok := bexpr.X.(*ast.BasicLit)
	if !ok {
		t.Fatal("BasicLit expected")
	}

	if e := "div"; lit.Value != e {
		t.Errorf("got: %s wanted: %s", lit.Value, e)
	}

	if e := token.TIL; bexpr.Op != e {
		t.Errorf("got: %s wanted: %s", bexpr.Op, e)
	}

	lit, ok = bexpr.Y.(*ast.BasicLit)
	if !ok {
		t.Fatal("BasicLit expected")
	}

	if e := "span"; lit.Value != e {
		t.Errorf("got: %s wanted: %s", lit.Value, e)
	}
}

func TestBackRef(t *testing.T) {
	// Selectors act like boolean math
	in := `div { & { color: red; } }`
	f, fset := testString(t, in, 0)
	_, _ = f, fset

	sel, ok := f.Decls[0].(*ast.SelDecl)
	if !ok {
		t.Fatal("SelDecl expected")
	}

	lit, ok := sel.Sel.(*ast.BasicLit)
	if !ok {
		t.Fatal("BasicLit expected")
	}
	if e := "div"; lit.Value != e {
		t.Errorf("got: %s wanted: %s", lit.Value, e)
	}

	nested, ok := sel.Body.List[0].(*ast.SelStmt)
	if !ok {
		t.Fatal("expected SelStmt")
	}

	if e := "&"; nested.Name.String() != e {
		t.Fatal("got: %s wanted: %s", nested.Name.String(), e)
	}

	lit, ok = nested.Sel.(*ast.BasicLit)
	if !ok {
		t.Fatal("expected lit")
	}

	if e := "div"; lit.Value != e {
		t.Errorf("got: %s wanted: %s", lit.Value, e)
	}
}

func TestExprMath(t *testing.T) {
	// Selectors act like boolean math
	in := `
div {
  value: 1*(2+3);
}`
	f, fset := testString(t, in, 0)
	ast.Print(fset, f.Decls[0].(*ast.SelDecl).Body.List[0].(*ast.DeclStmt).Decl.(*ast.GenDecl).Specs[0].(*ast.RuleSpec).Values[0])
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
