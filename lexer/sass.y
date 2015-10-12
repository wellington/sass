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
)

%}

%union {
    x *Item
}

%type	<x>	expr
//                       %type   <x> line
%token  <x> LBRACKET RBRACKET COLON SEMIC TEXT

%token  <x>           STR


%%
top:
                expr
                {
                    log.Println("EXPR", yytoken)
                        log.Printf("DUMP % #v\n", $1)
                }
        ;
expr:
                LBRACKET
                {
                    log.Println("hello")
                        // $$ = $1
                }
        |       RBRACKET
                {
                    log.Println("hello")
                        // $$ = $1
                }
        |       SEMIC
                {
                    log.Println("hello")
                    $$ = $1
                }
        |       TEXT COLON TEXT
                {
                    $$.Value = $1.Value + $3.Value
                }
|       TEXT { log.Println("WOOP"); $$ = $1 }
        ;
        |       COLON { }
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

    _ = sin
    p := yyParse(lex)
    log.Printf("HI % #v\n", p)

        /*for {
        if _, err := os.Stdout.WriteString("> "); err != nil {
            log.Fatalf("WriteString: %s", err)
        }
        line, err := in.ReadBytes('\n')
        if err == io.EOF {
                log.Println("err", err, string(line))
            return
        }
        if err != nil {
            log.Fatalf("ReadBytes: %s", err)
        }

        p := yyParse(New(func(l *Lexer) StateFn {
            return l.Action()
        }, string(line)))
        log.Printf("HI % #v\n", p)
        }*/
}
