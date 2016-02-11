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
		log.Fatalf("illegal walk % #v\n", v)
	}
	return d
}

var builtins = make(map[string]call)

func init() {
	builtin.BindRegister(register)
}

func register(s string, ch builtin.CallHandler) {
	fset := token.NewFileSet()
	pf, err := ParseFile(fset, "", s, 0)
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
		return notfoundCall(expr), nil
	}
	fmt.Printf("ident % #v\n", ident)
	callargs := make([]*ast.BasicLit, len(fn.args))
	for i := range fn.args {
		expr := fn.args[i].Value
		if expr != nil {
			callargs[i] = expr.(*ast.BasicLit)
		}
	}
	var argpos int
	// Verify args and convert to BasicLit before passing along
	for i, arg := range expr.Args {
		if argpos < i {
			argpos = i
		}
		fmt.Printf("arg[%d] % #v\n", i, arg)
		switch v := arg.(type) {
		case *ast.BasicLit:
			callargs[argpos] = v
		case *ast.KeyValueExpr:
			fmt.Printf("k: % #v v: % #v\n", v.Key, v.Value)
			pos := fn.Pos(v.Key.(*ast.Ident))
			fmt.Println("found arg at pos:", pos)
			callargs[pos] = v.Value.(*ast.BasicLit)
		case *ast.Ident:
			assign := v.Obj.Decl.(*ast.AssignStmt)
			fmt.Printf("% #v\n", assign.Rhs[0])
			switch v := assign.Rhs[0].(type) {
			case *ast.BasicLit:
				callargs[argpos] = v
			case *ast.CallExpr:
				// variable pointing to a function
				for i := range v.Args {
					callargs[argpos] = v.Args[i].(*ast.BasicLit)
					argpos++
				}
			}
		case *ast.CallExpr:
			// Nested function call
			lit, err := evaluateCall(v)
			if err != nil {
				return nil, err
			}
			fmt.Println(argpos)
			fmt.Println("len", len(callargs))
			callargs[argpos] = lit
		default:
			log.Fatalf("eval call unsupported % #v\n", v)
		}

	}
	fmt.Println("callargs")
	for i := range callargs {
		fmt.Printf("%d: % #v\n", i, callargs[i])
	}
	return fn.ch(callargs)
}

// there's no such thing as a failure in Sass. Resolve idents in callexpr
// and return result as BasicLit
func notfoundCall(call *ast.CallExpr) (lit *ast.BasicLit) {
	return
}
