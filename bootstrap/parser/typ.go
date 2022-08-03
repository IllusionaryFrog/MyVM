package parser

type Ast struct {
	Lets []*Let
	Funs []*Fun
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
}

type Builtin string

func (b *Builtin) String() string {
	return string(*b)
}

func (b *Builtin) Size() int {
	switch *b {
	case U8:
		return 1
	case U16:
		return 2
	case U32:
		return 4
	case U64:
		return 8
	case U128:
		return 16
	case I8:
		return 1
	case I16:
		return 2
	case I32:
		return 4
	case I64:
		return 8
	case I128:
		return 16
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

func (e *Char) GetCall() *Call {
	return nil
}

func (e *Char) GetNumber() *Number {
	return nil
}

func (e *Char) GetString() *String {
	return nil
}

func (e *Char) GetChar() *Char {
	return e
}
