//line sass.y:13
package main

import __yyfmt__ "fmt"

//line sass.y:14
import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

//line sass.y:27
type yySymType struct {
	yys int
	s   string
	x   *Item
}

const STMT = 57346
const RULE = 57347
const LBRACKET = 57348
const RBRACKET = 57349
const COLON = 57350
const SEMIC = 57351
const TEXT = 57352
const ITEM = 57353

var yyToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"STMT",
	"RULE",
	"LBRACKET",
	"RBRACKET",
	"COLON",
	"SEMIC",
	"TEXT",
	"ITEM",
}
var yyStatenames = [...]string{}

const yyEofCode = 1
const yyErrCode = 2
const yyMaxDepth = 200

//line sass.y:127

func debugPrint(name string, vs ...interface{}) {
	if !debug {
		return
	}
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

//line yacctab:1
var yyExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
}

const yyNprod = 15
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 31

var yyAct = [...]int{

	6, 4, 7, 9, 12, 11, 20, 12, 11, 23,
	15, 14, 5, 10, 19, 17, 16, 18, 24, 22,
	5, 9, 13, 21, 1, 12, 11, 3, 5, 8,
	2,
}
var yyPact = [...]int{

	23, -1000, -1000, -1000, -3, -1000, -1000, -1000, -6, 15,
	-1000, -1000, 8, -1000, -3, 7, -4, 16, -3, -1000,
	0, -1000, 11, -1000, -1000,
}
var yyPgo = [...]int{

	0, 30, 1, 29, 13, 0, 24,
}
var yyR1 = [...]int{

	0, 6, 6, 1, 1, 2, 2, 5, 5, 5,
	5, 3, 3, 4, 4,
}
var yyR2 = [...]int{

	0, 0, 1, 1, 2, 1, 2, 1, 4, 5,
	3, 1, 2, 1, 4,
}
var yyChk = [...]int{

	-1000, -6, -1, 4, -2, 5, -5, 5, -3, 6,
	-4, 11, 10, -4, -2, -5, 8, -5, -2, 7,
	10, 7, -5, 9, 7,
}
var yyDef = [...]int{

	1, -2, 2, 3, 0, 5, 4, 6, 7, 0,
	11, 13, 0, 12, 0, 0, 0, 0, 0, 10,
	0, 8, 0, 14, 9,
}
var yyTok1 = [...]int{

	1,
}
var yyTok2 = [...]int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
}
var yyTok3 = [...]int{
	0,
}

var yyErrorMessages = [...]struct {
	state int
	token int
	msg   string
}{}

//line yaccpar:1

/*	parser for yacc output	*/

var (
	yyDebug        = 0
	yyErrorVerbose = false
)

type yyLexer interface {
	Lex(lval *yySymType) int
	Error(s string)
}

type yyParser interface {
	Parse(yyLexer) int
	Lookahead() int
}

type yyParserImpl struct {
	lookahead func() int
}

func (p *yyParserImpl) Lookahead() int {
	return p.lookahead()
}

func yyNewParser() yyParser {
	p := &yyParserImpl{
		lookahead: func() int { return -1 },
	}
	return p
}

const yyFlag = -1000

