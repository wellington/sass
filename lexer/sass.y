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
    "strings"
)

%}

%union {
    s string
    x *Item
}

%token  <s>             STMT
%type   <s>             stmt

%token  <x>             RULE
%type   <x>             selectors

%type	<x>             props prop nested
%token  <x>             LBRACKET RBRACKET COLON SEMIC TEXT
%token  <x>             ITEM


%%
top: /*         empty */
        |       stmt {
                    fmt.Fprint(out, $1)
                }
                ;
stmt:           STMT { debugPrint("stmt1", $1) }
                | selectors nested {
                    debugPrint("stmt2", $1, $2)
                    rules := append($1.Rules, $2.Rules...)
                    props := append($1.Props, $2.Props...)
                    debugPrint("rules:", rules)
                    debugPrint("props:", props)
                    if len(rules) != len(props) {
                            fmt.Println(rules)
                            fmt.Println(props)
                        $$ = fmt.Sprintf(
                            "props/rules mismatch rules(%d) props(%d)",
                            len(rules), len(props))
                    } else {
                        var sout string
                        for i := range rules {
                            r := strings.Join(rules[0:i+1], " ")
                            if len(props[i]) > 0 {
                                sout += r + " {" + props[i] + "}"
                            }
                        }
                        $$ = sout
                    }
                }
                ;
selectors:
                RULE {
                    debugPrint("sel1:", $1)
                    $$.Rules = []string{$1.Value}
                    $$.Value = ""
                }
        |       selectors RULE {
                    debugPrint("sel2:", $1, $2)
                    $$.Rules = append($1.Rules, $2.Rules...)
                    $$.Value = ""
                }
                ;
nested:
                props {
                    debugPrint("nested1:", $1)
                    $$.Rules = $1.Rules
                    $$.Value = ""
                }
        |       LBRACKET selectors nested RBRACKET {
                    debugPrint("nested2:", $1, $2, $3, $4)
                    $$.Rules = append($2.Rules, $3.Rules...)
                    $$.Props = append([]string{""}, $3.Props...)
                }
        |       LBRACKET nested selectors nested RBRACKET {
                debugPrint("nested3:", $1, $2, $3, $4, $5)
                $$.Rules = $3.Rules
                $$.Props = append($2.Props, $4.Props...)
                $$.Value = ""
                // $$.Value = $1.Value + $2.Value + $4.Value + $5.Value
                }
        |       LBRACKET nested RBRACKET {
                debugPrint("nested4:", $1, $2, $3)
                $$.Rules = $2.Rules
                $$.Props = $2.Props
                $$.Value = ""
                }
                ;
props:
                prop
        |       props prop {
                    debugPrint("props2:", $1, $2)
                    $$.Props = []string{$1.Props[0] + $2.Props[0]}
                }
                ;
prop:
                ITEM
        |       TEXT COLON TEXT SEMIC {
                debugPrint("prop2:", $1, $2, $3, $4)
                $$.Props = []string{$1.Value + $2.Value +
                $3.Value + $4.Value}
                $$.Value = ""
                }
                ;
%%

func debugPrint(name string, vs ...interface{}) {
    if !debug { return }
    app := fmt.Sprint(vs)
    fmt.Println(name, app)
}

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
