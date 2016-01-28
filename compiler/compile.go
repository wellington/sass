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
	sels      [][]*ast.Ident
	firstRule bool
	level     int
	printers  map[ast.Node]func(*Context, ast.Node)

	scope Scope
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
	pf, err := parser.ParseFile(fset, path, nil, parser.ParseComments) //|parser.Trace)
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
func (ctx *Context) out(v string) {
	fr, _ := utf8.DecodeRuneInString(v)
	if fr == '\n' {
		fmt.Fprintf(ctx.buf, v)
		return
	}
	ws := []byte("                                              ")
	lvl := ctx.level

	format := append(ws[:lvl*2], "%s"...)
	fmt.Fprintf(ctx.buf, string(format), v)
}

// This needs a new name, it prints on every stmt
func (ctx *Context) blockIntro() {

	// this isn't a new block
	if !ctx.firstRule {
		fmt.Fprint(ctx.buf, "\n")
		return
	}

	ctx.firstRule = false

	// Only print newlines if there is text in the buffer
	if ctx.buf.Len() > 0 {
		if ctx.level == 0 {
			fmt.Fprint(ctx.buf, "\n")
		} else {

		}
	}

	// Will probably need better logic around this
	sels := strings.Join(ctx.combineSels(), ", ")
	ctx.out(fmt.Sprintf("%s {\n", sels))
}

func (ctx *Context) blockOutro() {
	// Remove the innermost selector scope
	if len(ctx.sels) > 0 {
		ctx.sels = ctx.sels[:len(ctx.sels)-1]
	}
	// Don't print } if there are no rules at this level
	if ctx.firstRule {
		return
	}

	ctx.firstRule = true
	// if !skipParen {
	fmt.Fprintf(ctx.buf, " }\n")
	// }
}

func (ctx *Context) combineSels() []string {
	return walkSelectors(ctx.sels)
}

func walkSelectors(in [][]*ast.Ident) []string {
	if len(in) == 1 {
		ret := make([]string, len(in[0]))
		for i, ident := range in[0] {
			ret[i] = ident.String()
		}
		return ret
	}

	d := in[0]
	w := walkSelectors(in[1:])
	var ret []string
	for i := 0; i < len(d); i++ {
		for j := 0; j < len(w); j++ {
			ret = append(ret, d[i].String()+" "+w[j])
		}
	}
	return ret
}

func (ctx *Context) Visit(node ast.Node) ast.Visitor {

	var key ast.Node
	switch v := node.(type) {
	case *ast.BlockStmt:
		if ctx.scope.RuleLen() > 0 {
			ctx.level = ctx.level + 1
			if !ctx.firstRule {
				fmt.Fprintf(ctx.buf, " }\n")
			}
		}
		ctx.scope = NewScope(ctx.scope)
		ctx.firstRule = true
		for _, node := range v.List {
			ast.Walk(ctx, node)
		}
		if ctx.level > 0 {
			ctx.level = ctx.level - 1
		}
		ctx.scope = CloseScope(ctx.scope)
		ctx.blockOutro()
		ctx.firstRule = true
		// ast.Walk(ctx, v.List)
		// fmt.Fprintf(ctx.buf, "}")
		return nil
	case *ast.SelDecl:
		key = selDecl
	case *ast.File, *ast.GenDecl, *ast.Value:
		// Nothing to print for these
	case *ast.Ident:
		// The first IDENT is always the filename, just preserve
		// it somewhere
		key = ident
	case *ast.PropValueSpec:
		key = propSpec
	case *ast.DeclStmt:
		key = declStmt
	case *ast.IncludeSpec:
		key = includeSpec
	case *ast.ValueSpec:
		key = valueSpec
	case *ast.RuleSpec:
		key = ruleSpec
	case *ast.SelStmt:
		// We will need to combine parent selectors
		// while printing these
		key = selStmt
		// Nothing to do
	case *ast.CommStmt:
	case *ast.CommentGroup:
	case *ast.Comment:
		key = comment
	case *ast.FuncDecl:
		ctx.printers[funcDecl](ctx, node)
		// Do not traverse mixins in the regular context
		return nil
	case *ast.BasicLit:
		return ctx
	case nil:
		return ctx
	case *ast.EmptyStmt:
	case *ast.AssignStmt:
		key = assignStmt
	default:
		fmt.Printf("add printer for: %T\n", v)
		fmt.Printf("% #v\n", v)
	}
	ctx.printers[key](ctx, node)
	return ctx
}

