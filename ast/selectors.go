package ast

import (
	"fmt"
	"log"
	"math"
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
	if stmt.Sel == nil {
		panic(fmt.Errorf("invalid selector: % #v\n", stmt))
	}
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
	fmt.Printf(`=ghetto=============================
     op: %q
 parent: %q
 childs: %q
====================================
`,
		delim, pval, nodes,
	)
	gdelim := ", "

	if len(pval) == 0 {
		return strings.Join(nodes, gdelim)
	}

	sdelim := ", "
	parts := strings.Split(pval, sdelim)
	ret := make([]string, 0, len(parts)*len(nodes))
	var s string
	for i := range parts {
		for j := range nodes {
			// When no & is present, & is implicit ie. `& parts[i]`
			if strings.Contains(nodes[j], amper) {
				s = strings.Replace(nodes[j], "&", parts[i], -1)
			} else {
				s = parts[i] + delim + nodes[j]
			}
			ret = append(ret, s)
		}
	}
	fmt.Printf(`=ghetto return======================
 %q
====================================
`, ret)
	return strings.Join(ret, gdelim)
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
	// fmt.Printf("Visit %T: % #v\n", node, node)
	var pos token.Pos
	var add *BasicLit
	delim := " "
	defer func() {
		if add == nil {
			return
		}
		if add.Kind == token.ILLEGAL {
			log.Println("Warning invalid Kind for", add)
		}
		// Do not add Lits with invalid positions
		if pos >= 0 {
			s.add(pos, add)
			fmt.Printf("adding %s at %d: % #v\n", add.Kind, pos, add)
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
			add = s.switchExpr(v)
		case token.GTR, token.TIL, token.ADD:
			fmt.Println("unary binary add!")
			add = s.switchExpr(v)
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

			litX := s.switchExpr(v.X)
			litY := s.switchExpr(v.Y)
			sx := mergeLits(","+delim, litX.Value, litY.Value)
			add = &BasicLit{
				Kind:     token.STRING,
				ValuePos: pos,
				Value:    sx,
			}
		}
		return nil
	}

	return s
}

// after parent multiplication, lits are out of order. Fix the ordering
// Examples of out of orderness
// [1 3] [2] => [1 2 3]
// [1 3] [2 4] => [1 2 3 4]
func mergeLits(delim, left, right string) string {
	lefts, rights := strings.Split(left, delim), strings.Split(right, delim)
	ll, lr := len(lefts), len(rights)
	fmt.Printf("reordering %d %d\nleft: %q\nrigh: %q\n",
		ll, lr, lefts, rights)

	if math.Remainder(float64(ll), float64(lr)) > 0 {
		panic(fmt.Errorf("Incompatible lengths left:%d right:%d", ll, lr))
	}
	var ss []string
	mod := ll / lr
	for i := range lefts {
		ss = append(ss, lefts[i])
		if (i+1)%mod == 0 {
			ss = append(ss, rights[i/mod])
		}
	}
	fmt.Printf("%q\n", ss)
	r := strings.Join(ss, delim)
	fmt.Println("mergeLits returns", r)
	return r
}

func parseBackRef(delim string, parent *BasicLit, in *BasicLit) *BasicLit {
	fmt.Printf("parseBackRef % #v\n", in)
	if in.Value == "&" {
		return ExprCopy(parent).(*BasicLit)
	}
	pval := parent.Value
	ret := ghettoResolvedParentInject(delim, pval, in.Value)
	return &BasicLit{
		Kind:     token.STRING,
		Value:    ret,
		ValuePos: in.Pos(),
	}
}

func (s *sel) switchExpr(expr Expr) *BasicLit {
	fmt.Printf("switchExpr %T: % #v\n", expr, expr)
	delim := " "
	switch v := expr.(type) {
	case *BasicLit:
		copy := ExprCopy(expr).(*BasicLit)
		copy.ValuePos = v.ValuePos
		copy.Value = ghettoParentInject(" ", s.parent, v.Value)
		return copy
	case *UnaryExpr:
		plit := parseBackRef(delim+v.Op.String()+delim, s.parent.Resolved, v.X.(*BasicLit))
		fmt.Printf("switchExpr exit % #v\n", plit)
		return plit
	case *BinaryExpr:
		fmt.Printf("switching bin\n  X:% #v\n  Y:% #v\n", v.X, v.Y)
		return s.joinBinary(v)
	default:
		panic(fmt.Errorf("switch expr: % #v\n", v))
	}
}

func (s *sel) joinBinary(bin *BinaryExpr) *BasicLit {
	fmt.Println("joinBinary")
	var x, y *BasicLit
	// If either are Unary, must use ghetto math to multiply them
	_, unx := bin.X.(*UnaryExpr)
	_, uny := bin.Y.(*UnaryExpr)
	_, _ = unx, uny
	delim := " " // This will change with compiler mode
	switch bin.Op {
	case token.COMMA:
		delim = "," + delim
	default:
		delim = delim + bin.Op.String() + delim
	}

	x = s.switchExpr(bin.X)
	y = s.switchExpr(bin.Y)
	fmt.Printf("joining with (%q)\n  X: % #v\n  Y: % #v\n", delim, x, y)
	var val string
	if unx && uny {
		fmt.Println("join unx&uny\nleft:", x.Value, "\nright:", y.Value)
		val = ghettoResolvedParentInject(delim, x.Value, y.Value)
	} else if unx {
		fmt.Println("join unx")
		// This is actually a unary operation, treat as so
		un := &UnaryExpr{}
		un.Op = bin.Op
		un.OpPos = bin.OpPos
		un.X = bin.Y
		fmt.Printf("unary switch (%q): % #v", bin.Op, bin.Y)
		return s.switchExpr(un)
	} else if bin.Op == token.COMMA {
		val = mergeLits(delim, x.Value, y.Value)
	} else {
		fmt.Println("join other")
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
