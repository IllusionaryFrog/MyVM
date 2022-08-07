package parser

import "fmt"

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

type Types struct {
	ts map[string]*Type
}

func NewTypes() *Types {
	ts := make(map[string]*Type)
	return &Types{ts: ts}
}

func (ts *Types) Set(ident string, typ *Type) bool {
	if ts.ts[ident] != nil {
		return true
	}
	ts.ts[ident] = typ
	return false
}

func (ts *Types) Get(ident string) *Type {
	typ := ts.ts[ident]
	if typ == nil {
		panic(fmt.Sprintf("unknown type '%s'", ident))
	}
	return typ
}

type Typ interface {
	String(*Types) string
	Size(*Types) int
	Sub(*Types) []Typ
	LoadSizes(*Types) []int
	IsNever() bool
}

type Custom struct {
	Ident string
}

type Builtin string

func (b Builtin) IsNever() bool {
	return b == NEVER
}

func (b Builtin) LoadSizes(*Types) []int {
	switch b {
	case U8, I8, BOOL:
		return []int{1}
	case U16, I16:
		return []int{2}
	case U32, I32:
		return []int{4}
	case U64, I64:
		return []int{8}
	case U128, I128:
		return []int{16}
	case STRING:
		return []int{8, 8}
	default:
		panic("unreachable")
	}
}

func (b Builtin) String(*Types) string {
	return string(b)
}

func (b Builtin) Size(*Types) int {
	switch b {
	case U8, I8, BOOL:
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

func (b Builtin) Sub(*Types) []Typ {
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

func (c *Custom) IsNever() bool {
	return false
}

func (t *Custom) String(ts *Types) string {
	return ts.Get(t.Ident).Ident.Content
}

func (t *Custom) LoadSizes(ts *Types) []int {
	sizes := []int{}
	for _, f := range ts.Get(t.Ident).Fields {
		sizes = append(sizes, f.LoadSizes(ts)...)
	}
	return sizes
}

func (t *Custom) Size(ts *Types) int {
	size := 0
	for _, f := range ts.Get(t.Ident).Fields {
		size += f.Size(ts)
	}
	return size
}

func (t *Custom) Sub(ts *Types) []Typ {
	return ts.Get(t.Ident).Fields
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
	NEVER  Builtin = "NEVER"
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
	AsReturn() *Return
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

func (e *DefaultExpr) AsReturn() *Return {
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

type Return struct {
	DefaultExpr
}

func (e *Return) AsReturn() *Return {
	return e
}
