package compiler

import (
	"bootstrap/parser"
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

func Compile(ast parser.Ast) []uint8 {
	funs := make(map[funIdent]*Fun)
	for i := 0; i < len(ast.Funs); i++ {
		fun := ast.Funs[i]
		ident := makeFunIdent(fun.Ident.Content, fun.Inputs, fun.Outputs)
		if funs[ident] != nil {
			panic(fmt.Sprintf("the fun '%s' already exists", ident))
		}
		funs[ident] = &Fun{fun: fun}
	}
	c := Ctx{funs: funs}
	return c.compile()
}

type funIdent string

func (f *Fun) makeFunIdent() funIdent {
	return makeFunIdent(f.fun.Ident.Content, f.fun.Inputs, f.fun.Outputs)
}

func makeFunIdent(ident string, inputs []parser.Typ, outputs []parser.Typ) funIdent {
	var buffer bytes.Buffer
	buffer.WriteString(ident)
	buffer.WriteString("(")
	for i := 0; i < len(inputs); i++ {
		if i != 0 {
			buffer.WriteString(",")
		}
		buffer.WriteString(inputs[i].String())
	}
	buffer.WriteString(":")
	for i := 0; i < len(outputs); i++ {
		if i != 0 {
			buffer.WriteString(",")
		}
		buffer.WriteString(outputs[i].String())
	}
	buffer.WriteString(")")
	return funIdent(buffer.String())
}

type Info struct {
	asm      bool
	inline   bool
	tailCall bool
	size     uint64
	pos      uint64
}

type Fun struct {
	info *Info
	fun  *parser.Fun
}

func (f *Fun) getInfo(c *Ctx) *Info {
	if f.info == nil {
		f.comInfo(c)
	}
	return f.info
}

func containsOpt(opts []*parser.Ident, opt string) bool {
	for i := 0; i < len(opts); i++ {
		if opts[i].Content == opt {
			return true
		}
	}
	return false
}

func (f *Fun) comInfo(c *Ctx) {
	if f.info != nil {
		return
	}

	if len(f.fun.Block.Lets) != 0 {
		panic("unimplemented")
	}

	f.info = &Info{}

	f.info.asm = containsOpt(f.fun.Opts, "asm")
	f.info.inline = containsOpt(f.fun.Opts, "inline")
	if f.info.asm && !f.info.inline {
		panic(fmt.Sprintf("the asm fun '%s' needs to also be inline", f.makeFunIdent()))
	}
	last := f.fun.Block.Exprs[len(f.fun.Block.Exprs)-1].GetCall()
	if last != nil {
		fun := c.funs[makeFunIdent(last.Ident.Content, last.Inputs, last.Outputs)]
		if fun != nil {
			f.info.tailCall = !fun.getInfo(c).inline
		}
	}

	if f.info.asm {
		f.info.size += uint64(len(f.fun.Block.Exprs))
	} else {
		for i := 0; i < len(f.fun.Block.Exprs); i++ {
			expr := f.fun.Block.Exprs[i]
			ident := expr.GetIdent()
			if ident != nil {
				panic("unimplemented")
			}
			call := expr.GetCall()
			if call != nil {
				ident := makeFunIdent(call.Ident.Content, call.Inputs, call.Outputs)
				fun := c.funs[ident]
				if fun == nil {
					panic(fmt.Sprintf("unknown fun '%s'", ident))
				}
				finfo := fun.getInfo(c)
				if finfo.inline {
					f.info.size += finfo.size
				} else {
					f.info.size += 1 + 8
				}
			}
			number := expr.GetNumber()
			if number != nil {
				f.info.size += 1 + uint64(number.Size)
			}
			str := expr.GetString()
			if str != nil {
				c.pushStr(str.Content)
				f.info.size += 1 + 8 + 1 + 8
			}
			char := expr.GetChar()
			if char != nil {
				f.info.size += 1 + 1
			}
		}
	}

	if !f.info.inline {
		if !f.info.tailCall {
			f.info.size += 1
		}
		f.info.pos = c.getNextPos(f.info.size)
	}
}

type Ctx struct {
	size uint64
	strs string
	funs map[funIdent]*Fun
}

func initialBytes() []uint8 {
	return []uint8{220, 0, 0, 0, 0, 0, 0, 0, 0}
}

func finalBytes() []uint8 {
	return []uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
}

func (c *Ctx) compile() []uint8 {
	bytes := initialBytes()
	c.size = uint64(len(bytes))

	start := c.funs[makeFunIdent("__start", []parser.Typ{}, []parser.Typ{})]

	if start == nil {
		panic("no __start(:) fun")
	}

	info := start.getInfo(c)
	if info.inline || info.asm {
		panic("__start(:) can't be a inline or asm fun")
	}

	start.typeCheck(c)

	putUvarint(bytes[1:9], info.pos)
	bytes = append(bytes, start.compile(c)...)
	bytes = append(bytes, []uint8(c.strs)...)

	return append(bytes, finalBytes()...)
}

func (c *Ctx) getNextPos(size uint64) uint64 {
	res := c.size
	c.size += size
	return res
}

func (f *Fun) compile(c *Ctx) []uint8 {
	bytes := []uint8{}

	if len(f.fun.Block.Lets) != 0 {
		panic("unimplemented")
	}

	info := f.getInfo(c)

	if info.asm {
		bytes = append(bytes, f.compileAsm()...)
		if !info.inline {
			bytes = append(bytes, 3)
		}
	} else {
		for i := 0; i < len(f.fun.Block.Exprs); i++ {
			expr := f.fun.Block.Exprs[i]
			call := expr.GetCall()
			if call != nil {
				ident := makeFunIdent(call.Ident.Content, call.Inputs, call.Outputs)
				fun := c.funs[ident]
				finfo := fun.getInfo(c)
				if !finfo.inline {
					bytes = append(bytes, fun.compile(c)...)
				}
			}
		}
		for i := 0; i < len(f.fun.Block.Exprs); i++ {
			expr := f.fun.Block.Exprs[i]
			ident := expr.GetIdent()
			if ident != nil {
				panic("unimplemented")
			}
			call := expr.GetCall()
			if call != nil {
				ident := makeFunIdent(call.Ident.Content, call.Inputs, call.Outputs)
				fun := c.funs[ident]
				finfo := fun.getInfo(c)
				if finfo.inline {
					bytes = append(bytes, fun.compile(c)...)
				} else {
					buf := []uint8{229, 0, 0, 0, 0, 0, 0, 0, 0}
					if info.tailCall {
						buf[0] = 220
					}
					putUvarint(buf[1:], finfo.pos)
					bytes = append(bytes, buf...)

				}
			}
			number := expr.GetNumber()
			if number != nil {
				num, err := strconv.ParseUint(number.Content, number.Base, number.Size*8)
				if err != nil {
					panic(fmt.Sprintf("unable to convert '%s' to a number", number.Content))
				}
				var buf []uint8
				switch number.Size {
				case 1:
					buf = []uint8{10, 0}
				case 2:
					buf = []uint8{11, 0, 0}
				case 4:
					buf = []uint8{12, 0, 0, 0, 0}
				case 8:
					buf = []uint8{13, 0, 0, 0, 0, 0, 0, 0, 0}
				default:
					panic("invalid size")
				}
				putUvarint(buf[1:], num)
				bytes = append(bytes, buf...)
			}
			str := expr.GetString()
			if str != nil {
				ptr := c.getStr(str.Content)
				buf := []uint8{13, 0, 0, 0, 0, 0, 0, 0, 0, 13, 0, 0, 0, 0, 0, 0, 0, 0}
				putUvarint(buf[1:9], ptr)
				putUvarint(buf[10:], uint64(len(str.Content)))
				bytes = append(bytes, buf...)
			}
			char := expr.GetChar()
			if char != nil {
				if len(char.Content) != 1 {
					panic(fmt.Sprintf("the char '%s' can only contain one byte", char.Content))
				}
				bytes = append(bytes, 10, char.Content[0])
			}
		}
	}

	if !info.inline && !info.tailCall {
		bytes = append(bytes, 3)
	}

	return bytes
}

func (c *Ctx) pushStr(s string) {
	if strings.Index(c.strs, s) == -1 {
		c.strs += s
	}
}

func (c *Ctx) getStr(s string) uint64 {
	return uint64(strings.Index(c.strs, s)) + c.size
}

func putUvarint(buf []uint8, x uint64) {
	i := 0
	for x > 0xff {
		buf[i] = uint8(x)
		x >>= 8
		i++
	}
	buf[i] = uint8(x)
}
