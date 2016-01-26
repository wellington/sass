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
	fmt.Println("========\ninclude", name)
	ctx.scope = NewScope(ctx.scope)
	fmt.Printf("% #v\n\n\n", mix.fn.Type.Params.List[0].Type)
	mixargs := mix.fn.Type.Params.List
	for i := range mixargs {
		// Param passed by include
		var param *ast.Field
		if len(params) > i {
			param = params[i]
		}
		key := mixargs[i].Type.(*ast.BasicLit)

		switch v := param.Type.(type) {
		case *ast.KeyValueExpr:
			// Key args specify their argument, so use their key
			// instead of the mixins argument for this position
			// Params with defaults
			key = v.Key.(*ast.BasicLit)
			val := v.Value.(*ast.Ident)
			ctx.scope.Set(key.Value, val.Name)
		case *ast.Ident:
			ctx.scope.Set(key.Value, v.Name)
		default:
			fmt.Printf("dropped param: % #v\n", v)
		}
	}
	if len(params) > len(mixargs) {
		fmt.Printf("dropped extra params: % #v\n", params[len(mixargs):])
	}
	for _, stmt := range mix.fn.Body.List {
		ast.Walk(ctx, stmt)
	}
	ctx.scope = CloseScope(ctx.scope)

	// Exit new scope, removing args
}
