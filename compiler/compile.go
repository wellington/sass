package compiler

import (
	"bytes"
	"fmt"

	"github.com/wellington/sass/ast"
	"github.com/wellington/sass/parser"
	"github.com/wellington/sass/token"
)

type Context struct {
	buf *bytes.Buffer
}

func (ctx *Context) Init() {
	ctx.buf = bytes.NewBuffer(nil)
}

// Run takes a single Sass file and compiles it
func (ctx *Context) Run(path string) (string, error) {
	// func ParseFile(fset *token.FileSet, filename string, src interface{}, mode Mode) (f *ast.File, err error) {
	fset := token.NewFileSet()
	pf, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return "", err
	}
	ctx.printSels(pf.Decls)
	return ctx.buf.String(), nil
}

func (ctx *Context) printSels(decls []ast.Decl) {
	for _, decl := range decls {
		switch v := decl.(type) {
		case *ast.SelDecl:
			fmt.Fprint(ctx.buf, v.Name, " ")
			fmt.Fprint(ctx.buf, "{")
			// fmt.Printf("body: % #v\n", v.Body)
			ctx.printStmts(v.Body.List)
			fmt.Fprintln(ctx.buf, " }")
		default:
			fmt.Printf("type: %s % #v\n", v, decl)
		}
	}
}

func (ctx *Context) printStmts(stmts []ast.Stmt) {
	for _, stmt := range stmts {
		fmt.Fprint(ctx.buf, "\n  ")
		switch v := stmt.(type) {
		case *ast.DeclStmt:
			ctx.printDecl(v.Decl)
		default:
			fmt.Printf("defaul stmts: % #v\n", v)
		}
	}
}

func (ctx *Context) printDecl(decl ast.Decl) {
	switch v := decl.(type) {
	case *ast.GenDecl:
		// fmt.Printf("gendecl: % #v\n", v)
		ctx.printSpecs(v.Specs)
	default:
		fmt.Printf("default: % #v\n", v)
	}
}

func (ctx *Context) printSpecs(specs []ast.Spec) {
	for _, spec := range specs {
		switch v := spec.(type) {
		case *ast.RuleSpec:
			ctx.printIdents([]*ast.Ident{v.Name})
			fmt.Fprint(ctx.buf, ": ")
		case *ast.ValueSpec:
			ctx.printIdents(v.Names)
			fmt.Fprintf(ctx.buf, ";")
		default:
			fmt.Printf("printSpecs: % #v\n", v)
		}
	}
}

func (ctx *Context) printIdents(idents []*ast.Ident) {
	for _, ident := range idents {
		fmt.Fprint(ctx.buf, ident.Name)
	}
}
