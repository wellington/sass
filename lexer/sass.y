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

%type	<x>	props prop selectors selector nested
%token  <x> LBRACKET RBRACKET COLON SEMIC TEXT

%token  <x>           STR


%%
top: /* empty */
                nested {
                    fmt.Fprint(out, $1.Value)
                }
        ;
nested:
                selectors
        |       TEXT LBRACKET nested RBRACKET {
                    $$.Value = $1.Value + " " + $3.Value
                }
        ;
selectors:
                selector
        |       selectors selector
        ;
selector:
                prop
        |       TEXT LBRACKET props RBRACKET {
                    $$.Value = $1.Value + $2.Value + $3.Value + $4.Value
                }
        ;
props:
                prop
        |       props prop
        ;
prop:
                STR
        |       TEXT COLON TEXT SEMIC {
                    $$.Value = $1.Value + $2.Value + $3.Value + $4.Value
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