func yyTokname(c int) string {
	if c >= 1 && c-1 < len(yyToknames) {
		if yyToknames[c-1] != "" {
			return yyToknames[c-1]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func yyStatname(s int) string {
	if s >= 0 && s < len(yyStatenames) {
		if yyStatenames[s] != "" {
			return yyStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func yyErrorMessage(state, lookAhead int) string {
	const TOKSTART = 4

	if !yyErrorVerbose {
		return "syntax error"
	}

	for _, e := range yyErrorMessages {
		if e.state == state && e.token == lookAhead {
			return "syntax error: " + e.msg
		}
	}

	res := "syntax error: unexpected " + yyTokname(lookAhead)

	// To match Bison, suggest at most four expected tokens.
	expected := make([]int, 0, 4)

	// Look for shiftable tokens.
	base := yyPact[state]
	for tok := TOKSTART; tok-1 < len(yyToknames); tok++ {
		if n := base + tok; n >= 0 && n < yyLast && yyChk[yyAct[n]] == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if yyDef[state] == -2 {
		i := 0
		for yyExca[i] != -1 || yyExca[i+1] != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; yyExca[i] >= 0; i += 2 {
			tok := yyExca[i]
			if tok < TOKSTART || yyExca[i+1] == 0 {
				continue
			}
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}

		// If the default action is to accept or reduce, give up.
		if yyExca[i+1] != 0 {
			return res
		}
	}

	for i, tok := range expected {
		if i == 0 {
			res += ", expecting "
		} else {
			res += " or "
		}
		res += yyTokname(tok)
	}
	return res
}

func yylex1(lex yyLexer, lval *yySymType) (char, token int) {
	token = 0
	char = lex.Lex(lval)
	if char <= 0 {
		token = yyTok1[0]
		goto out
	}
	if char < len(yyTok1) {
		token = yyTok1[char]
		goto out
	}
	if char >= yyPrivate {
		if char < yyPrivate+len(yyTok2) {
			token = yyTok2[char-yyPrivate]
			goto out
		}
	}
	for i := 0; i < len(yyTok3); i += 2 {
		token = yyTok3[i+0]
		if token == char {
			token = yyTok3[i+1]
			goto out
		}
	}

out:
	if token == 0 {
		token = yyTok2[1] /* unknown char */
	}
	if yyDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", yyTokname(token), uint(char))
	}
	return char, token
}

func yyParse(yylex yyLexer) int {
	return yyNewParser().Parse(yylex)
}

func (yyrcvr *yyParserImpl) Parse(yylex yyLexer) int {
	var yyn int
	var yylval yySymType
	var yyVAL yySymType
	var yyDollar []yySymType
	_ = yyDollar // silence set and not used
	yyS := make([]yySymType, yyMaxDepth)

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	yystate := 0
	yychar := -1
	yytoken := -1 // yychar translated into internal numbering
	yyrcvr.lookahead = func() int { return yychar }
	defer func() {
		// Make sure we report no lookahead when not parsing.
		yystate = -1
		yychar = -1
		yytoken = -1
	}()
	yyp := -1
	goto yystack

ret0:
	return 0

ret1:
	return 1

yystack:
	/* put a state and value onto the stack */
	if yyDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", yyTokname(yytoken), yyStatname(yystate))
	}

	yyp++
	if yyp >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyS[yyp] = yyVAL
	yyS[yyp].yys = yystate

yynewstate:
	yyn = yyPact[yystate]
	if yyn <= yyFlag {
		goto yydefault /* simple state */
	}
	if yychar < 0 {
		yychar, yytoken = yylex1(yylex, &yylval)
	}
	yyn += yytoken
	if yyn < 0 || yyn >= yyLast {
		goto yydefault
	}
	yyn = yyAct[yyn]
	if yyChk[yyn] == yytoken { /* valid shift */
		yychar = -1
		yytoken = -1
		yyVAL = yylval
		yystate = yyn
		if Errflag > 0 {
			Errflag--
		}
		goto yystack
	}

yydefault:
	/* default state action */
	yyn = yyDef[yystate]
	if yyn == -2 {
		if yychar < 0 {
			yychar, yytoken = yylex1(yylex, &yylval)
		}

		/* look through exception table */
		xi := 0
		for {
			if yyExca[xi+0] == -1 && yyExca[xi+1] == yystate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			yyn = yyExca[xi+0]
			if yyn < 0 || yyn == yytoken {
				break
			}
		}
		yyn = yyExca[xi+1]
		if yyn < 0 {
			goto ret0
		}
	}
	if yyn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			yylex.Error(yyErrorMessage(yystate, yytoken))
			Nerrs++
			if yyDebug >= 1 {
				__yyfmt__.Printf("%s", yyStatname(yystate))
				__yyfmt__.Printf(" saw %s\n", yyTokname(yytoken))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for yyp >= 0 {
				yyn = yyPact[yyS[yyp].yys] + yyErrCode
				if yyn >= 0 && yyn < yyLast {
					yystate = yyAct[yyn] /* simulate a shift of "error" */
					if yyChk[yystate] == yyErrCode {
						goto yystack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if yyDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", yyS[yyp].yys)
				}
				yyp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if yyDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", yyTokname(yytoken))
			}
			if yytoken == yyEofCode {
				goto ret1
			}
			yychar = -1
			yytoken = -1
			goto yynewstate /* try again in the same state */
		}
	}

	/* reduction by production yyn */
	if yyDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", yyn, yyStatname(yystate))
	}

	yynt := yyn
	yypt := yyp
	_ = yypt // guard against "declared and not used"

	yyp -= yyR2[yyn]
	// yyp is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if yyp+1 >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyVAL = yyS[yyp+1]

	/* consult goto table to find next state */
	yyn = yyR1[yyn]
	yyg := yyPgo[yyn]
	yyj := yyg + yyS[yyp].yys + 1

	if yyj >= yyLast {
		yystate = yyAct[yyg]
	} else {
		yystate = yyAct[yyj]
		if yyChk[yystate] != -yyn {
			yystate = yyAct[yyg]
		}
	}
	// dummy call; replaced with literal code
	switch yynt {

	case 2:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sass.y:45
		{
			fmt.Fprint(out, yyDollar[1].s)
		}
	case 3:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sass.y:49
		{
			debugPrint("stmt1", yyDollar[1].s)
		}
	case 4:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sass.y:50
		{
			debugPrint("stmt2", yyDollar[1].x, yyDollar[2].x)
			rules := append(yyDollar[1].x.Rules, yyDollar[2].x.Rules...)
			props := append(yyDollar[1].x.Props, yyDollar[2].x.Props...)
			debugPrint("rules:", rules)
			debugPrint("props:", props)
			if len(rules) != len(props) {
				fmt.Println(rules)
				fmt.Println(props)
				yyVAL.s = fmt.Sprintf(
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
				yyVAL.s = sout
			}
		}
	case 5:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sass.y:75
		{
			debugPrint("sel1:", yyDollar[1].x)
			yyVAL.x.Rules = []string{yyDollar[1].x.Value}
			yyVAL.x.Value = ""
		}
	case 6:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sass.y:80
		{
			debugPrint("sel2:", yyDollar[1].x, yyDollar[2].x)
			yyVAL.x.Rules = append(yyDollar[1].x.Rules, yyDollar[2].x.Rules...)
			yyVAL.x.Value = ""
		}
	case 7:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line sass.y:87
		{
			debugPrint("nested1:", yyDollar[1].x)
			yyVAL.x.Rules = yyDollar[1].x.Rules
			yyVAL.x.Value = ""
		}
	case 8:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sass.y:92
		{
			debugPrint("nested2:", yyDollar[1].x, yyDollar[2].x, yyDollar[3].x, yyDollar[4].x)
			yyVAL.x.Rules = append(yyDollar[2].x.Rules, yyDollar[3].x.Rules...)
			yyVAL.x.Props = append([]string{""}, yyDollar[3].x.Props...)
		}
	case 9:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line sass.y:97
		{
			debugPrint("nested3:", yyDollar[1].x, yyDollar[2].x, yyDollar[3].x, yyDollar[4].x, yyDollar[5].x)
			yyVAL.x.Rules = yyDollar[3].x.Rules
			yyVAL.x.Props = append(yyDollar[2].x.Props, yyDollar[4].x.Props...)
			yyVAL.x.Value = ""
			// $$.Value = $1.Value + $2.Value + $4.Value + $5.Value
		}
	case 10:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sass.y:104
		{
			debugPrint("nested4:", yyDollar[1].x, yyDollar[2].x, yyDollar[3].x)
			yyVAL.x.Rules = yyDollar[2].x.Rules
			yyVAL.x.Props = yyDollar[2].x.Props
			yyVAL.x.Value = ""
		}
	case 12:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sass.y:113
		{
			debugPrint("props2:", yyDollar[1].x, yyDollar[2].x)
			yyVAL.x.Props = []string{yyDollar[1].x.Props[0] + yyDollar[2].x.Props[0]}
		}
	case 14:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sass.y:120
		{
			debugPrint("prop2:", yyDollar[1].x, yyDollar[2].x, yyDollar[3].x, yyDollar[4].x)
			yyVAL.x.Props = []string{yyDollar[1].x.Value + yyDollar[2].x.Value +
				yyDollar[3].x.Value + yyDollar[4].x.Value}
			yyVAL.x.Value = ""
		}
	}
	goto yystack /* stack new state and value */
}
