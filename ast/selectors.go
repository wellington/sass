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
	Print(token.NewFileSet(), s.stmt.Sel)
	// This could be more efficient, it should inspect precision of
	// the top node
	for prec := token.UnaryPrec; prec > 1; prec-- {
		// Walk the selectors resolving ops found at the active
		// precision
		if s.parent != nil {
			s.inject = true
		}
		s.prec = prec
		fmt.Printf("Let's walk (%p)\n", s.stmt.Sel)
		Walk(s, s.stmt.Sel)
	}

	// stmt.Resolved = stmt.Sel.(*BasicLit)
	Print(token.NewFileSet(), s.stmt.Sel)
	fmt.Println("parts len", len(s.parts))
	var vals []string
	for i, part := range s.parts {
		fmt.Printf("%d: % #v\n", i, part)
		vals = append(vals, part.Value)
	}
	val := strings.Join(vals, " ")
	stmt.Resolved = &BasicLit{Value: val}

}

type sel struct {
	stmt   *SelStmt
	parent *SelStmt
	parts  []*BasicLit
	prec   int    // Resolve each precendence in order
	stack  []Expr // Nesting stack
	inject bool   // inject parent to start
}

func (s *sel) Visit(node Node) Visitor {
	var add *BasicLit
	defer func() {
		if add != nil {
			s.parts = append(s.parts, add)
		}
	}()
	fmt.Printf("%d: (%p) % #v\n", s.prec, node, node)
	switch v := node.(type) {
	case *UnaryExpr:
		// Nesting, collapse &
		if v.Op == token.ILLEGAL {
			return nil
		}
		if s.prec < 5 {
			panic(fmt.Errorf("invalid nest token: %s prec: %d", v.Op, s.prec))
		}
		if s.prec != 5 {
			return s
		}
		s.inject = false
		v.Op = token.ILLEGAL
		x := v.X.(*BasicLit)
		add = x
		return nil
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
		add = v
		return nil
	case *BinaryExpr:
		switch v.Op {
		case token.NEST:
			if s.prec < 5 {
				panic(fmt.Errorf("invalid binary nest token: %s prec: %d", v.Op, s.prec))
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
			add = s.joinBinary(v)
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
			add = s.joinBinary(v)
		}
		v.Op = token.ILLEGAL
		return nil
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

	// Mark Op as illegal to indicate resolved
	return &BasicLit{
		ValuePos: bin.Pos(),
		Value:    val,
		Kind:     token.STRING,
	}
}
