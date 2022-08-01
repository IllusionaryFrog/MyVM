package compiler

import (
	"bootstrap/parser"
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
)

func Compile(ast parser.Ast) []byte {
	funs := make(map[funIdent]*Fun)
	for i := 0; i < len(ast.Funs); i++ {
		fun := ast.Funs[i]
		ident := makeFunIdent(fun.Ident.Content, fun.Inputs, fun.Outputs)
		funs[ident] = &Fun{fun: fun}
	}
	c := ctx{funs: funs}
	return c.compile()
}

type funIdent string

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

func (f *Fun) getInfo(c *ctx) *Info {
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

func (f *Fun) comInfo(c *ctx) {
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
		panic("non inlined asm")
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
					fmt.Println(ident)
					panic("unkown function")
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
				f.info.size += 1 + number.Size
			}
			str := expr.GetString()
			if str != nil {
				panic("unimplemented")
			}
			c := expr.GetChar()
			if c != nil {
				panic("unimplemented")
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

type ctx struct {
	size uint64
	funs map[funIdent]*Fun
}

func initialBytes() []byte {
	return []byte{220, 0, 0, 0, 0, 0, 0, 0, 0}
}

func finalBytes() []byte {
	return []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
}

func (c *ctx) compile() []byte {
	bytes := initialBytes()
	c.size = uint64(len(bytes))

	start := c.funs[makeFunIdent("__start", []parser.Typ{}, []parser.Typ{})]

	if start == nil {
		panic("no __start(:)")
	}

	info := start.getInfo(c)

	if info.inline || info.asm {
		panic("wrong __start(:)")
	}

	bytes = append(bytes, start.compile(c)...)

	binary.PutUvarint(bytes[1:9], start.info.pos)

	return append(bytes, finalBytes()...)
}

func (c *ctx) getNextPos(size uint64) uint64 {
	res := c.size
	c.size += size
	return res
}

func (f *Fun) compile(c *ctx) []byte {
	bytes := []byte{}

	if len(f.fun.Block.Lets) != 0 {
		panic("unimplemented")
	}

	info := f.getInfo(c)

	if info.asm {
		bytes = append(bytes, compileAsm(f.fun.Block.Exprs)...)
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
					buf := []byte{229, 0, 0, 0, 0, 0, 0, 0, 0}
					if info.tailCall {
						buf[0] = 220
					}
					binary.PutUvarint(buf[1:], finfo.pos)
					bytes = append(bytes, buf...)

				}
			}
			number := expr.GetNumber()
			if number != nil {
				num, err := strconv.ParseUint(number.Content, 10, 64)
				if err != nil {
					panic("invalid number")
				}
				buf := []byte{13, 0, 0, 0, 0, 0, 0, 0, 0}
				binary.PutUvarint(buf[1:], num)
				bytes = append(bytes, buf...)
			}
			str := expr.GetString()
			if str != nil {
				panic("unimplemented")
			}
			c := expr.GetChar()
			if c != nil {
				panic("unimplemented")
			}
		}
	}

	if !info.inline && !info.tailCall {
		bytes = append(bytes, 3)
	}

	return bytes
}
