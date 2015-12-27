package parser

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"unicode"

	goast "go/ast"
	goscanner "go/scanner"
	gotoken "go/token"

	"github.com/wellington/sass/scanner"
	"github.com/wellington/sass/token"
)

const (
	basic = iota
	labelOk
	rangeOk
)

// If src != nil, readSource converts src to a []byte if possible;
// otherwise it returns an error. If src == nil, readSource returns
// the result of reading the file specified by filename.
//
func readSource(filename string, src interface{}) ([]byte, error) {
	if src != nil {
		switch s := src.(type) {
		case string:
			return []byte(s), nil
		case []byte:
			return s, nil
		case *bytes.Buffer:
			// is io.Reader, but src is already available in []byte form
			if s != nil {
				return s.Bytes(), nil
			}
		case io.Reader:
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, s); err != nil {
				return nil, err
			}
			return buf.Bytes(), nil
		}
		return nil, errors.New("invalid source")
	}
	return ioutil.ReadFile(filename)
}

type parser struct {
	file    *gotoken.File
	errors  goscanner.ErrorList
	scanner scanner.Scanner

	mode   Mode // parsing mode
	trace  bool // tracing
	indent int  // indention used for tracing

	// Comments, probably going to be unused
	comments    []*goast.CommentGroup
	leadComment *goast.CommentGroup
	lineComment *goast.CommentGroup

	pos gotoken.Pos
	tok token.Token
	lit string

	// Scopes
	pkgScope   *goast.Scope
	topScope   *goast.Scope
	unresolved []*goast.Ident
	imports    []*goast.ImportSpec
}

func ParseFile(fset *gotoken.FileSet, filename string, src interface{}, mode Mode) (f *goast.File, err error) {

	text, err := readSource(filename, src)
	if err != nil {
		return nil, err
	}

	var p parser
	p.trace = true
	p.init(fset, filename, text, mode)
	f = p.parseFile()
	return
}

func (p *parser) parseFile() *goast.File {

	if p.trace {
		defer un(trace(p, "File"))
	}

	if p.errors.Len() != 0 {
		return nil
	}

	ident := p.parseIdent()
	p.openScope()
	p.pkgScope = p.topScope
	var decls []goast.Decl
	for p.tok == token.IMPORT {
		decls = append(decls, p.parseGenDecl(token.IMPORT, p.parseImportSpec))
	}
	p.closeScope()

	i := 0
	for _, ident := range p.unresolved {
		ident.Obj = p.pkgScope.Lookup(ident.Name)

		if ident.Obj == nil {
			p.unresolved[i] = ident
			i++
		}
	}
	return &goast.File{
		Name:       ident,
		Decls:      decls,
		Scope:      p.pkgScope,
		Imports:    p.imports,
		Unresolved: p.unresolved[0:i],
		Comments:   p.comments,
	}
}

func (p *parser) expectSemi() {
	// semicolon is optional before a closing ')' or '}'
	if p.tok != token.RPAREN && p.tok != token.RBRACE {
		switch p.tok {
		case token.COMMA:
			// permit a ',' instead of a ';' but complain
			p.errorExpected(p.pos, "';'")
			fallthrough
		case token.SEMICOLON:
			p.next()
		default:
			p.errorExpected(p.pos, "';'")
			// syncStmt(p)
		}
	}
}

func (p *parser) parseImportSpec(doc *goast.CommentGroup, _ token.Token, _ int) goast.Spec {
	if p.trace {
		defer un(trace(p, "ImportSpec"))
	}

	var ident *goast.Ident
	switch p.tok {
	case token.PERIOD:
		ident = &goast.Ident{NamePos: p.pos, Name: "."}
		p.next()
	case token.IDENT:
		ident = p.parseIdent()
	}

	pos := p.pos
	var path string
	if p.tok == token.STRING {
		path = p.lit
		if !isValidImport(path) {
			p.error(pos, "invalid import path: "+path)
		}
		p.next()
	} else {
		p.expect(token.STRING) // use expect() error handling
	}
	p.expectSemi() // call before accessing p.linecomment

	// collect imports
	spec := &goast.ImportSpec{
		Doc:     doc,
		Name:    ident,
		Path:    &goast.BasicLit{ValuePos: pos, Kind: gotoken.STRING, Value: path},
		Comment: p.lineComment,
	}
	p.imports = append(p.imports, spec)

	return spec
}