var (
	ident       *ast.Ident
	expr        ast.Expr
	declStmt    *ast.DeclStmt
	assignStmt  *ast.AssignStmt
	valueSpec   *ast.ValueSpec
	ruleSpec    *ast.RuleSpec
	selDecl     *ast.SelDecl
	selStmt     *ast.SelStmt
	propSpec    *ast.PropValueSpec
	typeSpec    *ast.TypeSpec
	comment     *ast.Comment
	funcDecl    *ast.FuncDecl
	includeSpec *ast.IncludeSpec
)

func (ctx *Context) Init() {
	ctx.buf = bytes.NewBuffer(nil)
	ctx.printers = make(map[ast.Node]func(*Context, ast.Node))
	ctx.printers[valueSpec] = visitValueSpec
	ctx.printers[funcDecl] = visitFunc
	ctx.printers[assignStmt] = visitAssignStmt

	ctx.printers[ident] = printIdent
	ctx.printers[includeSpec] = printInclude
	ctx.printers[declStmt] = printDecl
	ctx.printers[ruleSpec] = printRuleSpec
	ctx.printers[selDecl] = printSelDecl
	ctx.printers[selStmt] = printSelStmt
	ctx.printers[propSpec] = printPropValueSpec
	ctx.printers[expr] = printExpr
	ctx.printers[comment] = printComment
	ctx.scope = NewScope(empty)
	// ctx.printers[typeSpec] = visitTypeSpec
	// assign printers
}

func printComment(ctx *Context, n ast.Node) {
	ctx.blockIntro()
	cmt := n.(*ast.Comment)
	// These additional spaces should be handled by out()
	ctx.out("  " + cmt.Text)
}

func printExpr(ctx *Context, n ast.Node) {
	switch v := n.(type) {
	case *ast.File:
	case *ast.BasicLit:
		fmt.Fprintf(ctx.buf, "%s;", v.Value)
	case *ast.Value:
	case *ast.GenDecl:
		// Ignoring these for some reason
	default:
		// fmt.Printf("unmatched expr %T: % #v\n", v, v)
	}
}

func (ctx *Context) storeSelector(idents []*ast.Ident) {
	ctx.sels = append(ctx.sels, idents)
}

func printSelStmt(ctx *Context, n ast.Node) {
	stmt := n.(*ast.SelStmt)
	ctx.storeSelector(stmt.Names)
}

func printSelDecl(ctx *Context, n ast.Node) {
	decl := n.(*ast.SelDecl)
	ctx.storeSelector(decl.Names)
}

func printRuleSpec(ctx *Context, n ast.Node) {
	// Inspect the sel buffer and dump it
	// Also need to track what level was last dumped
	// so selectors don't get printed twice
	ctx.blockIntro()

	spec := n.(*ast.RuleSpec)
	ctx.scope.RuleAdd(spec)
	ctx.out(fmt.Sprintf("  %s: ", spec.Name))
	fmt.Fprintf(ctx.buf, "%s;", simplifyExprs(ctx, spec.Values))
}

func printPropValueSpec(ctx *Context, n ast.Node) {
	spec := n.(*ast.PropValueSpec)
	fmt.Fprintf(ctx.buf, spec.Name.String()+";")
}

// Variable assignments inside blocks ie. mixins
func visitAssignStmt(ctx *Context, n ast.Node) {
	return
	stmt := n.(*ast.AssignStmt)
	var key, val *ast.Ident
	_, _ = key, val
	switch v := stmt.Lhs[0].(type) {
	case *ast.Ident:
		key = v
	default:
		log.Fatalf("unsupported key: % #v", v)
	}

	switch v := stmt.Rhs[0].(type) {
	case *ast.Ident:
		val = v
	default:
		log.Fatalf("unsupported key: % #v", v)
	}

}

// Variable declarations
func visitValueSpec(ctx *Context, n ast.Node) {
	return
}

func exprString(expr ast.Expr) string {
	switch v := (expr).(type) {
	case *ast.Ident:
		return v.String()
	case *ast.BasicLit:
		return v.Value
	default:
		panic(fmt.Sprintf("exprString: %T", v))
	}
	return ""
}

