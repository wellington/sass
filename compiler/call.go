package compiler

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/wellington/sass/parser"
	"github.com/wellington/sass/token"

	"github.com/wellington/sass/ast"
)

var ErrNotFound = errors.New("function does not exist")

type CallHandler func(args []*ast.BasicLit) (*ast.BasicLit, error)

type call struct {
	name string
	args []*ast.KeyValueExpr
	ch   CallHandler
}

func (c *call) Pos(key *ast.Ident) int {
	for i, arg := range c.args {
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

var funcs map[string]call = make(map[string]call)

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
		d.c.name = v.Fun.(*ast.BasicLit).Value
		for _, arg := range v.Args {
			switch v := arg.(type) {
			case *ast.KeyValueExpr:
				d.c.args = append(d.c.args, v)
			case *ast.Ident:
				d.c.args = append(d.c.args, &ast.KeyValueExpr{
					Key: v,
				})
			default:
				log.Fatalf("failed to parse arg % #v\n", v)
			}
		}
		return nil
	case nil:
		return nil
	default:
		log.Fatalf("illegal walk % #v\n", v)
	}
	return d
}

func walkFunctionDescription(node ast.Node) call {
	var args []*ast.Ident
	_ = args
	return call{}
}

func Register(s string, ch CallHandler) {
	fset := token.NewFileSet()
	pf, err := parser.ParseFile(fset, "", s, 0)
	if err != nil {
		if !strings.HasSuffix(err.Error(), "expected ';', found 'EOF'") {
			log.Fatal(err)
		}
	}
	d := &desc{c: call{ch: ch}}
	// ast.Print(fset, pf.Decls[0])
	ast.Walk(d, pf.Decls[0])
	if d.err != nil {
		log.Fatal("failed to parse func description", d.err)
	}
	if _, ok := funcs[d.c.name]; ok {
		log.Println("already registered", d.c.name)
	}
	funcs[d.c.name] = d.c
}

// This might not be enough
func evaluateCall(expr *ast.CallExpr) (*ast.BasicLit, error) {

	ident := expr.Fun.(*ast.Ident)
	fn, ok := funcs[ident.Name]
	if !ok {
		return notfoundCall(expr), nil
	}

	callargs := make([]*ast.BasicLit, len(fn.args))
	for i := range fn.args {
		callargs[i] = fn.args[i].Value.(*ast.BasicLit)
	}
	// Verify args and convert to BasicLit before passing along
	for i, arg := range expr.Args {
		switch v := arg.(type) {
		case *ast.BasicLit:
			callargs[i] = v
		case *ast.KeyValueExpr:
			pos := fn.Pos(v.Key.(*ast.Ident))
			callargs[pos] = v.Value.(*ast.BasicLit)
		case *ast.Ident:
			assign := v.Obj.Decl.(*ast.AssignStmt)
			fmt.Printf("% #v\n", assign.Rhs[0])
			callargs[i] = assign.Rhs[0].(*ast.BasicLit)
		default:
			log.Fatalf("eval call unsupported % #v\n", v)
		}
	}

	return fn.ch(callargs)
}

// there's no such thing as a failure in Sass. Resolve idents in callexpr
// and return result as BasicLit
func notfoundCall(call *ast.CallExpr) (lit *ast.BasicLit) {

	return
}