func isValidImport(lit string) bool {
	const illegalChars = `!"#$%&'()*,:;<=>?[\]^{|}` + "`\uFFFD"
	s, _ := strconv.Unquote(lit) // go/scanner returns a legal string literal
	for _, r := range s {
		if !unicode.IsGraphic(r) || unicode.IsSpace(r) || strings.ContainsRune(illegalChars, r) {
			return false
		}
	}
	return s != ""
}

type parseSpecFunction func(doc *goast.CommentGroup, keyword token.Token, iota int) goast.Spec

func (p *parser) parseTmtList() (list []goast.Stmt) {
	if p.trace {
		defer un(trace(p, "StatementList"))
	}

	for p.tok != token.EOF {
		list = append(list, p.parseStmt())
	}
}

func (p *parser) parseStmt() (s goast.Stmt) {
	if p.trace {
		defer un(trace(p, "Statement"))
	}

	switch p.tok {
	// case token.CONST, token.TYPE, token.VAR:
	// 	s = &goast.DeclStmt{Decl: p.parseDecl(syncStmt)}
	case
		token.IDENT, token.INT, token.STRING, token.FUNC, token.LPAREN,
		token.LBRACK,
		// composite types
		token.ADD, token.SUB, token.MUL, token.AND, token.XOR, token.NOT:
		s, _ = p.parseSimpleStmt()
	}
}

func (p *parser) parseSimpleStmt(mode int) {
	if p.trace {
		defer un(trace(p, "SimpleStmt"))
	}

	// x := p.parseLhsList()

	switch p.tok {
	case
		token.DEFINE, token.ASSIGN:
		// assignment statement, possibly part of a range clause
		pos, tok := p.pos, p.tok
		p.next()
		var y []goast.Expr
		isRange := false
		// if mode == rangeOk && (tok == token.DEFINE || tok == token.ASSIGN) {
		// 	pos := p.pos
		// 	p.next()
		// 	y = []goast.Expr{&goast.UnaryExpr{OpPos: pos, Op: token.RANGE, X: p.parseRhs()}}
		// 	isRange = true
		// } else {
		y = p.parseRhsList()
		// }

		as := &goast.AssignStmt{Lhs: x, TokPos: pos, Tok: tok, Rhs: y}
		if tok == token.DEFINE {
			p.shortVarDecl(as, x)
		}
		return as, isRange
	}

	if len(x) > 1 {
		p.errorExpected(x[0].Pos(), "1 expression")
		// continue with first expression
	}

	switch p.tok {
	case token.COLON:
		// labeled statement
		colon := p.pos
		p.next()
		if label, isIdent := x[0].(*ast.Ident); mode == labelOk && isIdent {
			// Go spec: The scope of a label is the body of the function
			// in which it is declared and excludes the body of any nested
			// function.
			stmt := &ast.LabeledStmt{Label: label, Colon: colon, Stmt: p.parseStmt()}
			p.declare(stmt, nil, p.labelScope, ast.Lbl, label)
			return stmt, false
		}
		// The label declaration typically starts at x[0].Pos(), but the label
		// declaration may be erroneous due to a token after that position (and
		// before the ':'). If SpuriousErrors is not set, the (only) error re-
		// ported for the line is the illegal label error instead of the token
		// before the ':' that caused the problem. Thus, use the (latest) colon
		// position for error reporting.
		p.error(colon, "illegal label declaration")
		return &goast.BadStmt{From: x[0].Pos(), To: colon + 1}, false

	case token.ARROW:
		// send statement
		arrow := p.pos
		p.next()
		y := p.parseRhs()
		return &goast.SendStmt{Chan: x[0], Arrow: arrow, Value: y}, false

	case token.INC, token.DEC:
		// increment or decrement
		s := &goast.IncDecStmt{X: x[0], TokPos: p.pos, Tok: p.tok}
		p.next()
		return s, false
	}

	// expression
	return &goast.ExprStmt{X: x[0]}, false
}

