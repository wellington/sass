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
)

//line sass.y:26
type yySymType struct {
	yys int
	s   string
	x   *Item
	r   *Item
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

//line sass.y:79

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

//line yacctab:1
var yyExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
}

const yyNprod = 13
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 23

var yyAct = [...]int{

	7, 9, 17, 9, 6, 12, 11, 12, 11, 12,
	11, 18, 10, 15, 14, 16, 3, 5, 1, 8,
	4, 13, 2,
}
var yyPact = [...]int{

	12, -1000, -1000, -1000, -5, -1000, -1000, -1000, -1, -3,
	-1000, -1000, 5, -1000, 8, -8, -1000, 2, -1000,
}
var yyPgo = [...]int{

	0, 22, 20, 19, 12, 4, 18,
}
var yyR1 = [...]int{

	0, 6, 6, 1, 1, 2, 2, 5, 5, 3,
	3, 4, 4,
}
var yyR2 = [...]int{

	0, 0, 1, 1, 2, 1, 2, 1, 3, 1,
	2, 1, 4,
}
var yyChk = [...]int{

	-1000, -6, -1, 4, -2, 5, -5, 5, -3, 6,
	-4, 11, 10, -4, -5, 8, 7, 10, 9,
}
var yyDef = [...]int{

	1, -2, 2, 3, 0, 5, 4, 6, 7, 0,
	9, 11, 0, 10, 0, 0, 8, 0, 12,
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
	case 4:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sass.y:50
		{
			yyVAL.s = yyDollar[1].r.Value + yyDollar[2].x.Value
		}
	case 6:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line sass.y:56
		{
			yyVAL.r.Value = yyDollar[1].r.Value + " " + yyDollar[2].r.Value
		}
	case 8:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line sass.y:62
		{
			fmt.Println("nested:", yyDollar[1].x, yyDollar[2].x, yyDollar[3].x)
			yyVAL.x.Value = yyDollar[1].x.Value + yyDollar[2].x.Value + yyDollar[3].x.Value
		}
	case 12:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line sass.y:73
		{
			fmt.Printf("prop: %s %s %s %s\n",
				yyDollar[1].x.Value, yyDollar[2].x.Value, yyDollar[3].x.Value, yyDollar[4].x.Value)
			yyVAL.x.Value = yyDollar[1].x.Value + yyDollar[2].x.Value + yyDollar[3].x.Value + yyDollar[4].x.Value
		}
	}
	goto yystack /* stack new state and value */
}
