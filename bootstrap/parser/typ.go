package parser

type Ast struct {
	Imports []*Import
	Lets    []*Let
	Funs    []*Fun
	Types   []*Type
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
	String(map[string]*Type) string
	Size(map[string]*Type) int
	Sub(map[string]*Type) []Typ
	LoadSizes(map[string]*Type) []int
}

type Custom struct {
	Ident string
}

type Builtin string

func (b Builtin) LoadSizes(map[string]*Type) []int {
	switch b {
	case U8, I8, BOOL:
		return []int{1}
	case U16, I16:
		return []int{2}
	case U32, I32:
		return []int{4}
	case U64, I64, ADDR:
		return []int{8}
	case U128, I128:
		return []int{16}
	case STRING:
		return []int{8, 8}
	default:
		panic("unreachable")
	}
}

func (b Builtin) String(_ map[string]*Type) string {
	return string(b)
}

func (b Builtin) Size(_ map[string]*Type) int {
	switch b {
	case U8, I8, BOOL:
		return 1
	case U16, I16:
		return 2
	case U32, I32:
		return 4
	case U64, I64, ADDR:
		return 8
	case U128, I128, STRING:
		return 16
	default:
		panic("unreachable")
	}
}

func (b Builtin) Sub(_ map[string]*Type) []Typ {
	switch b {
	case U8:
		return []Typ{I8}
	case I8, BOOL:
		return []Typ{U8}
	case U16:
		return []Typ{I16}
	case I16:
		return []Typ{U16}
	case U32:
		return []Typ{I32}
	case I32:
		return []Typ{U32}
	case U64:
		return []Typ{I64}
	case I64, ADDR:
		return []Typ{U64}
	case U128:
		return []Typ{I128}
	case I128:
		return []Typ{U128}
	case STRING:
		return []Typ{U64, U64}
	default:
		panic("unreachable")
	}
}

type Type struct {
	Opts   []*Ident
	Ident  *Ident
	Fields []Typ
}

func (t *Custom) String(ts map[string]*Type) string {
	return ts[t.Ident].Ident.Content
}

func (t *Custom) LoadSizes(ts map[string]*Type) []int {
	sizes := []int{}
	for _, f := range ts[t.Ident].Fields {
		sizes = append(sizes, f.LoadSizes(ts)...)
	}
	return sizes
}

func (t *Custom) Size(ts map[string]*Type) int {
	size := 0
	for _, f := range ts[t.Ident].Fields {
		size += f.Size(ts)
	}
	return size
}

func (t *Custom) Sub(ts map[string]*Type) []Typ {
	return ts[t.Ident].Fields
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

	STRING Builtin = "STRING"
	BOOL   Builtin = "BOOL"
	ADDR   Builtin = "ADDR"
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
	AsIf() *If
	AsUnwrap() *Unwrap
	AsWrap() *Wrap
	AsAddr() *Addr
}

type DefaultExpr struct{}

func (e DefaultExpr) AsWrap() *Wrap {
	return nil
}

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

func (e *DefaultExpr) AsIf() *If {
	return nil
}

func (e *DefaultExpr) AsUnwrap() *Unwrap {
	return nil
}

func (e *DefaultExpr) AsAddr() *Addr {
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

type Wrap struct {
	DefaultExpr
	Typ Typ
}

func (e *Wrap) AsWrap() *Wrap {
	return e
}

type Addr struct {
	DefaultExpr
	Ident *Ident
	Call  *Call
}

func (e *Addr) AsAddr() *Addr {
	return e
}
