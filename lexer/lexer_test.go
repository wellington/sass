package main

import (
	"fmt"
	"testing"
)

func printItems(items []Item) {
	for i, item := range items {
		fmt.Printf("%4d: %s %s\n", i, item.Type, item.Value)
	}
}

func TestLexerBools(t *testing.T) {
	if IsEOF('%', 0) != true {
		t.Errorf("Did not detect EOF")
	}
}

func TestLexerRule(t *testing.T) {
	in := `div  { color: blue; }`
	items, err := testParse(in)
	if err != nil {
		t.Fatal(err)
	}

	if e := RULE; e != int(items[0].Type) {
		t.Errorf("got: %s wanted: %s", items[0].Type, e)
	}

	if e := LBRACKET; e != int(items[1].Type) {
		t.Errorf("got: %s wanted: %s", items[1].Type, e)
	}

	if e := TEXT; e != int(items[2].Type) {
		t.Errorf("got: %s wanted: %s", items[2].Type, e)
	}

	if e := COLON; e != int(items[3].Type) {
		t.Errorf("got: %s wanted: %s", items[3].Type, e)
	}

	if e := TEXT; e != int(items[4].Type) {
		t.Errorf("got: %s wanted: %s", items[4].Type, e)
	}

	if e := SEMIC; e != int(items[5].Type) {
		t.Errorf("got: %s wanted: %s", items[5].Type, e)
	}

	if e := 7; e != len(items) {
		t.Fatal("wrong number of lexems returned")
	}

	s := items[0].Value + items[1].Value + items[2].Value + items[3].Value +
		items[4].Value + items[5].Value + items[6].Value
	if e := `div{color:blue;}`; e != s {
		t.Errorf("got: %s wanted: %s", s, e)
	}
}

func TestLexerComment(t *testing.T) {
	in := `/* some;
multiline comments +*-0
with symbols in them*/
//*Just a specially crafted single line comment
div {}
/* Invalid multiline comment`
	items, err := testParse(in)
	if err != nil {
		t.Fatal(err)
	}

	if e := `/* some;
multiline comments +*-0
with symbols in them*/`; items[0].Value != e {
		t.Errorf("Multiline comment mismatch expected:%s\nwas:%s",
			e, items[0].Value)

	}
	if e := CMT; e != items[0].Type {
		t.Errorf("Multiline CMT mismatch expected:%s, was:%s",
			e, items[0].Type)
	}
	if e := CMT; e != items[1].Type {
		t.Errorf("CMT with special chars mismatch expected:%s, was:%s",
			e, items[1].Type)
	}

	if e := CMT; e != items[5].Type {
		t.Errorf("CMT with invalid ending expected: %s, was: %s",
			e, items[5].Type)
	}
	if e := 6; len(items) != e {
		t.Errorf("Invalid number of comments expected: %d, was: %d",
			len(items), e)
	}
}

func TestLexerSub(t *testing.T) {
	in := `$name: foo;
$attr: border;
p.#{$name} {
  #{$attr}-color: blue;
}`
	items, err := testParse(in)

	if err != nil {
		panic(err)
	}
	vals := map[int]string{
		4:  "$attr",
		13: "#{",
		0:  "$name",
	}
	errors := false
	for i, v := range vals {
		if v != items[i].Value {
			errors = true
			t.Errorf("at %d expected: %s, was: %s", i, v, items[i].Value)
		}
	}
	if errors {
		printItems(items)
	}
}

func TestLexerCmds(t *testing.T) {
	in := `$s: sprite-map("test/*.png");
$file: sprite-file($s, 140);
div {
  width: image-width($file, 140);
  height: image-height(sprite-file($s, 140));
  url: sprite-file($s, 140);
}`
	items, err := testParse(in)
	if err != nil {
		panic(err)
	}

	types := map[int]ItemType{
		0:  VAR,
		2:  CMDVAR,
		4:  FILE,
		7:  VAR,
		9:  CMD,
		11: SUB,
		12: FILE,
		17: TEXT,
		19: CMD,
		21: SUB,
		22: FILE,
		27: CMD,
		29: CMD,
		32: FILE,
		40: SUB,
		41: FILE,
	}
	errors := false
	for i, tp := range types {
		if tp != items[i].Type {
			errors = true
			t.Errorf("at %d expected: %s, was: %s", i, tp, items[i].Type)
		}
	}
	if errors {
		printItems(items)
	}
}

func TestLexerImport(t *testing.T) {
	fvar := `@import "var";
`
	items, _ := testParse(fvar)
	vals := map[int]string{
		0: "@import",
		1: "var",
		2: ";",
	}
	errors := false
	for i, v := range vals {
		if v != items[i].Value {
			errors = true
			t.Errorf("at %d expected: %s, was: %s", i, v, items[i].Value)
		}
	}
	if errors {
		printItems(items)
	}
}

