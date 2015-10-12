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
)

%}

%union {
    x *Item
}

%type	<x>	expr expr1
//                       %type   <x> line
%token  <x> LBRACKET RBRACKET COLON SEMIC TEXT

%token  <x>           STR


%%
top:
                expr { }
        ;
expr:
                expr1
        |       expr LBRACKET { $$ = $2 }
        |       expr RBRACKET { $$ = $1 }
        |       SEMIC { $$ = $1 }
        |       TEXT { $$ = $1 }
        |       COLON { }
        ;
expr1:
                STR
        |       expr TEXT { $$ = $2 }
        |       expr COLON { $$ = $2 }
        |       expr SEMIC { $$ = $2 }
        ;
%%

func main() {
    yyDebug = 10
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

        p := yyParse(New(func(l *Lexer) StateFn {
            return l.Action()
        }, string(line)))
        _ = p
     }
}