func (p *parser) parseLhsList() []goast.Expr {
	old := p.inRhs
	p.inRhs = false
	list := p.parseExprList(true)
	switch p.tok {
	case token.DEFINE:
		// lhs of a short variable declaration
		// but doesn't enter scope until later:
		// caller must call p.shortVarDecl(p.makeIdentList(list))
		// at appropriate time.
	case token.COLON:
		// lhs of a label declaration or a communication clause of a select
		// statement (parseLhsList is not called when parsing the case clause
		// of a switch statement):
		// - labels are declared by the caller of parseLhsList
		// - for communication clauses, if there is a stand-alone identifier
		//   followed by a colon, we have a syntax error; there is no need
		//   to resolve the identifier in that case
	default:
		// identifiers must be declared elsewhere
		for _, x := range list {
			p.resolve(x)
		}
	}
	p.inRhs = old
	return list
}

func (p *parser) parseRhsList() []goast.Expr {
	old := p.inRhs
	p.inRhs = true
	list := p.parseExprList(false)
	p.inRhs = old
	return list
}

func syncStmt(p *parser) {

}

func (p *parser) parseGenDecl(keyword token.Token, f parseSpecFunction) *goast.GenDecl {
	if p.trace {
		defer un(trace(p, "GenDecl("+keyword.String()+")"))
	}

	doc := p.leadComment
	pos := p.expect(keyword)
	var lparen, rparen gotoken.Pos
	var list []goast.Spec
	if p.tok == token.LPAREN {
		lparen = p.pos
		p.next()
		for iota := 0; p.tok != token.RPAREN && p.tok != token.EOF; iota++ {
			list = append(list, f(p.leadComment, keyword, iota))
		}
		rparen = p.expect(token.RPAREN)
	} else {
		list = append(list, f(nil, keyword, 0))
	}

	// Hack to delay creating ast package
	gokeyword := gotoken.Token(keyword)

	return &goast.GenDecl{
		Doc:    doc,
		TokPos: pos,
		Tok:    gokeyword,
		Lparen: lparen,
		Specs:  list,
		Rparen: rparen,
	}
}

func (p *parser) expect(tok token.Token) gotoken.Pos {
	pos := p.pos
	if p.tok != tok {
		p.errorExpected(pos, "'"+tok.String()+"'")
	}
	p.next()
	return pos
}

func (p *parser) openScope() {
	p.topScope = goast.NewScope(p.topScope)
}

func (p *parser) closeScope() {
	p.topScope = p.topScope.Outer
}

func (p *parser) errorExpected(pos gotoken.Pos, msg string) {
	msg = "expected " + msg
	if pos == p.pos {
		// the error happened at the current position;
		// make the error message more specific
		if p.tok == token.SEMICOLON && p.lit == "\n" {
			msg += ", found newline"
		} else {
			msg += ", found '" + p.tok.String() + "'"
			if p.tok.IsLiteral() {
				msg += " " + p.lit
			}
		}
	}
	p.error(pos, msg)
}

// Identifiers
func (p *parser) parseIdent() *goast.Ident {
	pos := p.pos
	name := "_"
	fmt.Printf("tok %s: % #v\n", p.tok, p.tok)
	if p.tok == token.IDENT {
		name = p.lit
		p.next()
	} else {
		p.expect(token.IDENT) // use expect() error handling
	}
	return &goast.Ident{NamePos: pos, Name: name}
}

// Advance to the next non-comment token. In the process, collect
// any comment groups encountered, and remember the last lead and
// and line comments.
//
// A lead comment is a comment group that starts and ends in a
// line without any other tokens and that is followed by a non-comment
// token on the line immediately after the comment group.
//
// A line comment is a comment group that follows a non-comment
// token on the same line, and that has no tokens after it on the line
// where it ends.
//
// Lead and line comments may be considered documentation that is
// stored in the AST.
//
func (p *parser) next() {
	fmt.Println("next")
	p.leadComment = nil
	p.lineComment = nil
	prev := p.pos
	p.next0()

	if p.tok == token.COMMENT {
		var comment *goast.CommentGroup
		var endline int

		if p.file.Line(p.pos) == p.file.Line(prev) {
			// The comment is on same line as the previous token; it
			// cannot be a lead comment but may be a line comment.
			comment, endline = p.consumeCommentGroup(0)
			if p.file.Line(p.pos) != endline {
				// The next token is on a different line, thus
				// the last comment group is a line comment.
				p.lineComment = comment
			}
		}

		// consume successor comments, if any
		endline = -1
		for p.tok == token.COMMENT {
			comment, endline = p.consumeCommentGroup(1)
		}

		if endline+1 == p.file.Line(p.pos) {
			// The next token is following on the line immediately after the
			// comment group, thus the last comment group is a lead comment.
			p.leadComment = comment
		}
	}
}

