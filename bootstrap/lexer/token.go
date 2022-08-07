package lexer

type Typ string

const (
	EOF     Typ = "EOF"
	IGNORED Typ = "IGNORED"

	COLON     Typ = ":"
	SEMICOLON Typ = ";"
	COMMA     Typ = ","
	LPAREN    Typ = "("
	RPAREN    Typ = ")"
	LBRACE    Typ = "{"
	RBRACE    Typ = "}"

	COMMENT Typ = "//"

	STRING Typ = "STRING"
	NUMBER Typ = "NUMBER"

	FUN    Typ = "FUN"
	LET    Typ = "LET"
	IF     Typ = "IF"
	ELSE   Typ = "ELSE"
	IMPORT Typ = "IMPORT"
	UNWRAP Typ = "UNWRAP"
	TYPE   Typ = "TYPE"
	WRAP   Typ = "WRAP"
	ADDR   Typ = "ADDR"
	RETURN Typ = "RETURN"
	WHILE  Typ = "WHILE"

	IDENT Typ = "IDENT"
)

type Token struct {
	Typ     Typ
	Content string
}
