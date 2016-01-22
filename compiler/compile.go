package compiler

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/wellington/sass/ast"
	"github.com/wellington/sass/parser"
	"github.com/wellington/sass/token"
)

type Context struct {
	buf      *bytes.Buffer
	fileName *ast.Ident
	level    int
	printers map[ast.Node]func(*Context, ast.Node)
}

// Run takes a single Sass file and compiles it
func (ctx *Context) Run(path string) (string, error) {
	// func ParseFile(fset *token.FileSet, filename string, src interface{}, mode Mode) (f *ast.File, err error) {
	fset := token.NewFileSet()
	pf, err := parser.ParseFile(fset, path, nil, parser.ParseComments|parser.Trace)
	if err != nil {
		return "", err
	}

	ast.Walk(ctx, pf)
	fmt.Fprintf(ctx.buf, "\n")
	// ctx.printSels(pf.Decls)
	return ctx.buf.String(), nil
}

func (ctx *Context) out(v interface{}) {
	ws := []byte("                                              ")
	format := append(ws[:ctx.level*2], "%s"...)
	fmt.Fprintf(ctx.buf, string(format), v)
}

func (ctx *Context) Visit(node ast.Node) ast.Visitor {
	switch v := node.(type) {
	case *ast.BlockStmt:
		fmt.Fprintf(ctx.buf, "{\n")
		ctx.level = ctx.level + 1
		for _, node := range v.List {
			ast.Walk(ctx, node)
		}
		ctx.level = ctx.level - 1
		fmt.Fprintf(ctx.buf, " }")
		// ast.Walk(ctx, v.List)
		// fmt.Fprintf(ctx.buf, "}")
		return nil
	case *ast.SelDecl:
		ctx.printers[selDecl](ctx, v)
	case *ast.File:
		// Nothing to print for these
	case *ast.GenDecl:

	case *ast.Ident:
		// The first IDENT is always the filename, just preserve
		// it somewhere
		if ctx.fileName == nil {
			ctx.fileName = ident
			return ctx
		}
		ctx.printers[ident](ctx, v)
	case *ast.DeclStmt:
		ctx.printers[declStmt](ctx, v)
	case *ast.ValueSpec:
		ctx.printers[valueSpec](ctx, v)
	case *ast.RuleSpec:
		ctx.printers[ruleSpec](ctx, v)
	case nil:

	default:
		fmt.Printf("add printer for: %T\n", v)
		fmt.Printf("% #v\n", v)
	}
	return ctx
}

var (
	ident     *ast.Ident
	declStmt  *ast.DeclStmt
	valueSpec *ast.ValueSpec
	ruleSpec  *ast.RuleSpec
	selDecl   *ast.SelDecl
)

func (ctx *Context) Init() {
	ctx.buf = bytes.NewBuffer(nil)
	ctx.printers = make(map[ast.Node]func(*Context, ast.Node))
	ctx.printers[ident] = printIdent
	ctx.printers[declStmt] = printDecl
	ctx.printers[valueSpec] = printValueSpec
	ctx.printers[ruleSpec] = printRuleSpec
	ctx.printers[selDecl] = printSelDecl
	// assign printers
}

func printSelDecl(ctx *Context, n ast.Node) {
	decl := n.(*ast.SelDecl)
	fmt.Fprintf(ctx.buf, "%s ", decl.Name)
}

func printRuleSpec(ctx *Context, n ast.Node) {
	spec := n.(*ast.RuleSpec)
	ctx.out(spec.Name.String() + ": ")
	// fmt.Fprintf(ctx.buf, "%s:", spec.Name)
}

func printValueSpec(ctx *Context, n ast.Node) {
	spec := n.(*ast.ValueSpec)
	names := make([]string, len(spec.Names))
	for i, nm := range spec.Names {
		names[i] = nm.Name
	}
	fmt.Fprintf(ctx.buf, "%s;", strings.Join(names, " "))
}

func printDecl(ctx *Context, ident ast.Node) {
	// I think... nothing to print we'll see
}

func printIdent(ctx *Context, ident ast.Node) {
	fmt.Printf("% #v\n", ident)
	fmt.Fprintf(ctx.buf, "%s", ident)
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
