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

	decl, ok := f.Decls[0].(*ast.SelDecl)
	if !ok {
		t.Fatal("SelDecl expected")
	}
	sel := decl.SelStmt
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
		t.Fatalf("got: %s wanted: %s", nested.Name.String(), e)
	}

	if e := "div"; e != nested.Resolved.Value {
		t.Errorf("got: %s wanted: %s", nested.Resolved.Value, e)
	}
}
