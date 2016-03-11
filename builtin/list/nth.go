package list

import (
	"log"
	"strconv"

	"github.com/wellington/sass/ast"
	"github.com/wellington/sass/builtin"
)

func init() {
	builtin.Reg("nth($list, $pos)", nth)
}

func nth(call *ast.CallExpr, args ...ast.Expr) (ast.Expr, error) {

	in := args[0]
	_ = in
	epos := args[1]
	spos := epos.(*ast.BasicLit)
	pos, err := strconv.Atoi(spos.Value)
	if err != nil {
		return nil, err
	}

	// Sass is 1 index, of course
	pos = pos - 1

	for _, arg := range call.Args {
		log.Printf("% #v\n", arg)
	}
	log.Fatalf("% #v\n", call.Args)
	return nil, nil
}
