package ast

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/wellington/sass/token"
)

var (
	regEql = regexp.MustCompile("\\s*(\\*?=)\\s*").ReplaceAll
	regBkt = regexp.MustCompile("\\s*(\\[)\\s*(\\S+)\\s*(\\])").ReplaceAll
)

func NewSelStmt(lit string, pos token.Pos) *SelStmt {
	sel := &SelStmt{
		Name: &Ident{
			NamePos: pos,
			Name:    lit,
		},
	}
	// Parse Ident looking goodies
	sel.init()
	return sel
}

// Init preps a SelStmt by rendering Individual selectors ',' delimited
// and expanding again with CSS operators '+', '>'
func (s *SelStmt) init() {
	// Break selectors on commas
	s.Names = selExpandComma(s.Name.Name, s.Pos())
	s.lexemes = selExpand(s.Name.Name, s.Pos(), "")
}

// Applies unique spacing rules to selectors
// a  +  b => a + b
// [hey = 'ho'] => [hey='ho']
func trimSelSpace(lit []byte) ([]byte, int) {
	// Regexps are the worst way to do this

	var offset int
	l := len(lit)
	lit = bytes.TrimLeft(lit, " ")
	offset = l - len(lit)
	// Remove extra spaces
	lit = bytes.TrimRight(lit, " ")
	lit = regEql(lit, []byte("$1"))
	lit = regBkt(lit, []byte("$1$2$3"))
	// Remove spaces around '=' blocks

	return lit, offset
}

const seldelims string = "&>+,"

func IsSelDelim(r rune) bool {
	return strings.ContainsRune(seldelims, r)
}

func selExpand(lit string, start token.Pos, runes string) []*BasicLit {
	var cur, pos int
	var parts []string
	var lits []*BasicLit
	_, _ = parts, lits
	sublit := []byte(lit)

	cur = bytes.IndexAny(sublit, seldelims)
	for cur != -1 {
		s := sublit[:cur]

		val, offset := trimSelSpace([]byte(s))
		lits = append(lits, &BasicLit{
			Value:    string(val),
			ValuePos: token.Pos(pos + offset),
			Kind:     token.STRING,
		})
		delim := sublit[cur : cur+1]
		r := rune(delim[0])
		var tok token.Token
		switch r {
		case '+':
			tok = token.ADD
		case '&':
			tok = token.AND
		case '>':
			tok = token.GTR
		case ',':
			tok = token.COMMA
		}

		lits = append(lits, &BasicLit{
			Value:    string(delim),
			ValuePos: token.Pos(pos + cur),
			Kind:     tok,
		})
		sublit = sublit[cur+1:]
		pos = pos + cur + 1
		cur = bytes.IndexAny(sublit, seldelims)
	}
	return lits
}

// expand separates comma separated CSS rules into []Names
func selExpandComma(lit string, pos token.Pos) []*Ident {
	// lit := s.Name.Name
	// pos := s.NamePos
	lits := strings.SplitAfter(lit, ",")
	idents := make([]*Ident, len(lits))
	var l int

	// Expand rules on commas
	for i, olit := range lits {
		lit := []byte(olit)
		if lr, _ := utf8.DecodeLastRune(lit); lr == ',' {
			lit = lit[:len(lit)-1]
		}
		lit, offset := trimSelSpace(lit)
		idents[i] = &Ident{
			// TODO: NamePos will point to whitespace following ,
			NamePos: pos + token.Pos(l+offset),
			Name:    string(lit),
		}
		l = l + len(olit)
	}

	return idents

}

func resolveParent(node, parent *SelStmt, sep string) (repeat bool) {
	parNames := parent.Names
	names := node.Names
	ret := make([]*Ident, len(parNames)*len(names))
	var pos int
	for i := range parNames {
		pos = i * len(names)
		for j := range names {
			ident := IdentCopy(names[j])
			parName := parNames[i].Name
			nodeName := names[j].Name
			count := strings.Count(nodeName, "&")
			switch count {
			case 0:
				sels := []string{parName, nodeName}
				ident.Name = strings.Join(sels, sep)
			case 1:
				// Simple substitution
				ident.Name = strings.Replace(nodeName, "&", parName, 1)
			}
			ret[pos+j] = ident
		}
	}
	node.Names = ret
	return
}

func identsToSlice(idents []*Ident) []string {
	ss := make([]string, len(idents))
	for i := range idents {
		ss[i] = idents[i].Name
	}
	return ss
}

// Collapse takes the tree of nested selectors and generates
// CSS base rules out of them. Collapse must be a non-destructive
// process, it is re-run inside every mixin include.
//
// SelStmt.Name will always have the original selector
func (s *SelStmt) Collapse(parents []*SelStmt, backRefOk bool, errFn func(token.Pos, string)) {
	// Expands selectors into slice separated by ','

	if len(parents) == 0 {
		if strings.Contains(s.Name.Name, "&") {
			errFn(s.NamePos, "Back references (&) are not allowed in base-rule")
		}
		s.Names = selExpandComma(s.Name.Name, s.NamePos)
		return
	}

	// Rest requires math
	lastParent := parents[len(parents)-1]

	s.Names = []*Ident{NewIdent(selMultiply(" ",
		strings.Join(identsToSlice(lastParent.Names), ", "),
		s.Name.Name,
	))}
	fmt.Println("we found", s.Names)

}

// selMultiply takes two selectors and multiplies them
// a { b {} }       ~> a b
// a, b { c {} }    ~> a c, b c
// a, b { c, d {} } ~> a b, a d, b c, bd
// Some support for backreferences
func selMultiply(sep string, sels ...string) string {
	if len(sels) == 0 {
		return ""
	}
	if len(sels) == 1 {
		return sels[0]
	}

	a, b := sels[0], sels[1]
	if len(sels) > 2 {
		b = selMultiply(" ", sels[1:]...)
	}

	if b == "&" {
		return a
	}

	aa := selStringExpandComma(a)
	bb := selStringExpandComma(b)
	laa := len(aa)
	lbb := len(bb)
	ct := strings.Count(b, "&")
	// Simple replacement of parents (no expansion)
	if len(aa) == 1 && ct > 0 {
		return strings.Replace(b, "&", a, ct)
	}
	// Still simple replacement
	if len(aa) > 1 && ct == 1 {
		return strings.Replace(b, "&", a, ct)
	}

	// Multiple ampersand replacement with multiple parents
	// requires detailed token analysis of the selector for
	// determining the output
	// ie. a, b { & + & } ~> a + a, a + b, b + a, b + b
	if len(aa) > 1 && ct > 1 {
		fmt.Println("a:", a, "b:", b)
		panic("unsupported back reference resolution")
	}

	ret := make([]string, laa*lbb)
	var pos int
	for i := range aa {
		pos = i * lbb
		for j := range bb {
			ret[pos+j] = strings.Join([]string{aa[i], bb[j]}, " ")
		}
	}
	s := strings.Join(ret, ", ")
	return s
}

func selStringExpandComma(s string) []string {
	// FIXME: backreference work is two steps back 0 steps forward
	idents := selExpandComma(s, 0)
	return identsToSlice(idents)

	sels := strings.Split(s, ",")
	for i := range sels {
		sels[i] = strings.TrimSpace(sels[i])
	}
	return sels
}
