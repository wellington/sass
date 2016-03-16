package parser

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/wellington/sass/ast"
	"github.com/wellington/sass/builtin"
	"github.com/wellington/sass/token"

	// Include defined builtins
	_ "github.com/wellington/sass/builtin/colors"
	_ "github.com/wellington/sass/builtin/introspect"
	_ "github.com/wellington/sass/builtin/list"
	_ "github.com/wellington/sass/builtin/strops"
	_ "github.com/wellington/sass/builtin/url"
)

var ErrNotFound = errors.New("function does not exist")

type call struct {
	name   string
	params []*ast.KeyValueExpr
	ch     builtin.CallHandler
	handle builtin.CallHandle
}

func (c *call) Pos(key *ast.Ident) int {
	for i, arg := range c.params {
		switch v := arg.Key.(type) {
		case *ast.Ident:
			if key.Name == v.Name {
				return i
			}
		default:
			log.Fatalf("failed to lookup key % #v\n", v)
		}
	}
	return -1
}

type desc struct {
	err error
	c   call
}

func (d *desc) Visit(node ast.Node) ast.Visitor {
	switch v := node.(type) {
	case *ast.RuleSpec:
		for i := range v.Values {
			ast.Walk(d, v.Values[i])
		}
	case *ast.GenDecl:
		for _, spec := range v.Specs {
			ast.Walk(d, spec)
		}
		return nil
	case *ast.CallExpr:
		d.c.name = v.Fun.(*ast.Ident).Name
		for _, arg := range v.Args {
			switch v := arg.(type) {
			case *ast.KeyValueExpr:
				d.c.params = append(d.c.params, v)
			case *ast.Ident:
				d.c.params = append(d.c.params, &ast.KeyValueExpr{
					Key: v,
				})
			default:
				ast.Print(token.NewFileSet(), v)
				panic(fmt.Errorf("%s failed to parse arg % #v\n",
					d.c.name, v))
			}
		}
		return nil
	case nil:
		return nil
	default:
		panic(fmt.Errorf("illegal walk % #v\n", v))
	}
	return d
}

var builtins = make(map[string]call)

func init() {
	builtin.BindRegister(register)
}

func register(s string, ch builtin.CallHandler, h builtin.CallHandle) {
	fset := token.NewFileSet()
	pf, err := ParseFile(fset, "", s, FuncOnly)
	if err != nil {
		if !strings.HasSuffix(err.Error(), "expected ';', found 'EOF'") {
			log.Fatal(err)
		}
	}
	d := &desc{c: call{
		ch:     ch,
		handle: h,
	}}
	// ast.Print(fset, pf.Decls[0])
	ast.Walk(d, pf.Decls[0])
	if d.err != nil {
		log.Fatal("failed to parse func description", d.err)
	}
	if _, ok := builtins[d.c.name]; ok {
		log.Println("already registered", d.c.name)
	}
	builtins[d.c.name] = d.c
}

