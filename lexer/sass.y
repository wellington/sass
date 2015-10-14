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
    "regexp"
    "strings"
)

%}

%union {
    s Set
    v map[string]string
    x *Item
}

%token  <s>             STMT
%type   <s>             stmt

%token  <x>             VAR
%token  <x>             SUB
%type   <x>             vars

%token  <x>             RULE
%type   <x>             selectors

%type	<x>             props prop nested
%token  <x>             LBRACKET RBRACKET COLON SEMIC TEXT
%token  <x>             ITEM


%%
top: /*         empty */
        |       stmt {
                    debugPrint("stmt", $1)
                    var sout string
                    rules := $1.Rules
                    props := $1.Props
                    debugPrint("rules:", rules)
                    debugPrint("props:", props)
                    if len(rules) != len(props) {
                        fmt.Println(rules)
                        fmt.Println(props)
                        sout = fmt.Sprintf(
                            "props/rules mismatch rules(%d) props(%d)",
                            len(rules), len(props))
                    } else {
                        for i := range rules {
                            r := strings.Join(rules[0:i+1], " ")
                            if len(props[i]) > 0 {
                                sigh := strings.Replace(props[i],
                                                        ":", ": ", -1)
                                sout += r + " { " + sigh + " }" + "\n"
                            }
                        }
                    }
                    fmt.Fprint(out, sout)
                }
                ;
stmt:           STMT
        |       props stmt {
                    // do variable substitutions here
                    debugPrint("stmt2", $1, $2)
                    vars := $1.Vars
                    props := $2.Props
                    re := regexp.MustCompile("\\$[a-zA-Z0-9]+")
                    for i := range props {
                        m := re.FindString(props[i])
                        if rep, ok := vars[m]; ok && len(m) > 0 {
                            props[i] = strings.Replace(props[i], m, rep, 1)
                        }
                    }
                    $$.Props = props
                    $$.Rules = $2.Rules
                }
        |       selectors nested {
                    rules := append($1.Rules, $2.Rules...)
                    props := append($1.Props, $2.Props...)
                    vars := make(map[string]string)
                    for k, v := range $1.Vars {
                        $1.Vars[k] = v
                    }
                    for k, v := range $2.Vars {
                        $2.Vars[k] = v
                    }

                    $$.Rules = rules
                    $$.Props = props
                    $$.Vars = vars
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
                    $$.Vars = $1.Vars
                }
        |       vars
        |       selectors vars {
                    debugPrint("")
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
        |       selectors nested {
            debugPrint("nested5", $1, $2)
                }
                ;
props:
                prop
        |       props prop {
                    debugPrint("props2:", $1, $2)
                    $$.Props = []string{$1.Props[0] + $2.Props[0]}
                }
                ;
vars:
                ITEM
        ;
prop:
                ITEM
        |       TEXT COLON SUB SEMIC { // variable replacement
                    fmt.Println("prop2", $1, $2, $3, $4)
                    s := []string{$1.Value+$2.Value+$3.Value+$4.Value}
                    $$.Props = s
                }
        |       VAR COLON TEXT SEMIC { // variable assignment
                    fmt.Println("var3", $1, $2, $3, $4)
                    if $$.Vars == nil {
                        $$.Vars = make(map[string]string)
                    }
                    $$.Vars[$1.Value] = $3.Value
                }
        |       TEXT COLON TEXT SEMIC {
                    debugPrint("prop3:", $1, $2, $3, $4)
                    $$.Props = []string{$1.Value + $2.Value +
                    $3.Value + $4.Value}
                    $$.Value = ""
                }
                ;
%%

type Set struct {
    Rules []string
    Props []string
    Vars map[string]string
}

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
