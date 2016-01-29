package ast

import (
	"fmt"
	"log"
)

// StmtCopy performs a deep copy of passed stmt
func StmtCopy(in Stmt) (out Stmt) {

	switch v := in.(type) {
	case *AssignStmt:
		stmt := &AssignStmt{
			Lhs: ExprsCopy(v.Lhs),
			Rhs: ExprsCopy(v.Rhs),
		}
		out = stmt
	case *DeclStmt:
		stmt := &DeclStmt{}
		if v.Decl != nil {
			stmt.Decl = DeclCopy(v.Decl)
		}
		out = stmt
	case *BlockStmt:
		block := &BlockStmt{}
		list := make([]Stmt, 0, len(v.List))
		for _, stmt := range v.List {
			list = append(list, StmtCopy(stmt))
		}
		block.List = list
		out = block
	case *SelStmt:
		stmt := &SelStmt{
			Name: NewIdent(v.Name.Name),
		}
		names := make([]*Ident, 0, len(v.Names))
		for _, ident := range v.Names {
			names = append(names, NewIdent(ident.Name))
		}
		stmt.Names = names
		stmt.Body = StmtCopy(v.Body).(*BlockStmt)
		out = stmt
	case *CommStmt:
		out = v
		return
	default:
		log.Fatalf("unsupported stmt copy %T: % #v\n", v, v)
	}
	fmt.Printf("StmtCopy (%p)% #v\n      ~> (%p)% #v\n",
		in, in, out, out)
	return
}

func ExprsCopy(in []Expr) []Expr {
	out := make([]Expr, 0, len(in))
	for i := range in {
		if in[i] != nil {
			out = append(out, ExprCopy(in[i]))
		}
	}
	return out
}

func ExprCopy(in Expr) (out Expr) {
	switch expr := in.(type) {
	case *Ident:
		out = NewIdent(expr.Name)
	}
	return
}

func SpecCopy(in Spec) (out Spec) {
	switch v := in.(type) {
	case *RuleSpec:
		spec := &RuleSpec{
			Name: NewIdent(v.Name.Name),
		}
		list := make([]Expr, 0, len(v.Values))
		for i := range v.Values {
			if v.Values[i] != nil {
				list = append(list, ExprCopy(v.Values[i]))
			}
		}
		spec.Values = list
		out = spec
	default:
		out = v
		log.Printf("unsupported spec copy %T: % #v\n", v, v)
		return
	}
	fmt.Printf("SpecCopy % #v\n      ~> % #v\n", in, out)
	return
}

func DeclCopy(in Decl) (out Decl) {
	switch v := in.(type) {
	case *GenDecl:
		decl := *v
		fmt.Printf("% #v\n", decl)
		list := make([]Spec, 0, len(decl.Specs))
		for i := range decl.Specs {
			if decl.Specs[i] != nil {
				list = append(list, SpecCopy(decl.Specs[i]))
			} else {
				fmt.Println("nil!")
			}
		}
		decl.Specs = list
		out = &decl
	default:
		log.Fatalf("unsupported decl copy %T: % #v\n", v, v)
	}
	fmt.Printf("DeclCopy (%p)% #v\n      ~> (%p)% #v\n",
		in, in, out, out)
	return
}
