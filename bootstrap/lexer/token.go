package lexer

const (
	EOF     = "EOF"
	IGNORED = "IGNORED"

	COLON     = ":"
	SEMICOLON = ";"
	COMMA     = ","
	EQUALS    = "="
	COMMENT   = "#"

	LPAREN = "("
	RPAREN = ")"
	LBRACE = "{"
	RBRACE = "}"

	CHAR   = "CHAR"
	STRING = "STRING"
	NUMBER = "NUMBER"

	FUN   = "FUN"
	TYPE  = "TYPE"
	LET   = "LET"
	IF    = "IF"
	ELSE  = "ELSE"
	WHILE = "WHILE"

	IDENT = "IDENT"
)

type Typ string

type Token struct {
	Typ     Typ
	Content string
}