func TestLexerBasicVars(t *testing.T) {
	in := `$color: "black";
$color: red;
$background: "blue";

a {
  color: $color;
  background: $background;
}

$y: before;

$x: 1 2 $y;

foo {
  a: $x;
}

$y: after;

foo {
  a: $x;
}`

	items, err := testParse(in)
	if err != nil {
		t.Fatal(err)
	}
	for _, t := range items {
		fmt.Print(t.Type, t)
	}
}

// Test disabled due to not working
func TestLexerSubModifiers(t *testing.T) {
	in := `$s: sprite-map("*.png");
div {
  height: -1 * sprite-height($s,"140");
  width: -sprite-width($s,"140");
  margin: - sprite-height($s, "140")px;
  height: image-height(test/140.png);
  width: image-width(sprite-file($s, 140));
}`

	items, err := testParse(in)
	if err != nil {
		panic(err)
	}
	if e := ":"; items[1].Value != e {
		t.Errorf("Failed to parse symbol expected: %s, was: %s",
			e, items[1].Value)
	}
	if e := "*.png"; items[4].Value != e {
		t.Errorf("Failed to parse file expected: %s, was: %s",
			e, items[4].Value)
	}

	if e := "*"; items[13].Value != e {
		t.Errorf("Failed to parse text expected: %s, was: %s",
			e, items[13].Value)
	}

	if e := MINUS; items[22].Type != e {
		t.Errorf("Failed to parse CMD expected: %s, was: %s",
			e, items[22].Type)
	}

	if e := CMD; items[23].Type != e {
		t.Errorf("Failed to parse CMD expected: %s, was: %s",
			e, items[23].Type)
	}

	if e := TEXT; int(items[37].Type) != e {
		t.Errorf("Failed to parse TEXT expected: %s, was: %s",
			e, items[37].Type)
	}

	if e := FILE; int(items[43].Type) != e {
		t.Errorf("Type mismatch expected: %s, was: %s", e, items[43].Type)
	}
	types := map[int]ItemType{
		48: CMD,
		50: CMD,
		52: SUB,
		53: FILE,
	}
	for i, ty := range types {
		if types[i] != ty {
			t.Errorf("Type mismatch at %d expected: %s, was: %s", i, types[i], ty)
		}
	}
}

func TestLexerVars(t *testing.T) {
	in := `$a: 1;
$b: $1;
$c: ();
$d: $c`

	items, err := testParse(in)
	if err != nil {
		panic(err)
	}
	_ = items
}

func TestLexerWhitespace(t *testing.T) {
	in := `$s: sprite-map("*.png");
div {
  background:sprite($s,"140");
}`
	items, err := testParse(in)
	if err != nil {
		panic(err)
	}

	if e := TEXT; int(items[9].Type) != e {
		t.Errorf("Type parsed improperly expected: %s, was: %s",
			e, items[9].Type)
	}

	if e := CMD; items[11].Type != e {
		t.Errorf("Type parsed improperly expected: %s, was: %s",
			e, items[11].Type)
	}

	if e := "sprite"; items[11].Value != e {
		t.Errorf("Command parsed improperly expected: %s, was: %s",
			e, items[11].Value)
	}
}

// create a parser for the language.
func testParse(input string) ([]Item, error) {
	lex := New(func(lex *Lexer) StateFn {
		return lex.Action()
	}, input)

	var status []Item
	for {
		item := lex.Next()
		err := item.Error()

		if err != nil {
			return nil, fmt.Errorf("Error: %v (pos %d)", err, item.Pos)
		}
		switch item.Type {
		case ItemEOF:
			return status, nil
		case CMD, SPRITE, TEXT, VAR, FILE, SUB:
			fallthrough
		case LPAREN, RPAREN,
			LBRACKET, RBRACKET:
			fallthrough
		case IMPORT:
			fallthrough
		case EXTRA:
			status = append(status, *item)
		default:
			status = append(status, *item)
			//fmt.Printf("Default: %d %s\n", item.Pos, item)
		}
	}
}

func TestLexerLookup(t *testing.T) {
	it := Lookup("sprite-file")
	if e := "sprite-file"; it.String() != e {
		t.Errorf("Directive should be found was: %s, expected: %s",
			it.String(), e)
	}
	it = Lookup("NOT GONNA FIND")
	if e := ""; it.String() != e {
		t.Errorf("Not a token was: %s, expected: %s", it.String(), e)
	}
	it = Lookup("/")
	if e := ""; it.String() != e {
		t.Errorf("Non-directive was: %s, expected: %s", it.String(), e)
	}
}
