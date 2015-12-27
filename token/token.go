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
	COMMENT

	literal_beg
	// Identifiers
	IDENT
	INT
	FLOAT
	TEXT
	RULE
	STRING // "main.scss"
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

	keyword_beg
	IF      // @if
	ELSE    // @else
	EACH    // @each
	IMPORT  // @import
	INCLUDE // @include
	FUNC    // @function
	MIXIN   // @mixin

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

	include_mixin_beg
	FILE
	BKND
	include_mixin_end
	FIN
)

var Tokens = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
	COMMENT: "comment",

	IDENT: "IDENT",
	INT:   "INT",
	FLOAT: "FLOAT",

	CMDVAR:  "command-variable",
	VALUE:   "value",
	FILE:    "file",
	SPRITE:  "sprite",
	SPRITEF: "sprite-file",
	SPRITED: "sprite-dimensions",
	SPRITEH: "sprite-height",
	SPRITEW: "sprite-width",
	TEXT:    "text",
	RULE:    "rule",

	ADD: "+",
	SUB: "-",
	MUL: "*",
	QUO: "/",
	REM: "%",

	AND: "&",
	//OR: "|",
	XOR: "^",

	AT:     "@",
	EQL:    "==",
	LSS:    "<",
	GTR:    ">",
	ASSIGN: "=",
	NOT:    "!",

	DOLLAR: "$",

	NEQ:    "!=",
	LEQ:    "<=",
	GEQ:    ">=",
	DEFINE: ":=",

	LPAREN: "(",
	LBRACK: "[",
	LBRACE: "{",
	COMMA:  ",",
	PERIOD: ".",

	RPAREN:    ")",
	RBRACK:    "]",
	RBRACE:    "}",
	SEMICOLON: ";",
	COLON:     ":",

	NUMBER: "#",
	QUOTE:  "\"",

	IF:      "@if",
	ELSE:    "@else",
	EACH:    "@each",
	IMPORT:  "@import",
	INCLUDE: "@include",
	FUNC:    "@function",
	MIXIN:   "@mixin",

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

// A set of constants for precedence-based expression parsing.
// Non-operators have lowest precedence, followed by operators
// starting with precedence 1 up to unary operators. The highest
// precedence serves as "catch-all" precedence for selector,
// indexing, and other operator and delimiter tokens.
//
const (
	LowestPrec  = 0 // non-operators
	UnaryPrec   = 6
	HighestPrec = 7
)

// Precedence returns the operator precedence of the binary
// operator op. If op is not a binary operator, the result
// is LowestPrecedence.
//
func (op Token) Precedence() int {
	switch op {
	case LOR:
		return 1
	case LAND:
		return 2
	case EQL, NEQ, LSS, LEQ, GTR, GEQ:
		return 3
	case ADD, SUB, OR, XOR:
		return 4
	case MUL, QUO, REM, SHL, SHR, AND, AND_NOT:
		return 5
	}
	return LowestPrec
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
