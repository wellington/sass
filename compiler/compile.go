package compiler

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"unicode/utf8"

	"github.com/wellington/sass/ast"
	"github.com/wellington/sass/parser"
	"github.com/wellington/sass/token"
)

type Context struct {
	buf      *bytes.Buffer
	fileName *ast.Ident
	// Records the current level of selectors
	// Each time a selector is encountered, increase
	// by one. Each time a block is exited, remove
	// the last selector
	sels      []string
	firstRule bool
	level     int
	printers  map[ast.Node]func(*Context, ast.Node)

	typ Typ
}

// stores types and values with scoping. To remove a scope
// use CloseScope(), to open a new Scope use OpenScope().
type Typ interface {
	// OpenScope() Typ
	// CloseScope() Typ
	Get(string) interface{}
	Set(string, interface{})
}

var (
	empty = new(emptyTyp)
)

type emptyTyp struct{}

func (*emptyTyp) Get(name string) interface{} {
	return nil
}

func (*emptyTyp) Set(name string, _ interface{}) {}

type valueTyp struct {
	Typ
	m map[string]interface{}
}

func (t *valueTyp) Get(name string) interface{} {
	return t.m[name]
}

func (t *valueTyp) Set(name string, v interface{} /* should this just be string? */) {
	t.m[name] = v
}

func NewTyp() Typ {
	return &valueTyp{Typ: empty, m: make(map[string]interface{})}
}

func NewScope(typ Typ) Typ {
	return &valueTyp{Typ: typ, m: make(map[string]interface{})}
}

func CloseScope(typ Typ) Typ {
	s, ok := typ.(*valueTyp)
	if !ok {
		return typ
	}
	return s.Typ
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
	lr, _ := utf8.DecodeLastRune(ctx.buf.Bytes())
	_ = lr
	if ctx.buf.Len() > 0 && lr != '\n' {
		ctx.out("\n")
	}
	// ctx.printSels(pf.Decls)
	return ctx.buf.String(), nil
}

// out prints with the appropriate indention, selectors always have indent
// 0
func (ctx *Context) out(v interface{}) {
	// ws := []byte("                                              ")
	// format := append(ws[:ctx.level*2], "%s"...)
	fmt.Fprintf(ctx.buf, "%s", v)
}

func (ctx *Context) blockIntro() {
	if !ctx.firstRule {
		panic("intro twice")
	}
	ctx.firstRule = false
	// Will probably need better logic around this
	sels := strings.Join(ctx.sels, " ")
	fmt.Fprintf(ctx.buf, "%s {\n", sels)
}

func (ctx *Context) blockOutro() {
	if ctx.firstRule {
		return
		fmt.Println("empty rule?")
	}
	ctx.firstRule = true
	ctx.sels = ctx.sels[:len(ctx.sels)-1]
	if len(ctx.sels) != ctx.level {
		panic(fmt.Sprintf("level mismatch lvl:%d sels:%d",
			ctx.level,
			len(ctx.sels)))
	}
	fmt.Fprintf(ctx.buf, " }\n")
}

func (ctx *Context) Visit(node ast.Node) ast.Visitor {
	switch v := node.(type) {
	case *ast.BlockStmt:
		ctx.level = ctx.level + 1
		ctx.firstRule = true
		for _, node := range v.List {
			ast.Walk(ctx, node)
		}
		ctx.level = ctx.level - 1
		ctx.blockOutro()
		ctx.firstRule = true
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
	case *ast.PropValueSpec:
		ctx.printers[propSpec](ctx, v)
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
	case *ast.BasicLit:
		ctx.printers[expr](ctx, v)
	case nil:

	default:
		fmt.Printf("add printer for: %T\n", v)
		fmt.Printf("% #v\n", v)
	}
	return ctx
}

var (
	ident     *ast.Ident
	expr      ast.Expr
	declStmt  *ast.DeclStmt
	valueSpec *ast.ValueSpec
	ruleSpec  *ast.RuleSpec
	selDecl   *ast.SelDecl
	selStmt   *ast.SelStmt
	propSpec  *ast.PropValueSpec
	typeSpec  *ast.TypeSpec
)

func (ctx *Context) Init() {
	ctx.buf = bytes.NewBuffer(nil)
	ctx.printers = make(map[ast.Node]func(*Context, ast.Node))
	ctx.printers[valueSpec] = visitValueSpec

	ctx.printers[ident] = printIdent
	ctx.printers[declStmt] = printDecl
	ctx.printers[ruleSpec] = printRuleSpec
	ctx.printers[selDecl] = printSelDecl
	ctx.printers[selStmt] = printSelStmt
	ctx.printers[propSpec] = printPropValueSpec
	ctx.printers[expr] = printExpr
	ctx.typ = NewScope(empty)
	// ctx.printers[typeSpec] = visitTypeSpec
	// assign printers
}

func printExpr(ctx *Context, n ast.Node) {
	switch v := n.(type) {
	case *ast.BasicLit:
		ctx.out(v.Value)
	}
}

func printSelStmt(ctx *Context, n ast.Node) {
	stmt := n.(*ast.SelStmt)
	ctx.sels = append(ctx.sels, stmt.Name.String())
}

func printSelDecl(ctx *Context, n ast.Node) {
	decl := n.(*ast.SelDecl)
	ctx.sels = append(ctx.sels, decl.Name.String())
}

func printRuleSpec(ctx *Context, n ast.Node) {
	// Inspect the sel buffer and dump it
	// We'll also need to track what level was last dumped
	// so selectors don't get printed twice
	if ctx.firstRule {
		ctx.blockIntro()
	}
	spec := n.(*ast.RuleSpec)
	ctx.out(fmt.Sprintf("  %s: ", spec.Name))
}

func printPropValueSpec(ctx *Context, n ast.Node) {
	spec := n.(*ast.PropValueSpec)
	ctx.out(spec.Name.String() + ";")
}

// Variable declarations
func visitValueSpec(ctx *Context, n ast.Node) {
	spec := n.(*ast.ValueSpec)

	names := make([]string, len(spec.Names))
	for i, nm := range spec.Names {
		names[i] = nm.Name
	}

	if len(spec.Values) > 0 {
		ctx.typ.Set(names[0], simplifyExprs(spec.Values))
	} else {
		ctx.out(fmt.Sprintf("%s;", ctx.typ.Get(names[0])))
	}
	// ctx.out(fmt.Sprintf("%s;", strings.Join(names, " ")))
}

func simplifyExprs(exprs []ast.Expr) string {
	var sum string
	for _, expr := range exprs {
		switch v := expr.(type) {
		case *ast.Ident:
			sum += v.Name
		case *ast.BasicLit:
			sum += v.Value
		default:
			log.Fatalf("unhandled expr: % #v\n", v)
		}
	}
	return sum
}

func printDecl(ctx *Context, ident ast.Node) {
	// I think... nothing to print we'll see
}

func printIdent(ctx *Context, ident ast.Node) {
	fmt.Printf("% #v\n", ident)
	fmt.Fprintf(ctx.buf, "%s", ident)
}
