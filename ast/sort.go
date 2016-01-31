package ast

import (
	"fmt"
	"sort"
)

type Stmts []Stmt

func (s Stmts) lookup(pos int) int {
	i := 0
	switch s[pos].(type) {
	case *DeclStmt, *IncludeStmt, *CommStmt:
	case *SelStmt:
		i = 1
	}
	return i
}

func (s Stmts) Len() int {
	return len(s)
}

func (s Stmts) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Stmts) Less(i, j int) bool {
	return s.lookup(i) < s.lookup(j)
}

// Sort statements for most efficient usage of CSS rules
// (rules first, then other tings)
func SortStatements(list Stmts) {
	for i, stmt := range list {
		fmt.Printf("%d: % #v\n", i, stmt)
	}
	sort.Sort(list)
	for i, stmt := range list {
		fmt.Printf("%d: % #v\n", i, stmt)
	}
}