func calculateExprs(ctx *Context, bin *ast.BinaryExpr) (string, error) {
	x := bin.X
	y := bin.Y

	// If X or Y are Ident, append as strings
	_, xok := x.(*ast.Ident)
	_, yok := y.(*ast.Ident)
	if xok || yok {
		var s string
		switch bin.Op {
		case token.ADD:
			s = exprString(x) + exprString(y)
		default:
			s = exprString(x) + bin.Op.String() + exprString(y)
		}
		return s, nil
	}

	var err error
	// Convert CallExpr to BasicLit
	if cx, ok := x.(*ast.CallExpr); ok {
		x, err = evaluateCall(cx)
		if err != nil {
			return "", fmt.Errorf("error execing %s: %s",
				cx.Fun, err)
		}
	}
	if cy, ok := y.(*ast.CallExpr); ok {
		y, err = evaluateCall(cy)
		if err != nil {
			return "", fmt.Errorf("error execing %s: %s",
				cy.Fun, err)
		}
	}

	if err != nil {
		return "", err
	}

	bx := x.(*ast.BasicLit)
	by := y.(*ast.BasicLit)
	// Attempt color math
	if bx.Kind == token.COLOR {
		z := bx.Op(bin.Op, by)
		if z == nil {
			panic(fmt.Sprintf("invalid return op: %q x: % #v y: % #v",
				bin.Op, bx, by,
			))
		}
		return z.Value, nil
	}

	// We're looking at INT and non-INT, treat as strings
	if bx.Kind == token.INT && by.Kind != token.INT {
		// Treat everything as strings
		return bx.Value + bin.Op.String() + by.Value, nil
	}

	return "", nil
}

func simplifyExprs(ctx *Context, exprs []ast.Expr) string {

	var sums []string
	for _, expr := range exprs {
		switch v := expr.(type) {
		case *ast.Value:
			panic("ast.Value")
		case *ast.BinaryExpr:
			s, err := calculateExprs(ctx, v)
			if err != nil {
				log.Fatal(err)
			}
			sums = append(sums, s)
		case *ast.CallExpr:
			s, err := evaluateCall(v)
			if err != nil {
				log.Fatalf("failed to call '%s': %s", v.Fun, err)
			}
			sums = append(sums, s.Value)
		case *ast.ParenExpr:
			sums = append(sums, simplifyExprs(ctx, []ast.Expr{v.X}))
		case *ast.Ident:
			if v.Obj == nil {
				sums = append(sums, v.Name)
				continue
			}
			fmt.Printf("printing (%p) % #v\n", v, v)
			switch vv := v.Obj.Decl.(type) {
			case *ast.Ident:
				sums = append(sums, vv.Name)
			case *ast.ValueSpec:
				var s []string
				for i := range vv.Values {
					if ident, ok := vv.Values[i].(*ast.Ident); ok {
						// If obj is set, resolve Obj and report
						if ident.Obj != nil {
							spec := ident.Obj.Decl.(*ast.ValueSpec)
							for _, val := range spec.Values {
								s = append(s, fmt.Sprintf("%s", val))
							}
						} else {
							// fmt.Printf("basic ident: % #v\n", ident)
							s = append(s, fmt.Sprintf("%s", ident))
						}
						continue
					}
					lit := vv.Values[i].(*ast.BasicLit)
					if len(lit.Value) > 0 {
						s = append(s, lit.Value)
					}
				}
				sums = append(sums, strings.Join(s, " "))
			case *ast.AssignStmt:
				// li := vv.Lhs[0].(*ast.Ident)
				// ri := vv.Rhs[0].(*ast.Ident)
				// fmt.Printf("lhs %s % #v\n", li)
				// fmt.Printf("rhs %s % #v\n", ri)
				sums = append(sums, simplifyExprs(ctx, vv.Rhs))
			default:
				fmt.Printf("unsupported VarDecl: % #v\n", vv)
				// Weird stuff here, let's just push the Ident in
				sums = append(sums, v.Name)
			}
		case *ast.BasicLit:
			switch v.Kind {
			case token.VAR:
				// s, ok := ctx.scope.Lookup(v.Value).(string)
				// if ok {
				// 	sums = append(sums, s)
				// }
			default:
				sums = append(sums, v.Value)
			}
		default:
			panic(fmt.Sprintf("unhandled expr: % #v\n", v))
		}
	}

	return strings.Join(sums, " ")
}

func printDecl(ctx *Context, node ast.Node) {
	// I think... nothing to print we'll see
}

func printIdent(ctx *Context, node ast.Node) {
	// ident := node.(*ast.Ident)
	// don't print these
	// fmt.Printf("ignoring % #v\n", ident)
}

func (c *Context) makeStrings(exprs []ast.Expr) (list []string) {
	list = make([]string, 0, len(exprs))
	for _, expr := range exprs {
		switch e := expr.(type) {
		case *ast.Ident:
			list = append(list, e.Name)
		}
	}
	return
}
