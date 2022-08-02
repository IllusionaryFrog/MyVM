package parser

type Ast struct {
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
	Content string
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
	GetIdent() *Ident
	GetCall() *Call
	GetNumber() *Number
	GetString() *String
	GetChar() *Char
}

func (e *Ident) GetIdent() *Ident {
	return e
}

func (e *Ident) GetCall() *Call {
	return nil
}

func (e *Ident) GetNumber() *Number {
	return nil
}

func (e *Ident) GetString() *String {
	return nil
}

func (e *Ident) GetChar() *Char {
	return nil
}

type Call struct {
	Ident   *Ident
	Inputs  []Typ
	Outputs []Typ
}

func (e *Call) GetIdent() *Ident {
	return nil
}

func (e *Call) GetCall() *Call {
	return e
}

func (e *Call) GetNumber() *Number {
	return nil
}

func (e *Call) GetString() *String {
	return nil
}

func (e *Call) GetChar() *Char {
	return nil
}

type Number struct {
	Content string
	Size    int
	Base    int
}

func (e *Number) GetIdent() *Ident {
	return nil
}

func (e *Number) GetCall() *Call {
	return nil
}

func (e *Number) GetNumber() *Number {
	return e
}

func (e *Number) GetString() *String {
	return nil
}

func (e *Number) GetChar() *Char {
	return nil
}

type String struct {
	Content string
}

func (e *String) GetIdent() *Ident {
	return nil
}

func (e *String) GetCall() *Call {
	return nil
}

func (e *String) GetNumber() *Number {
	return nil
}

func (e *String) GetString() *String {
	return e
}

func (e *String) GetChar() *Char {
	return nil
}

type Char struct {
	Content string
}

func (e *Char) GetIdent() *Ident {
	return nil
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