func exprToLit(x ast.Expr) (lit *ast.BasicLit, ok bool) {
	switch v := x.(type) {
	case *ast.BadExpr:
		panic("")
	case *ast.BasicLit:
		return v, true
	case *ast.Ident:
		if v.Obj == nil {
			panic(fmt.Errorf("ident unresolved: % #v\n",
				v))
		}
		assign := v.Obj.Decl.(*ast.AssignStmt)
		if len(assign.Rhs) > 0 {
			return
		}
		return exprToLit(assign.Rhs[0])
		// Resolving an assignment should also update the ctx.
		switch v := assign.Rhs[0].(type) {
		case *ast.BasicLit:
			return v, true
			// incoming[i] = v
			// callargs[argpos] = v
		case *ast.CallExpr:
			// incoming[i] = v
			// callargs[argpos] = v.Resolved
			return exprToLit(v.Resolved)
		}
	case *ast.StringExpr:
		if len(v.List) > 1 {
			log.Fatalf("% #v\n", v.List)
		}
		val, ok := v.List[0].(*ast.BasicLit)
		if !ok {
			// try again on interp
			val = v.List[0].(*ast.Interp).Obj.Decl.(*ast.BasicLit)
		}
		return &ast.BasicLit{
			Kind:  token.QSTRING,
			Value: val.Value,
		}, true
	case *ast.BinaryExpr:
		log.Fatalf("% #v\n", v)
	case *ast.CallExpr:
		// Nested function call
		x, err := evaluateCall(v)
		if err != nil {
			log.Printf("I need parser context: %s\n", err)
			return
		}
		return exprToLit(x)
		// callargs[argpos] = lit
	case *ast.ListLit:
		// During expr simplification, list are just string
		delim := " "
		if v.Comma {
			delim = ", "
		}
		ss := make([]string, len(v.Value))
		for i := range v.Value {
			ss[i] = v.Value[i].(*ast.BasicLit).Value
		}
		return &ast.BasicLit{
			Value:    strings.Join(ss, delim),
			ValuePos: v.Pos(),
		}, true
	case *ast.Interp:
		if v.Obj == nil {
			ast.Print(token.NewFileSet(), v)
			log.Fatalf("nil")
		}
		l, ok := v.Obj.Decl.(*ast.BasicLit)
		if ok {
			return l, true
		}
		// callargs[argpos] = v.Obj.Decl.(*ast.BasicLit)
	default:
		log.Fatalf("eval call unsupported % #v\n", v)
	}
	return
}

// This might not be enough
func evaluateCall(expr *ast.CallExpr) (ast.Expr, error) {
	ident := expr.Fun.(*ast.Ident)
	name := ident.Name
	fn, ok := builtins[name]
	if !ok {
		return nil, fmt.Errorf("func %s was not found", name)
	}

	// Walk through the function
	// These should be processed at registration time
	callargs := make([]ast.Expr, len(fn.params))
	for i := range fn.params {
		expr := fn.params[i].Value
		// if expr != nil {
		// 	callargs[i] = expr.(*ast.BasicLit)
		// }
		callargs[i] = expr
	}
	var argpos int
	incoming := expr.Args

	// Verify args and convert to BasicLit before passing along
	if len(callargs) < len(incoming) {
		for i, p := range incoming {
			lit, ok := p.(*ast.BasicLit)
			if !ok {
				log.Fatalf("failed to convert to lit % #v\n", p)
			}
			log.Printf("inc %d %s:% #v\n", i, lit.Kind, p)
		}
		return nil, fmt.Errorf("mismatched arg count %s got: %d wanted: %d",
			name, len(incoming), len(callargs))
	}

	for i, arg := range incoming {
		if argpos < i {
			argpos = i
		}
		fmt.Printf("%d % #v\n", i, arg)
		switch v := arg.(type) {
		case *ast.KeyValueExpr:
			pos := fn.Pos(v.Key.(*ast.Ident))
			callargs[pos] = v.Value.(*ast.BasicLit)
		case *ast.ListLit:
			callargs[argpos] = v
		case *ast.Ident:
			if v.Obj != nil {
				ass := v.Obj.Decl.(*ast.AssignStmt)
				fmt.Printf("ass % #v\n", ass)
				callargs[argpos] = ass.Rhs[0]
			} else {
				callargs[argpos] = v
			}

		default:
			lit, ok := exprToLit(v)
			if ok {
				callargs[argpos] = lit
			} else {
				log.Fatalf("boom: % #v\n", v)
			}
		}
	}
	if fn.ch != nil {
		lits := make([]*ast.BasicLit, len(callargs))
		for i, x := range callargs {
			lits[i], ok = exprToLit(x)
			if !ok {
				log.Fatalf("litize % #v\n", x)
			}
		}
		return fn.ch(expr, lits...)
	}
	ast.Print(token.NewFileSet(), callargs)
	return fn.handle(expr, callargs...)
}
