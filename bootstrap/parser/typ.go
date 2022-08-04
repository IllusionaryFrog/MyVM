package parser

type Ast struct {
	Imports []*Import
	Lets    []*Let
	Funs    []*Fun
}

type Import struct {
	Path *String
}

type Fun struct {
	Opts    []*Ident
	Ident   *Ident
	Inputs  []Typ
	Outputs []Typ
	Block   *Block
}

type Ident struct {
	DefaultExpr
	Content string
}

func (e *Ident) AsIdent() *Ident {
	return e
}

type Typ interface {
	String() string
	Size() int
	Sub() (bool, []Typ)
}

type Builtin string

func (b Builtin) String() string {
	return string(b)
}

func (b Builtin) Size() int {
	switch b {
	case VOID:
		return 0
	case U8, I8, CHAR:
		return 1
	case U16, I16:
		return 2
	case U32, I32:
		return 4
	case U64, I64:
		return 8
	case U128, I128, STRING:
		return 16
	default:
		panic("unreachable")
	}
}

func (b Builtin) Sub() (bool, []Typ) {
	switch b {
	case VOID:
		return true, []Typ{}
	case U8:
		return false, []Typ{I8}
	case I8:
		return false, []Typ{U8}
	case U16:
		return false, []Typ{I16}
	case I16:
		return false, []Typ{U16}
	case U32:
		return false, []Typ{I32}
	case I32:
		return false, []Typ{U32}
	case U64:
		return false, []Typ{I64}
	case I64:
		return false, []Typ{U64}
	case U128:
		return false, []Typ{I128}
	case I128:
		return false, []Typ{U128}
	case CHAR:
		return false, []Typ{U8}
	case STRING:
		return false, []Typ{U64, U64}
	default:
		panic("unreachable")
	}
}

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

	VOID   Builtin = "VOID"
	STRING Builtin = "STRING"
	CHAR   Builtin = "CHAR"
)

type Block struct {
	Lets  []*Let
	Exprs []Expr
}

type Let struct {
	Ident *Ident
	Typ   Typ
	Exprs []Expr
}

type Expr interface {
	AsIdent() *Ident
	AsCall() *Call
	AsNumber() *Number
	AsString() *String
	AsChar() *Char
	AsIf() *If
	AsUnwrap() *Unwrap
}

type DefaultExpr struct{}

func (e DefaultExpr) AsIdent() *Ident {
	return nil
}

func (e DefaultExpr) AsCall() *Call {
	return nil
}

func (e DefaultExpr) AsNumber() *Number {
	return nil
}

func (e DefaultExpr) AsString() *String {
	return nil
}

func (e DefaultExpr) AsChar() *Char {
	return nil
}

func (e *DefaultExpr) AsIf() *If {
	return nil
}

func (e *DefaultExpr) AsUnwrap() *Unwrap {
	return nil
}

type Call struct {
	DefaultExpr
	Ident   *Ident
	Inputs  []Typ
	Outputs []Typ
}

func (e *Call) AsCall() *Call {
	return e
}

type Number struct {
	DefaultExpr
	Content string
	Typ     Typ
	Size    int
	Base    int
}

func (e *Number) AsNumber() *Number {
	return e
}

type String struct {
	DefaultExpr
	Content string
}

func (e *String) AsString() *String {
	return e
}

type Char struct {
	DefaultExpr
	Content string
}

func (e *Char) AsChar() *Char {
	return e
}

type If struct {
	DefaultExpr
	Con   []Expr
	Exprs []Expr
	Else  []Expr
}

func (e *If) AsIf() *If {
	return e
}

type Unwrap struct {
	DefaultExpr
}

func (e *Unwrap) AsUnwrap() *Unwrap {
	return e
}
