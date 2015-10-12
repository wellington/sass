// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This is an example of a goyacc program.
// To build it:
// go tool yacc -p "expr" expr.y (produces y.go)
// go build -o expr y.go
// expr
// > <type an expression>

%{

package main

import (
	"os"
	"log"
	"bufio"
	"io"
    "fmt"
)

%}

%union {
    x *Item
    val string
    itype ItemType
}

%type	<x>	selector property nested
%token  <x> LBRACKET RBRACKET COLON SEMIC TEXT

%token  <x>           STR


%%
top:
                property {
                    fmt.Fprint(out, $1.Value)
                }
        ;
property:
                selector
        |       TEXT COLON TEXT SEMIC { $$.Value = $1.Value + $2.Value + $3.Value + $4.Value }
        ;
selector:
                nested
        |       TEXT LBRACKET property RBRACKET {
                    $$.Value = $1.Value + $2.Value + $3.Value + $4.Value
                        }
        ;
nested:
                STR
        |       TEXT LBRACKET selector RBRACKET {
                    $$.Value = $1.Value + " " + $3.Value
                }
        ;
%%

var out io.Writer

func init() {
    out = os.Stdout
}

func main() {
        yyErrorVerbose = true
        in := bufio.NewReader(os.Stdin)
        _ = in
        sin := `hello`

     lex := New(func(l *Lexer) StateFn {
         return l.Action()
     }, sin)

     if false {
        lval := new(yySymType)
        for {
            tok := lex.Lex(lval)
            log.Printf("tok - %d\n", tok)

            if tok == 0 {
                log.Println("break")
                return
            }
        }
        return
     }

     for {
        if _, err := os.Stdout.WriteString("> "); err != nil {
            log.Fatalf("WriteString: %s", err)
        }
        line, err := in.ReadBytes('\n')
        if err == io.EOF {
            return
        } else if err != nil {
            log.Fatalf("ReadBytes: %s", err)
        }

        yyParse(New(func(l *Lexer) StateFn {
            return l.Action()
        }, string(line)))
     }
}
