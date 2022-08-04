package lexer

type Typ string

const (
	EOF     Typ = "EOF"
	IGNORED Typ = "IGNORED"

	COLON     Typ = ":"
	SEMICOLON Typ = ";"
	COMMA     Typ = ","
	EQUALS    Typ = "="
	COMMENT   Typ = "#"
	LPAREN    Typ = "("
	RPAREN    Typ = ")"
	LBRACE    Typ = "{"
	RBRACE    Typ = "}"

	CHAR   Typ = "CHAR"
	STRING Typ = "STRING"
	NUMBER Typ = "NUMBER"

	FUN    Typ = "FUN"
	LET    Typ = "LET"
	IF     Typ = "IF"
	ELSE   Typ = "ELSE"
	IMPORT Typ = "IMPORT"
	UNWRAP Typ = "UNWRAP"

	IDENT Typ = "IDENT"
)

type Token struct {
	Typ     Typ
	Content string
}
