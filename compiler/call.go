package compiler

import (
	"errors"
	"log"
	"regexp"
	"strings"

	"github.com/wellington/sass/parser"
	"github.com/wellington/sass/token"

	"github.com/wellington/sass/ast"
	"github.com/wellington/sass/compiler/builtin"
)

var ErrNotFound = errors.New("function does not exist")

type CallHandler func(args []ast.Expr) (*ast.BasicLit, error)

type call struct {
	name string
	args []*ast.KeyValueExpr
	ch   CallHandler
}

var funcs map[string]call = make(map[string]call)

var (
	regName = regexp.MustCompile("^\\s+").FindString
	regArgs = regexp.MustCompile("\\$\\s+").FindAllString
	regDef  = regexp.MustCompile(":\\s+,").FindAllString
)

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
	name := regName(s)
	if _, ok := funcs[name]; ok {
		// panic("already registered: " + name)
	}
	args := regArgs(s, -1)
	_ = args
	fset := token.NewFileSet()
	pf, err := parser.ParseFile(fset, "", s, 0)
	if err != nil {
		if !strings.HasSuffix(err.Error(), "expected ';', found 'EOF'") {
			log.Fatal(err)
		}
	}
	d := &desc{c: call{}}
	// ast.Print(fset, pf.Decls[0])
	ast.Walk(d, pf.Decls[0])
	if d.err != nil {
		log.Fatal("failed to parse func description", d.err)
	}
	funcs[name] = d.c
}

// This might not be enough
func evaluateCall(expr *ast.CallExpr) (*ast.BasicLit, error) {

	ident := expr.Fun.(*ast.Ident)

	fn, ok := funcs[ident.Name]
	if !ok {
		return notfoundCall(expr), nil
	}

	// Verify args and convert to KeyValueExpr before passing along

	return fn.ch(expr.Args)
}

// there's no such thing as a failure in Sass. Resolve idents in callexpr
// and return result as BasicLit
func notfoundCall(call *ast.CallExpr) (lit *ast.BasicLit) {

	return
}

func init() {
	Register("rgb($red:0, $green:0, $blue:0)", builtin.RGB)
	Register("rgba($red:0, $green:0, $blue:0, $alpha:0)", builtin.RGBA)
}
