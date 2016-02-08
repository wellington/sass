package ast

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/wellington/sass/token"
)

var (
	regEql = regexp.MustCompile("\\s*(\\*?=)\\s*").ReplaceAll
	regBkt = regexp.MustCompile("\\s*(\\[)\\s*(\\S+)\\s*(\\])").ReplaceAll
)

// Resolves walks selector operations removing nested Op by prepending X
// on Y.
func (stmt *SelStmt) Resolve() {

	s := &sel{
		parent: stmt.Parent,
		stmt:   stmt,
		prec:   token.LowestPrec + 1,
	}

	// This could be more efficient, it should inspect precision of
	// the top node
	for prec := token.UnaryPrec; prec > 1; prec-- {
		// Walk the selectors resolving ops found at the active
		// precision
		if s.parent != nil {
			s.inject = true
		}
		s.prec = prec

		Walk(s, s.stmt.Sel)
	}

	stmt.Resolved = stmt.Sel.(*BasicLit)
	Print(token.NewFileSet(), s.stmt.Sel)
}

type sel struct {
	stmt   *SelStmt
	parent *SelStmt
	prec   int    // Resolve each precendence in order
	stack  []Expr // Nesting stack
	inject bool   // inject parent to start
}

func (s *sel) Visit(node Node) Visitor {

	fmt.Printf("%d: % #v\n", s.prec, node)
	switch v := node.(type) {
	case *UnaryExpr:
		// Nesting
		fmt.Println("nested")
	case *BasicLit:
		if s.prec != 2 {
			return nil
		}
		delim := " "
		var val = v.Value
		fmt.Printf("prec %d inject? %t\n", s.prec, s.inject)
		if s.inject && s.parent != nil {
			val = s.parent.Resolved.Value + delim + v.Value
		}
		v.Value = val
		return nil
	case *BinaryExpr:
		switch v.Op {
		case token.NEST:
			if s.prec < 5 {
				panic(fmt.Errorf("invalid nest token: %s prec: %d", v.Op, s.prec))
			}
			if s.prec != 5 {
				return s
			}
		case token.ADD, token.GTR, token.TIL:
			if s.prec < 4 {
				panic(fmt.Errorf("invalid Op token: %s prec: %d", v.Op, s.prec))
			}
			if s.prec != 4 {
				return s
			}
			node = s.joinBinary(v)
		case token.COMMA:
			if s.prec < 3 {
				panic(fmt.Errorf("invalid group token: %s prec: %d", v.Op, s.prec))
			}
			if s.prec != 3 {
				// Reset parent injector
				s.inject = true
				Walk(s, v.X)
				// Reset parent injector
				s.inject = true
				Walk(s, v.Y)
				return nil
			}
			node = s.joinBinary(v)
		}
	}

	return s
}

func (s *sel) joinBinary(bin *BinaryExpr) *BasicLit {
	x, ok := bin.X.(*BasicLit)
	if !ok {
		x = s.joinBinary(bin.X.(*BinaryExpr))
	}
	y, ok := bin.Y.(*BasicLit)
	if !ok {
		y = s.joinBinary(bin.Y.(*BinaryExpr))
	}
	delim := " " // This will change with compiler mode

	vals := []string{x.Value, bin.Op.String(), y.Value}
	val := strings.Join(vals, delim)

	return &BasicLit{
		ValuePos: bin.Pos(),
		Value:    val,
		Kind:     token.STRING,
	}
}
