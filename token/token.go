package token

// A type for all the types of items in the language being lexed.
// These only parse SASS specific language elements and not CSS.
type Token int

// const ItemEOF = 0
const NotFound = -1

// Special item types.
const (
	EOF Token = iota
	Error
	SPACE
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
	CMDVAR
	SUB
	VALUE
	// FILE
	cmd_beg
	SPRITE
	SPRITEF
	SPRITED
	SPRITEH
	SPRITEW
	cmd_end
	NUMBER
	TEXT
	RULE
	DOLLAR
	math_beg
	PLUS
	MINUS
	MULT
	DIVIDE
	math_end
	special_beg
	LPAREN
	RPAREN
	LBRACKET
	RBRACKET
	SEMIC
	COLON
	CMT
	special_end
	include_mixin_beg
	FILE
	BKND
	include_mixin_end
	FIN
)

var Tokens = [...]string{
	EOF:      "eof",
	Error:    "error",
	IF:       "@if",
	ELSE:     "@else",
	EACH:     "@each",
	IMPORT:   "@import",
	INCLUDE:  "@include",
	INTP:     "#{",
	FUNC:     "@function",
	MIXIN:    "@mixin",
	EXTRA:    "extra",
	CMD:      "command",
	VAR:      "variable",
	CMDVAR:   "command-variable",
	SUB:      "sub",
	VALUE:    "value",
	FILE:     "file",
	SPRITE:   "sprite",
	SPRITEF:  "sprite-file",
	SPRITED:  "sprite-dimensions",
	SPRITEH:  "sprite-height",
	SPRITEW:  "sprite-width",
	NUMBER:   "number",
	TEXT:     "text",
	RULE:     "rule",
	DOLLAR:   "$",
	PLUS:     "+",
	MINUS:    "-",
	MULT:     "*",
	DIVIDE:   "/",
	LPAREN:   "(",
	RPAREN:   ")",
	LBRACKET: "{",
	RBRACKET: "}",
	SEMIC:    ";",
	COLON:    ":",
	CMT:      "comment",
	BKND:     "background",
	FIN:      "FINISHED",
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
