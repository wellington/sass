package token

// A type for all the types of items in the language being lexed.
// These only parse SASS specific language elements and not CSS.
type Token int

// const ItemEOF = 0
const NotFound = -1

// Special item types.
const (
	ILLEGAL Token = iota
	EOF
	CMT
	IDENT

	keyword_beg
	IF
	ELSE
	EACH
	IMPORT
	INCLUDE
	INTP
	FUNC
	MIXIN
	EXTRA
	CMD
	VAR
	keyword_end

	CMDVAR
	VALUE

	cmd_beg
	SPRITE
	SPRITEF
	SPRITED
	SPRITEH
	SPRITEW
	cmd_end

	literal_beg
	INT
	FLOAT
	TEXT
	RULE
	literal_end

	operator_beg
	ADD // +
	SUB // -
	MUL // *
	QUO // /
	REM // %

	AND     // &
	OR      // |
	XOR     // ^
	SHL     // <<
	SHR     // >>
	AND_NOT // &^

	LAND  // &&
	LOR   // ||
	ARROW // <-
	INC   // ++
	DEC   // --

	EQL    // ==
	LSS    // <
	GTR    // >
	ASSIGN // =
	NOT    // !

	NEQ      // !=
	LEQ      // <=
	GEQ      // >=
	DEFINE   // :=
	ELLIPSIS // ...

	AT     // @
	DOLLAR // $
	NUMBER // #
	QUOTE  // "

	LPAREN // (
	LBRACK // [
	LBRACE // {
	COMMA  // ,
	PERIOD // .

	RPAREN    // )
	RBRACK    // ]
	RBRACE    // }
	SEMICOLON // ;
	COLON     // :
	operator_end

	include_mixin_beg
	FILE
	BKND
	include_mixin_end
	FIN
)

var Tokens = [...]string{
	EOF: "eof",
	CMT: "comment",

	IDENT:   "IDENT",
	IF:      "@if",
	ELSE:    "@else",
	EACH:    "@each",
	IMPORT:  "@import",
	INCLUDE: "@include",
	INTP:    "#{",
	FUNC:    "@function",
	MIXIN:   "@mixin",
	EXTRA:   "extra",
	CMD:     "command",
	VAR:     "variable",
	CMDVAR:  "command-variable",
	VALUE:   "value",
	FILE:    "file",
	SPRITE:  "sprite",
	SPRITEF: "sprite-file",
	SPRITED: "sprite-dimensions",
	SPRITEH: "sprite-height",
	SPRITEW: "sprite-width",
	INT:     "integer",
	FLOAT:   "float",
	TEXT:    "text",
	RULE:    "rule",

	XOR: "^",

	DOLLAR:    "$",
	ADD:       "+",
	SUB:       "-",
	MUL:       "*",
	QUO:       "/",
	LPAREN:    "(",
	RPAREN:    ")",
	LBRACE:    "{",
	RBRACE:    "}",
	SEMICOLON: ";",
	COLON:     ":",
	NUMBER:    "#",
	QUOTE:     "\"",

	BKND: "background",
	FIN:  "FINISHED",
}

func (i Token) String() string {
	if i < 0 {
		return ""
	}
	return Tokens[i]
}

var directives map[string]Token

func init() {
	directives = make(map[string]Token)
	for i := cmd_beg; i < cmd_end; i++ {
		directives[Tokens[i]] = i
	}
}

// Lookup Token by token string
func Lookup(ident string) Token {
	if tok, is_keyword := directives[ident]; is_keyword {
		return tok
	}
	return NotFound
}

// Predicates

// IsLiteral returns true for tokens corresponding to identifiers
// and basic type literals; it returns false otherwise.
//
func (tok Token) IsLiteral() bool { return literal_beg < tok && tok < literal_end }

// IsOperator returns true for tokens corresponding to operators and
// delimiters; it returns false otherwise.
//
func (tok Token) IsOperator() bool { return operator_beg < tok && tok < operator_end }

// IsKeyword returns true for tokens corresponding to keywords;
// it returns false otherwise.
//
func (tok Token) IsKeyword() bool { return keyword_beg < tok && tok < keyword_end }
