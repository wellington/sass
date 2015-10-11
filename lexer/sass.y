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

%type	<x>	expr last

%token '+' '-' '*' '/' '(' ')'
%token LBRACKET RBRACKET COLON SEMIC

%token TEXT
%token ItemEOF

%token  <x>           STR


%%
expr:
                ctrl
        | LBRACKET expr
                {
                    log.Println("hello")
                    $$ = $2
                }
        |       RBRACKET expr
                {
                    log.Println("hello")
                        $$ = $2
                }
        |       COLON expr
                {
                    log.Println("hello")
                        $$ = $2
                }
        |       SEMIC expr
                {
                    log.Println("hello")
                        $$ = $2
                        }
        |       TEXT expr
                {
                    $$ = $2
                }
        ;
ctrl:
                last
        |       COLON expr
                {
                    $$ = $1
                }
        ;
last:
                STR
        |       '(' expr ')'
                {
                    $$ = $2
                }
        ;
%%

func main() {
    yyDebug = 10
        in := bufio.NewReader(os.Stdin)
        _ = in
        sin := `{ p { color: red; } }
    `
        p := yyParse(New(func(l *Lexer) StateFn {
                    return l.Action()
                        }, sin))
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