// Advance to the next token.
func (p *parser) next0() {
	// Because of one-token look-ahead, print the previous token
	// when tracing as it provides a more readable output. The
	// very first token (!p.pos.IsValid()) is not initialized
	// (it is token.ILLEGAL), so don't print it .
	if p.trace && p.pos.IsValid() {
		s := p.tok.String()
		switch {
		case p.tok.IsLiteral():
			p.printTrace(s, p.lit)
		case p.tok.IsOperator(), p.tok.IsKeyword():
			p.printTrace("\"" + s + "\"")
		default:
			p.printTrace(s)
		}
	}

	p.pos, p.tok, p.lit = p.scanner.Scan()
}

// Consume a comment and return it and the line on which it ends.
func (p *parser) consumeComment() (comment *goast.Comment, endline int) {
	// /*-style comments may end on a different line than where they start.
	// Scan the comment for '\n' chars and adjust endline accordingly.
	endline = p.file.Line(p.pos)
	if p.lit[1] == '*' {
		// don't use range here - no need to decode Unicode code points
		for i := 0; i < len(p.lit); i++ {
			if p.lit[i] == '\n' {
				endline++
			}
		}
	}

	comment = &goast.Comment{Slash: p.pos, Text: p.lit}
	p.next0()

	return
}

// Consume a group of adjacent comments, add it to the parser's
// comments list, and return it together with the line at which
// the last comment in the group ends. A non-comment token or n
// empty lines terminate a comment group.
//
func (p *parser) consumeCommentGroup(n int) (comments *goast.CommentGroup, endline int) {
	var list []*goast.Comment
	endline = p.file.Line(p.pos)
	for p.tok == token.COMMENT && p.file.Line(p.pos) <= endline+n {
		var comment *goast.Comment
		comment, endline = p.consumeComment()
		list = append(list, comment)
	}

	// add comment group to the comments list
	comments = &goast.CommentGroup{List: list}
	p.comments = append(p.comments, comments)

	return
}

func (p *parser) init(fset *gotoken.FileSet, filename string, src []byte, mode Mode) {
	p.file = fset.AddFile(filename, -1, len(src))
	var m scanner.Mode
	if mode&ParseComments != 0 {
		m = scanner.ScanComments
	}
	eh := func(pos gotoken.Position, msg string) { p.errors.Add(pos, msg) }
	p.scanner.Init(p.file, src, eh, m)

	p.mode = mode
	p.trace = mode&Trace != 0

	p.next()
}

// ----------------------------------------------------------------------------
// Parsing support

func (p *parser) printTrace(a ...interface{}) {
	const dots = ". . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . "
	const n = len(dots)
	pos := p.file.Position(p.pos)
	fmt.Printf("%5d:%3d: ", pos.Line, pos.Column)
	i := 2 * p.indent
	for i > n {
		fmt.Print(dots)
		i -= n
	}
	// i <= n
	fmt.Print(dots[0:i])
	fmt.Println(a...)
}

func trace(p *parser, msg string) *parser {
	p.printTrace(msg, "(")
	p.indent++
	return p
}

// Usage pattern: defer un(trace(p, "..."))
func un(p *parser) {
	p.indent--
	p.printTrace(")")
}

// A bailout panic is raised to indicate early termination.
type bailout struct{}

func (p *parser) error(pos gotoken.Pos, msg string) {
	epos := p.file.Position(pos)

	// If AllErrors is not set, discard errors reported on the same line
	// as the last recorded error and stop parsing if there are more than
	// 10 errors.
	// if p.mode&AllErrors == 0 {
	n := len(p.errors)
	if n > 0 && p.errors[n-1].Pos.Line == epos.Line {
		return // discard - likely a spurious error
	}
	if n > 10 {
		panic(bailout{})
	}
	// }

	p.errors.Add(epos, msg)
}
