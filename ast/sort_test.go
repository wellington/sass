package ast

import (
	"reflect"
	"testing"
)

func TestSort(t *testing.T) {
	startcmt := &CommStmt{
		Group: &CommentGroup{
			List: []*Comment{
				&Comment{Text: "/* start */"},
			},
		},
	}
	endcmt := &CommStmt{
		Group: &CommentGroup{
			List: []*Comment{
				&Comment{Text: "/* end */"},
			},
		},
	}

	list := []Stmt{
		&SelStmt{Name: &Ident{Name: "div"}},
		startcmt,
		&DeclStmt{},
		endcmt,
		&IncludeStmt{},
		&AssignStmt{},
	}

	sorted := []Stmt{
		&CommStmt{},
		&DeclStmt{},
		&CommStmt{},
		&IncludeStmt{},
		&AssignStmt{},
		&SelStmt{},
	}

	SortStatements(list)

	for i := range list {
		l := reflect.TypeOf(list[i])
		if e := reflect.TypeOf(sorted[i]); e != l {
			t.Errorf("%d got: %v wanted: %v\n", i, l, e)
		}
	}

}
