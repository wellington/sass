package compiler

import (
	"fmt"

	"github.com/wellington/sass/ast"
	"github.com/wellington/sass/token"
)

func visitFunc(ctx *Context, n ast.Node) {
	fn := n.(*ast.FuncDecl)

	switch fn.Tok {
	case token.MIXIN:
		// make a copy of the context
		mixctx := *ctx
		// Need a way to walk through the body without
		// triggering blockIntro/Outro
		// for _, l := range fn.Body.List {
		// 	ast.Walk(&mixctx, l)
		// }
		RegisterMixin(fn.Name.String(),
			len(fn.Body.List),
			&MixFn{
				minArgs: len(fn.Body.List),
				ctx:     &mixctx,
				fn:      fn,
			})
	case token.FUNC:
	default:
		fmt.Printf("% #v\n", fn)
		panic("unsupported visitFunc")
	}

}
