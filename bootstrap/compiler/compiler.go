package compiler

import (
	"bootstrap/lexer"
	"bootstrap/parser"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func containsPath(path string, imported []string) bool {
	for _, imp := range imported {
		if path == imp {
			return true
		}
	}
	return false
}

func getAll(ast parser.Ast, imported []string) ([]string, []*parser.Fun, []*parser.Let) {
	funs := ast.Funs
	lets := ast.Lets
	for _, imp := range ast.Imports {
		path, err := filepath.Abs(imp.Path.Content)
		if err != nil {
			panic(fmt.Sprintf("invalid import path '%s'", imp.Path.Content))
		}
		base := filepath.Dir(path)
		if os.Chdir(base) != nil {
			panic(fmt.Sprintf("invalid import path '%s'", imp.Path.Content))
		}
		if !containsPath(path, imported) {
			imported = append(imported, path)
			dat, err := os.ReadFile(path)
			if err != nil {
				panic(fmt.Sprintf("invalid import path '%s'", imp.Path.Content))
			}
			newAst, perr := parser.Parse(lexer.New(string(dat)))
			if perr {
				panic(fmt.Sprintf("error parsing file '%s'", imp.Path.Content))
			}
			newImported, newFuns, newLets := getAll(newAst, imported)
			imported = newImported
			funs = append(funs, newFuns...)
			lets = append(lets, newLets...)

		}
	}
	return imported, funs, lets
}

func Compile(ast parser.Ast) []uint8 {
	_, allFuns, allLets := getAll(ast, []string{})

	funs := make(map[Ident]*Fun)
	for i := 0; i < len(allFuns); i++ {
		fun := allFuns[i]
		ident := makeFunIdent(fun.Ident.Content, fun.Inputs, fun.Outputs)
		if funs[ident] != nil {
			panic(fmt.Sprintf("the fun '%s' already exists", ident))
		}
		funs[ident] = &Fun{fun: fun}
	}
	lets := make(map[Ident]*Let)
	for i := 0; i < len(allLets); i++ {
		let := allLets[i]
		ident := Ident(let.Ident.Content)
		if lets[ident] != nil {
			panic(fmt.Sprintf("the let '%s' already exists", ident))
		}
		lets[ident] = &Let{let: let}
	}
	c := Ctx{lets: lets, funs: funs}
	return c.compile()
}

type Ident string

func (f *Fun) makeFunIdent() Ident {
	return makeFunIdent(f.fun.Ident.Content, f.fun.Inputs, f.fun.Outputs)
}

func makeFunIdent(ident string, inputs []parser.Typ, outputs []parser.Typ) Ident {
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
	return Ident(buffer.String())
}

type FInfo struct {
	inline          bool
	asm             bool
	tailCall        bool
	unsafe          bool
	allowUnsafe     bool
	simpleTypeCheck bool
	letSize         uint64
	size            uint64
	pos             uint64
	lets            map[Ident]*Let
}

type Fun struct {
	info *FInfo
	fun  *parser.Fun
}

func (f *Fun) getInfo(c *Ctx) *FInfo {
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

	f.info = &FInfo{}

	f.info.asm = containsOpt(f.fun.Opts, "asm")
	f.info.inline = containsOpt(f.fun.Opts, "inline")
	f.info.unsafe = containsOpt(f.fun.Opts, "unsafe")
	f.info.allowUnsafe = containsOpt(f.fun.Opts, "allow_unsafe")
	f.info.simpleTypeCheck = containsOpt(f.fun.Opts, "simple_type_check")

	if f.info.asm && !f.info.inline {
		panic(fmt.Sprintf("the asm fun '%s' needs to also be inline", f.makeFunIdent()))
	}
	if f.info.asm && !(f.info.unsafe || f.info.allowUnsafe) {
		panic(fmt.Sprintf("the asm fun '%s' needs to either be unsafe or allow unsafe", f.makeFunIdent()))
	}
	if f.makeFunIdent() != c.start && len(f.fun.Block.Exprs) != 0 {
		last := f.fun.Block.Exprs[len(f.fun.Block.Exprs)-1].AsCall()
		if last != nil {
			fun := c.funs[makeFunIdent(last.Ident.Content, last.Inputs, last.Outputs)]
			if fun != nil {
				f.info.tailCall = !fun.getInfo(c).inline
			}
		}
	}
	if f.info.simpleTypeCheck && !(f.info.unsafe || f.info.allowUnsafe) {
		panic(fmt.Sprintf("the simple type check fun '%s' needs to either be unsafe or allow unsafe", f.makeFunIdent()))
	}

	if len(f.fun.Block.Lets) != 0 {
		if f.info.inline {
			panic(fmt.Sprintf("the inline fun '%s' can't have lets", f.makeFunIdent()))
		}
		if f.info.asm {
			panic(fmt.Sprintf("the asm fun '%s' can't have lets", f.makeFunIdent()))
		}

		f.info.lets = make(map[Ident]*Let)
		for _, let := range f.fun.Block.Lets {
			ident := Ident(let.Ident.Content)
			if f.info.lets[ident] != nil {
				panic(fmt.Sprintf("the let '%s' already exists", ident))
			}
			f.info.lets[ident] = &Let{let: let}
			let := f.info.lets[ident]
			let.getInfo(c).pos = f.info.letSize
			size := uint64(let.let.Typ.Size())
			f.info.letSize += size
			f.info.size += f.sizeOfExprs(c, let.let.Exprs)
			if size != 0 {
				f.info.size += 1 + 8
			}
		}

		f.info.size += f.info.letSize
	}

	if f.info.asm {
		f.info.size += uint64(len(f.fun.Block.Exprs))
	} else {
		f.info.size += f.sizeOfExprs(c, f.fun.Block.Exprs)
	}

	if !f.info.inline {
		if !f.info.tailCall {
			f.info.size += 1
		}
		f.info.pos = c.getNextPos(f.info.size)
	}
}

func (f *Fun) sizeOfExprs(c *Ctx, exprs []parser.Expr) uint64 {
	var size uint64 = 0
	for i := 0; i < len(exprs); i++ {
		expr := exprs[i]
		ident := expr.AsIdent()
		if ident != nil {
			ident := Ident(ident.Content)
			let := f.info.lets[ident]
			if let == nil {
				let = c.lets[ident]
				if let == nil {
					panic(fmt.Sprintf("unknown ident '%s'", ident))
				}
				let.getInfo(c)
			}
			if let.info.size != 0 {
				size += 1 + 8
			}
		}
		call := expr.AsCall()
		if call != nil {
			ident := makeFunIdent(call.Ident.Content, call.Inputs, call.Outputs)
			fun := c.funs[ident]
			if fun == nil {
				panic(fmt.Sprintf("unknown fun '%s'", ident))
			}
			if fun.makeFunIdent() == c.start {
				panic(fmt.Sprintf("fun '%s' can't call '%s'", f.makeFunIdent(), c.start))
			}
			finfo := fun.getInfo(c)
			if finfo.unsafe && !(f.info.unsafe || f.info.allowUnsafe) {
				panic(fmt.Sprintf("fun '%s' can't call core fun '%s'", f.makeFunIdent(), ident))
			}
			if finfo.inline {
				size += finfo.size
			} else {
				size += 1 + 8
			}
		}
		number := expr.AsNumber()
		if number != nil {
			size += 1 + uint64(number.Size)
		}
		str := expr.AsString()
		if str != nil {
			c.pushStr(str.Content)
			size += 1 + 8 + 1 + 8
		}
		char := expr.AsChar()
		if char != nil {
			size += 1 + 1
		}
		ifel := expr.AsIf()
		if ifel != nil {
			size += f.sizeOfExprs(c, ifel.Con) + 1 + 8
			size += f.sizeOfExprs(c, ifel.Else) + 1 + 8
			size += f.sizeOfExprs(c, ifel.Exprs)
		}
		unwrap := expr.AsUnwrap()
		if unwrap != nil {
		}
	}
	return size
}

type LInfo struct {
	size uint64
	pos  uint64
}

type Let struct {
	info *LInfo
	let  *parser.Let
}

func (l *Let) getInfo(c *Ctx) *LInfo {
	if l.info == nil {
		l.comInfo(c)
	}
	return l.info
}

func (l *Let) comInfo(c *Ctx) {
	if l.info != nil {
		return
	}

	l.info = &LInfo{}
	l.info.size = uint64(l.let.Typ.Size())
}

type Ctx struct {
	size  uint64
	strs  string
	lets  map[Ident]*Let
	funs  map[Ident]*Fun
	start Ident
}

func initialBytes() ([]uint8, int) {
	return []uint8{220, 0, 0, 0, 0, 0, 0, 0, 0}, 1
}

func finalBytes() []uint8 {
	return []uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
}

func (c *Ctx) compile() []uint8 {
	bytes, saddr := initialBytes()
	c.size = uint64(len(bytes))

	c.start = makeFunIdent("__start", []parser.Typ{}, []parser.Typ{})
	start := c.funs[c.start]

	if start == nil {
		panic(fmt.Sprintln("missing __start(:) fun"))
	}

	sinfo := start.getInfo(c)
	if sinfo.inline {
		panic(fmt.Sprintln("__start(:) can't be an inline fun"))
	}
	if !sinfo.unsafe {
		panic(fmt.Sprintln("__start(:) needs to be unsafe"))
	}

	funs := []*Fun{}
	for _, fun := range c.funs {
		funs = append(funs, fun)
	}

	sort.Slice(funs, func(i, j int) bool {
		if funs[i].info == nil || funs[j].info == nil {
			return funs[i].info == nil
		}
		return funs[i].info.pos < funs[j].info.pos
	})

	for _, f := range funs {
		if f.info != nil {
			f.typeCheck(c, f.info.simpleTypeCheck)
		}
	}

	lets := []*Let{}
	for _, let := range c.lets {
		lets = append(lets, let)
	}

	for _, l := range lets {
		if l.info != nil {
			l.info.pos = c.getNextPos(l.info.size)
		}
	}

	putUvarint(bytes[saddr:saddr+8], sinfo.pos)

	for _, f := range funs {
		if f.info != nil && !f.info.inline {
			bytes = append(bytes, f.compile(c)...)
		}
	}

	sort.Slice(lets, func(i, j int) bool {
		if lets[i].info == nil || lets[j].info == nil {
			return lets[i].info == nil
		}
		return lets[i].info.pos < lets[j].info.pos
	})

	for _, l := range lets {
		if l.info != nil {
			bytes = append(bytes, l.staticCompile(c)...)
		}
	}

	bytes = append(bytes, []uint8(c.strs)...)
	return append(bytes, finalBytes()...)
}

func (l *Let) staticCompile(c *Ctx) []uint8 {
	if len(l.let.Exprs) != 1 {
		panic(fmt.Sprintf("let '%s' has to haves exactly one expr", l.let.Ident.Content))
	}
	expr := l.let.Exprs[0]
	var buf []uint8
	char := expr.AsChar()
	number := expr.AsNumber()
	if char != nil {
		if len(char.Content) != 1 {
			panic(fmt.Sprintf("the char '%s' can only contain one byte", char.Content))
		}
		buf = []uint8{char.Content[0]}
	} else if number != nil {
		num, err := strconv.ParseUint(number.Content, number.Base, number.Size*8)
		switch l.info.size {
		case 1:
			buf = []uint8{0}
		case 2:
			buf = []uint8{0, 0}
		case 4:
			buf = []uint8{0, 0, 0, 0}
		case 8:
			buf = []uint8{0, 0, 0, 0, 0, 0, 0, 0}
		default:
			panic("invalid size")
		}
		if err != nil {
			panic(fmt.Sprintf("unable to convert '%s' to a number", number.Content))
		}
		putUvarint(buf, num)
	} else {
		panic(fmt.Sprintf("let '%s' can only have a char or number expr", l.let.Ident.Content))
	}
	return buf
}

func (c *Ctx) getNextPos(size uint64) uint64 {
	res := c.size
	c.size += size
	return res
}

func (f *Fun) compile(c *Ctx) []uint8 {
	bytes := []uint8{}

	if f.info.asm {
		bytes = f.compileAsm()
		if !f.info.inline && !f.info.tailCall {
			bytes = append(bytes, 3)
		}
	} else {
		lets := []*Let{}
		for _, let := range f.info.lets {
			lets = append(lets, let)
		}
		sort.Slice(lets, func(i, j int) bool {
			return lets[i].info.pos < lets[j].info.pos
		})
		for _, l := range lets {
			bytes = append(bytes, f.compileExprs(c, l.let.Exprs, false)...)
			buf := []uint8{0, 0, 0, 0, 0, 0, 0, 0, 0}
			switch l.let.Typ.Size() {
			case 0:
				continue
			case 1:
				buf[0] = 235
			case 2:
				buf[0] = 236
			case 4:
				buf[0] = 237
			case 8:
				buf[0] = 238
			case 16:
				buf[0] = 239
			default:
				panic("invalid size")
			}
			putUvarint(buf[1:], f.info.pos+f.info.size-f.info.letSize+l.info.pos)
			bytes = append(bytes, buf...)

		}

		bytes = append(bytes, f.compileExprs(c, f.fun.Block.Exprs, true)...)
		if f.makeFunIdent() == c.start {
			bytes = append(bytes, 1)
		} else if !f.info.inline && !f.info.tailCall {
			bytes = append(bytes, 3)
		}
		bytes = append(bytes, make([]uint8, f.info.letSize, f.info.letSize)...)
	}
	return bytes
}

func (f *Fun) compileExprs(c *Ctx, exprs []parser.Expr, canTailCall bool) []uint8 {
	bytes := []uint8{}
	for i := 0; i < len(exprs); i++ {
		expr := exprs[i]
		ident := expr.AsIdent()
		if ident != nil {
			ident := Ident(ident.Content)
			let := f.info.lets[ident]
			var pos uint64
			if let == nil {
				let = c.lets[ident]
				pos = let.info.pos
			} else {
				pos = f.info.pos + f.info.size - f.info.letSize + let.info.pos
			}
			buf := []uint8{0, 0, 0, 0, 0, 0, 0, 0, 0}
			switch let.info.size {
			case 0:
				continue
			case 1:
				buf[0] = 230
			case 2:
				buf[0] = 231
			case 4:
				buf[0] = 232
			case 8:
				buf[0] = 233
			case 16:
				buf[0] = 234
			default:
				panic("invalid size")
			}
			putUvarint(buf[1:], pos)
			bytes = append(bytes, buf...)
		}
		call := expr.AsCall()
		if call != nil {
			ident := makeFunIdent(call.Ident.Content, call.Inputs, call.Outputs)
			fun := c.funs[ident]
			if fun.info.inline {
				bytes = append(bytes, fun.compile(c)...)
			} else {
				buf := []uint8{229, 0, 0, 0, 0, 0, 0, 0, 0}
				if canTailCall && f.info.tailCall {
					buf[0] = 220
				}
				putUvarint(buf[1:], fun.info.pos)
				bytes = append(bytes, buf...)

			}
		}
		number := expr.AsNumber()
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
		str := expr.AsString()
		if str != nil {
			ptr := c.getStr(str.Content)
			buf := []uint8{13, 0, 0, 0, 0, 0, 0, 0, 0, 13, 0, 0, 0, 0, 0, 0, 0, 0}
			putUvarint(buf[1:9], ptr)
			putUvarint(buf[10:], uint64(len(str.Content)))
			bytes = append(bytes, buf...)
		}
		char := expr.AsChar()
		if char != nil {
			if len(char.Content) != 1 {
				panic(fmt.Sprintf("the char '%s' can only contain one byte", char.Content))
			}
			bytes = append(bytes, 10, char.Content[0])
		}
		ifel := expr.AsIf()
		if ifel != nil {
			bytes = append(bytes, f.compileExprs(c, ifel.Con, false)...)

			buf := []uint8{226, 0, 0, 0, 0, 0, 0, 0, 0}
			putUvarint(buf[1:], f.sizeOfExprs(c, ifel.Else)+18)
			bytes = append(bytes, buf...)
			bytes = append(bytes, f.compileExprs(c, ifel.Else, false)...)

			buf = []uint8{221, 0, 0, 0, 0, 0, 0, 0, 0}
			putUvarint(buf[1:], f.sizeOfExprs(c, ifel.Exprs)+9)
			bytes = append(bytes, buf...)
			bytes = append(bytes, f.compileExprs(c, ifel.Exprs, false)...)
		}
		unwrap := expr.AsUnwrap()
		if unwrap != nil {
		}
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
