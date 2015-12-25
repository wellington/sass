package parser

import (
	"bytes"
	"errors"
	"fmt"
	goast "go/ast"
	"go/scanner"
	gotoken "go/token"
	"io"
	"io/ioutil"
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
	errors  scanner.ErrorList
	scanner scanner.Scanner
	trace   bool

	pos gotoken.Pos
	tok gotoken.Token
	lit string
}

func ParseFile(fset *gotoken.FileSet, filename string, src interface{}) (f *goast.File, err error) {

	text, err := readSource(filename, src)
	if err != nil {
		return nil, err
	}

	var p parser
	p.init(fset, filename, text)
	f = p.parseFile()
	return
}

func (p *parser) parseFile() *goast.File {

	if p.trace {
		defer un(trace(p, "File"))
	}

	return &goast.File{}
}

func (p *parser) init(fset *gotoken.FileSet, filename string, text []byte) {

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
