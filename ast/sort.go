package ast

import "fmt"

// Sort statements for most efficient usage of CSS rules
// (rules first, then other tings)
func StatementsSort(list []Stmt) []Stmt {
	rules := make([]Stmt, 0, len(list))
	notrules := make([]Stmt, 0, len(list))
	for _, stmt := range list {
		switch v := stmt.(type) {
		case *DeclStmt:
			// Rule
			rules = append(rules, stmt)
		case *IncludeStmt, *CommStmt, *AssignStmt:
			rules = append(rules, stmt)
		case *SelStmt:
			notrules = append(notrules, stmt)
		default:
			fmt.Printf("unhandled sort! % #v\n", v)
			notrules = append(notrules, stmt)
		}
	}
	return append(rules, notrules...)
}
