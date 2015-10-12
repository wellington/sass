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

%type	<x>	expr expr1
%token  <x> LBRACKET RBRACKET COLON SEMIC TEXT

%token  <x>           STR


%%
top:
                expr {
                    fmt.Fprint(out, $1.Value)
                }
        ;
expr:
                expr1
        |       expr LBRACKET { $$.Value += $2.Value }
        |       expr RBRACKET { $$.Value += $2.Value }
        |       SEMIC { $$.Value = $1.Value }
        |       TEXT { $$.Value = $1.Value }
        |       COLON { }
        ;
expr1:
                STR
        |       expr TEXT { $$.Value = $1.Value+$2.Value }
        |       expr COLON { $$.Value = $1.Value+$2.Value }
        |       expr SEMIC { $$.Value = $1.Value+$2.Value }
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
