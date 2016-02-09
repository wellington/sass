package ast

import (
	"fmt"
	"log"
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
func (stmt *SelStmt) Resolve(fset *token.FileSet) {

	s := &sel{
		parent: stmt.Parent,
		stmt:   stmt,
		prec:   token.LowestPrec + 1,
		parts:  make(map[token.Pos]*BasicLit),
	}
	fmt.Println("Selector Resolve")
	Print(fset, s.stmt.Sel)
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

	// stmt.Resolved = stmt.Sel.(*BasicLit)
	var vals []string
	for i, part := range s.parts {
		fmt.Printf("%d: % #v\n", i, part)
		vals = append(vals, part.Value)
	}
	val := strings.Join(vals, " ")
	stmt.Resolved = &BasicLit{Value: val}
	fmt.Println("Resolver Output", val)
}

type sel struct {
	stmt   *SelStmt
	parent *SelStmt
	parts  map[token.Pos]*BasicLit
	prec   int    // Resolve each precendence in order
	stack  []Expr // Nesting stack
	inject bool   // inject parent to start
}

func (s *sel) add(pos token.Pos, lit *BasicLit) {
	s.parts[pos] = lit
	// FIXME: walk through all available positions and remove
	// any higher than pos. This indicates a reduce happened
	// and something was reported prematurely
	for i := range s.parts {
		if i > pos {
			delete(s.parts, i)
		}
	}
}

var amper = "&"

func ghettoResolvedParentInject(delim string, pval string, nodes ...string) string {
	fmt.Printf(`
=ghetto=============================
     op: %q
 parent: %q
 childs: %q
====================================
`,
		delim, pval, nodes,
	)
	gdelim := ", "
	if len(pval) > 0 {
		sdelim := ", "
		parts := strings.Split(pval, sdelim)
		ret := make([]string, 0, len(parts)*len(nodes))
		for i := range parts {
			for j := range nodes {
				// if no &, prepend to start
				var s string
				if strings.Contains(nodes[j], amper) {
					// ret = append(ret, parts[i]+delim+nodes[j])

					s = strings.Replace(nodes[j], "&", parts[i], -1)
				} else {
					s = parts[i] + delim + nodes[j]
				}
				ret = append(ret, s)
			}
		}
		fmt.Printf(`
=ghetto return======================
 %q
====================================
`, ret)
		return strings.Join(ret, gdelim)
	}
	return strings.Join(nodes, gdelim)
}

// FIXME: have no way to merge trees right now, so ghetto style
func ghettoParentInject(delim string, parent *SelStmt, nodes ...string) string {
	var pval string
	if parent != nil {
		pval = parent.Resolved.Value
	}
	return ghettoResolvedParentInject(delim, pval, nodes...)
}

func (s *sel) Visit(node Node) Visitor {
	var pos token.Pos
	var add *BasicLit
	delim := " "
	defer func() {
		if add != nil && add.Kind != token.ILLEGAL && pos >= 0 {
			s.add(pos, add)
			// s.parts = append(s.parts, add)
			fmt.Printf("adding %d: % #v\n", pos, add)
		}
	}()

	switch v := node.(type) {
	case *UnaryExpr:
		// UnaryExpr come in two flavors & (backref) and + ~ > (operators).
		// In any case, it must be nested selector or it is an error.
		if s.parent == nil {
			// TODO: pass through parser's exception logic
			log.Fatal("unary operator must be a nested selector",
				node.Pos())
		}
		if v.Visited {
			return nil
		}
		if s.prec < 5 {
			panic(fmt.Errorf("invalid nest token: %s prec: %d", v.Op, s.prec))
		}
		if s.prec != 5 {
			return nil
		}

		v.Visited = true

		plit := ExprCopy(s.parent.Resolved).(*BasicLit)
		plit.Kind = token.STRING
		plit.ValuePos = -1
		pos = v.OpPos
		switch v.Op {
		case token.NEST:
			fmt.Println("unary nest add!", v)
			add = s.switchExpr(v)
		case token.GTR, token.TIL, token.ADD:
			bin := &BinaryExpr{}
			bin.OpPos = v.Pos()
			bin.Op = v.Op
			left := plit
			// Override position to be where the current Unary is
			left.ValuePos = v.Pos()
			bin.X = left
			bin.Y = v.X
			fmt.Println("unary binary add!")
			add = s.joinBinary(bin)
		default:
			log.Fatal("invalid unary operation: ", v.Op)
		}
		return nil
	case *BasicLit:
		if v.Kind == token.ILLEGAL {
			return nil
		}
		if s.prec != 2 {
			return nil
		}

		if s.inject && s.parent != nil {
			v.Value = ghettoParentInject(delim, s.parent, v.Value)
		}
		add = v
		return nil
	case *BinaryExpr:
		pos = v.Pos()
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
				return nil
				panic(fmt.Errorf("invalid Op token: %s prec: %d", v.Op, s.prec))
			}
			if s.prec != 4 {
				return s
			}
			add = s.joinBinary(v)
		case token.COMMA:
			if s.prec < 3 {
				return nil
				panic(fmt.Errorf("invalid group token: %s prec: %d", v.Op, s.prec))
			}
			if s.prec != 3 {
				return nil
			}
			fmt.Println("COMMMA!")
			// Reset parent injector
			// s.inject = true
			// Reset parent injector
			// s.inject = true
			// Walk(s, v.Y)

			litX := s.switchExpr(v.X)
			litY := s.switchExpr(v.Y)
			lits := append(
				strings.Split(litX.Value, ","+delim),
				strings.Split(litY.Value, ","+delim)...)
			sx := ghettoParentInject(delim, s.parent, lits...) //litX.Value, litY.Value)
			add = &BasicLit{
				Kind:     token.STRING,
				ValuePos: pos,
				Value:    sx,
			}
			fmt.Println("returned comma")
			return nil
		}

		// v.Op = token.ILLEGAL
		return nil
	}

	return s
}

