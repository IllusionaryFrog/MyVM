package parser

type Ast struct {
	Funs []Fun
}

type Fun struct {
	Opts    []Ident
	Ident   Ident
	Inputs  []Typ
	Outputs []Typ
	Block   Block
}

type Ident struct {
	Content string
}

type Typ interface{}

type Builtin string

const (
	U8   Builtin = "U8"
	U16  Builtin = "U16"
	U32  Builtin = "U32"
	U64  Builtin = "U64"
	U128 Builtin = "U128"
	I8   Builtin = "I8"
	I16  Builtin = "I16"
	I32  Builtin = "I32"
	I64  Builtin = "I64"
	I128 Builtin = "I128"
)

type Block struct {
	Lets  []Let
	Exprs []Expr
}

type Let struct {
	Ident Ident
	Typ   Typ
	Exprs []Expr
}

type Expr interface{}

type Number struct {
	Content string
}

type String struct {
	Content string
}

type Char struct {
	Content string
}
