package compiler

import (
	"bytes"
	"fmt"
	"log"
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

func fileRun(path string) (string, error) {
	ctx := &Context{}
	ctx.Init()
	out, err := ctx.Run(path)
	if err != nil {
		log.Fatal(err)
	}
	return out, err
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
	if ctx.buf.Len() > 0 {
		ctx.out("\n")
	}
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
	case *ast.SelStmt:
		// We will need to combine parent selectors
		// while printing these
		ctx.printers[selStmt](ctx, v)
		// Nothing to do
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
	selStmt   *ast.SelStmt
)

func (ctx *Context) Init() {
	ctx.buf = bytes.NewBuffer(nil)
	ctx.printers = make(map[ast.Node]func(*Context, ast.Node))
	ctx.printers[ident] = printIdent
	ctx.printers[declStmt] = printDecl
	ctx.printers[valueSpec] = printValueSpec
	ctx.printers[ruleSpec] = printRuleSpec
	ctx.printers[selDecl] = printSelDecl
	ctx.printers[selStmt] = printSelStmt
	// assign printers
}

func printSelStmt(ctx *Context, n ast.Node) {
	stmt := n.(*ast.SelStmt)
	ctx.out(stmt.Name.String() + " ")
}

func printSelDecl(ctx *Context, n ast.Node) {
	decl := n.(*ast.SelDecl)
	ctx.out(decl.Name.String() + " ")
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