func parseBackRef(parent *BasicLit, in *BasicLit) *BasicLit {
	if in.Value == "&" {
		return ExprCopy(parent).(*BasicLit)
	}
	delim := " "
	pval := parent.Value
	// parts := strings.Split(in.Value, " ")
	// ret := make([]string, len(parts))
	// for i, part := range parts {
	// 	fmt.Println("parseBackRef part", i, part)
	// 	fmt.Println("parseBackRef ret ", i, ghettoResolvedParentInject(delim,
	// 		pval, part))
	// 	ret[i] = ghettoResolvedParentInject(delim, pval, part)
	// }
	ret := ghettoResolvedParentInject(delim, pval, in.Value)
	fmt.Printf("parseBackRef final return %q\n", ret)
	return &BasicLit{
		Kind:     token.STRING,
		Value:    ret, //strings.Join(ret, "*"),
		ValuePos: in.Pos(),
	}
}

func (s *sel) switchExpr(expr Expr) *BasicLit {
	switch v := expr.(type) {
	case *BasicLit:
		// v.Kind = token.ILLEGAL
		return v
	case *UnaryExpr:
		// plit := ExprCopy(s.parent.Resolved).(*BasicLit)
		plit := parseBackRef(s.parent.Resolved, v.X.(*BasicLit))
		return plit
	case *BinaryExpr:
		fmt.Printf("switching bin\n  X:% #v\n  Y:% #v\n", v.X, v.Y)
		return s.joinBinary(v)
	default:
		panic(fmt.Errorf("switch expr: % #v\n", v))
	}
}

func (s *sel) joinBinary(bin *BinaryExpr) *BasicLit {
	var x, y *BasicLit
	// If either are Unary, must use ghetto math to multiply them
	_, unx := bin.X.(*UnaryExpr)
	_, uny := bin.Y.(*UnaryExpr)
	_, _ = unx, uny
	x = s.switchExpr(bin.X)
	y = s.switchExpr(bin.Y)

	delim := " " // This will change with compiler mode
	switch bin.Op {
	case token.COMMA:
		delim = "," + delim
	default:
		delim = delim + bin.Op.String() + delim
	}

	fmt.Printf("joining with (%q)\n  X: % #v\n  Y: % #v\n", delim, x, y)
	var val string
	if unx {
		val = ghettoResolvedParentInject(delim, x.Value, y.Value)
	} else if uny {
		// This won't actually work, but hey have fun kid
		val = ghettoResolvedParentInject(delim, y.Value, x.Value)
	} else if bin.Op == token.COMMA {
		// Ghetto to the max, do the right side
		// val = ghettoParentInject(","+delim, s.parent, x.Value, y.Value)
		val = x.Value + delim + y.Value
	} else {
		vals := []string{x.Value, y.Value}
		val = strings.Join(vals, delim)
	}

	lit := &BasicLit{
		ValuePos: bin.Pos(),
		Value:    val,
		Kind:     token.STRING,
	}
	fmt.Printf("binJoined: %s\n", val)
	return lit
}
