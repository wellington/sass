package ast

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/wellington/sass/token"
)

var (
	regEql = regexp.MustCompile("\\s*(\\*?=)\\s*").ReplaceAll
	regBkt = regexp.MustCompile("\\s*(\\[)\\s*(\\S+)\\s*(\\])").ReplaceAll
	nilW   = bytes.NewBuffer(nil)
)

// Resolves walks selector operations removing nested Op by prepending X
// on Y.
func (stmt *SelStmt) Resolve(fset *token.FileSet) {
	if stmt.Sel == nil {
		panic(fmt.Errorf("invalid selector: % #v\n", stmt))
	}
	delim := " "
	var par string
	if stmt.Parent != nil {
		par = stmt.Parent.Resolved.Value
	}
	val := resolve3(par, stmt.Name.Name)
	// log.SetOutput(os.Stderr)
	stmt.Resolved = Selector(stmt)
	stmt.Resolved = &BasicLit{
		Kind:     token.STRING,
		Value:    strings.Join(val, ","+delim),
		ValuePos: stmt.Pos(),
	}
	return
}

var r = regexp.MustCompile("\\s{2,}")
var d = regexp.MustCompile("\\s*([+~>])\\s*")

// Third time is a charm
func resolve3(par, raw string) []string {
	delim := " "
	fmt.Println("raw", raw)
	// Replace consecutive whitespace with a single whitespace
	clean := d.ReplaceAllString(raw, " $1 ")
	fmt.Println("1", clean)
	//clean = d.ReplaceAllString(clean, " $1 ")
	fmt.Println("2", clean)
	nodes := selSplit(clean)
	fmt.Println("clean", clean)
	log.Printf("Resolve3 Sel        %q\n", nodes)
	log.Printf("Parent 3            %q\n", par)
	merged := joinParent(delim, par, nodes)
	log.Printf("Adopted3            %q\n", merged)
	return merged
}

func selSplit(s string) []string {
	ss := strings.Split(s, ",")
	for i := range ss {
		ss[i] = strings.TrimSpace(ss[i])
		if !strings.Contains(ss[i], "&") {
			// Add implicit ampersand
			ss[i] = "& " + ss[i]
		}
	}
	return ss
}

func joinParent(delim, parent string, nodes []string) []string {
	rep := "&"
	if len(parent) == 0 {
		rep = "& "
	}
	commadelim := "," + delim
	parts := strings.Split(parent, commadelim)
	var ret []string
	for i := range parts {
		for j := range nodes {
			rep := strings.Replace(nodes[j], rep, parts[i], -1)
			ret = append(ret, rep)
		}
	}
	return ret
}
