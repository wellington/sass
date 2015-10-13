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
    s string
    x *Item
    r *Item
}

%token  <s>             STMT
%type   <s>             stmt

%token  <r>             RULE
%type   <r>             selectors

%type	<x>             props prop nested
%token  <x>             LBRACKET RBRACKET COLON SEMIC TEXT
%token  <x>             ITEM


%%
top: /*         empty */
        |       stmt {
            fmt.Fprint(out, $1)
                }
                ;
stmt:           STMT
                | selectors nested {
    if debug{fmt.Println("stmt", $1, $2)}
                    $$ = $1.Value + $2.Value
                }
                ;
selectors:
                RULE
        |       selectors RULE {
            if debug {fmt.Println("sel2", $1, $2)}
                    $$.Value = $1.Value + " " + $2.Value
                        }
                ;
nested:
                props
        |       LBRACKET selectors nested RBRACKET {
                    if debug{fmt.Println("nested sel:", $1, $2, $3)}
                    $$.Value = " " + $2.Value + $3.Value
                }
        |       LBRACKET nested RBRACKET {
            if debug {fmt.Println("nested:", $1, $2, $3)}
                $$.Value = $1.Value + $2.Value + $3.Value
                }
                ;
props:
                prop
        |       props prop {
                    if debug {fmt.Println("props", $1, $2)}
                    $$.Value = $1.Value + $2.Value
                        }
                ;
prop:
                ITEM
        |       TEXT COLON TEXT SEMIC {
                    if debug {
                            fmt.Printf("prop: %s %s %s %s\n",
                                       $1.Value, $2.Value,
                                       $3.Value, $4.Value)
                    }
                    $$.Value = $1.Value + $2.Value + $3.Value + $4.Value
                }
                ;
%%

var out io.Writer
var debug bool

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
