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
	_ "github.com/wellington/sass/builtin/strops"
	_ "github.com/wellington/sass/builtin/url"
)

var ErrNotFound = errors.New("function does not exist")

type call struct {
	name string
	args []*ast.KeyValueExpr
	ch   builtin.CallHandler
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
		panic(fmt.Errorf("illegal walk % #v\n", v))
	}
	return d
}

var builtins = make(map[string]call)

func init() {
	builtin.BindRegister(register)
}

func register(s string, ch builtin.CallHandler) {
	fset := token.NewFileSet()
	pf, err := ParseFile(fset, "", s, FuncOnly)
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
	if _, ok := builtins[d.c.name]; ok {
		log.Println("already registered", d.c.name)
	}
	builtins[d.c.name] = d.c
}

// This might not be enough
func evaluateCall(expr *ast.CallExpr) (*ast.BasicLit, error) {
	ident := expr.Fun.(*ast.Ident)
	name := ident.Name

	fn, ok := builtins[name]
	if !ok {
		return nil, fmt.Errorf("func %s was not found", name)
	}

	callargs := make([]*ast.BasicLit, len(fn.args))
	for i := range fn.args {
		expr := fn.args[i].Value
		if expr != nil {
			callargs[i] = expr.(*ast.BasicLit)
		}
	}
	var argpos int
	incoming := expr.Args
	// Verify args and convert to BasicLit before passing along
	if len(callargs) < len(incoming) {
		for _, p := range incoming {
			log.Printf("inc % #v\n", p)
		}
		return nil, fmt.Errorf("mismatched arg count %s got: %d wanted: %d",
			name, len(incoming), len(callargs))
	}
	for i, arg := range incoming {
		if argpos < i {
			argpos = i
		}

		switch v := arg.(type) {
		case *ast.BasicLit:
			callargs[argpos] = v
		case *ast.KeyValueExpr:
			pos := fn.Pos(v.Key.(*ast.Ident))
			callargs[pos] = v.Value.(*ast.BasicLit)
		case *ast.Ident:
			if v.Obj == nil {
				panic(fmt.Errorf("ident unresolved: % #v\n",
					v))
			}
			assign := v.Obj.Decl.(*ast.AssignStmt)
			// Resolving an assignment should also update the ctx.
			switch v := assign.Rhs[0].(type) {
			case *ast.BasicLit:
				incoming[i] = v
				callargs[argpos] = v
			case *ast.CallExpr:
				incoming[i] = v
				callargs[argpos] = v.Resolved
			}
		case *ast.BinaryExpr:
			log.Fatalf("% #v\n", v)
		case *ast.CallExpr:
			// Nested function call
			lit, err := evaluateCall(v)
			if err != nil {
				return nil, err
			}
			callargs[argpos] = lit
		default:
			log.Fatalf("eval call unsupported % #v\n", v)
		}

	}
	return fn.ch(expr, callargs...)
}
