package compiler

import (
	"fmt"
	"log"

	"github.com/wellington/sass/ast"
)

func printInclude(ctx *Context, n ast.Node) {
	// Add new scope, register args
	stmt := n.(*ast.IncludeStmt)

	name := stmt.Spec.Name.String()
	var params []*ast.Field
	if stmt.Spec.Params != nil {
		params = stmt.Spec.Params.List
	}
	numargs := stmt.Spec.Params.NumFields()

	mix, err := ctx.scope.Mixin(name, numargs)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("include", name)
	ctx.scope = NewScope(ctx.scope)
	mixargs := mix.fn.Type.Params.List
	for i := range mixargs {
		// Param passed by include
		var param *ast.Field
		if len(params) > i {
			param = params[i]
		}
		arg := mixargs[i].Type.(*ast.BasicLit).Value
		val, ok := param.Type.(*ast.Ident)
		if param != nil && ok {
			fmt.Printf("var: % #v\nval: % #v\n",
				arg,
				val.Name,
			)
			ctx.scope.Set(arg, val.Name)
		} else {
			fmt.Printf("var: % #v\nNOVAL: % #v\n",
				mixargs[i].Type.(*ast.BasicLit).Value,
				param.Type,
			)
		}
	}
	// ctx.typ.Set(string, interface{})
	for _, stmt := range mix.fn.Body.List {
		ast.Walk(ctx, stmt)
	}
	ctx.scope = CloseScope(ctx.scope)

	// Exit new scope, removing args
}
