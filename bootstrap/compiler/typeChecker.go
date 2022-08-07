package compiler

import (
	"bootstrap/parser"
	"fmt"
)

func (c *Ctx) stackPrefix(stack []parser.Typ, typs ...parser.Typ) (bool, []parser.Typ) {
	typs_len := len(typs)
	stack_len := len(stack)
	if typs_len > stack_len {
		return true, stack
	}
	for i := 1; i <= typs_len; i++ {
		if stack[stack_len-i].String(c.types) != typs[typs_len-i].String(c.types) {
			return true, stack
		}
	}
	return false, stack[:stack_len-typs_len]
}

func stackPrefixSimple(c *Ctx, stack int, typs ...parser.Typ) (bool, int) {
	typs_len := c.typsSize(typs)
	if typs_len > stack {
		return true, stack
	}
	return false, stack - typs_len
}

func containsNever(ts []parser.Typ) bool {
	for _, t := range ts {
		if t.IsNever() {
			return true
		}
	}
	return false
}

func (f *Fun) typeCheck(c *Ctx, simple bool) {
	if containsNever(f.fun.Inputs) {
		panic(fmt.Sprintf("the never type can't be used as an input argument in '%s'", f.makeFunIdent(c)))
	}
	if containsNever(f.fun.Outputs) && len(f.fun.Outputs) != 1 {
		panic(fmt.Sprintf("the never type has to be the only output of '%s'", f.makeFunIdent(c)))
	}

	var ret bool
	var never bool

	if simple {
		f.typeCheckSimple(c)
	} else {
		stack := append([]parser.Typ{}, f.fun.Inputs...)
		if f.info.asm {
			stack = f.checkStackAsm(c, stack)
		} else {
			for _, let := range f.fun.Block.Lets {
				stackl := len(stack)
				never, ret, stack = f.checkStackExprs(c, stack, let.Exprs)
				if ret {
					panic(fmt.Sprintf("the let in fun '%s' does not have a valid stack", f.makeFunIdent(c)))
				}
				if never {
					return
				}
				err, nstack := c.stackPrefix(stack, let.Typ)
				if err || stackl < len(nstack) {
					panic(fmt.Sprintf("the let in fun '%s' does not have a valid stack", f.makeFunIdent(c)))
				}
				stack = nstack
			}
			never, _, stack = f.checkStackExprs(c, stack, f.fun.Block.Exprs)
			if never {
				return
			}
		}
		err, rest := c.stackPrefix(stack, f.fun.Outputs...)
		if err || len(rest) != 0 {
			panic(fmt.Sprintf("the fun '%s' does not have a valid stack", f.makeFunIdent(c)))
		}
	}
}

func (f *Fun) typeCheckSimple(c *Ctx) {
	var ret bool
	var never bool
	stack := c.typsSize(f.fun.Inputs)

	if f.info.asm {
		stack = f.checkStackAsmSimple(c, stack)
	} else {
		for _, let := range f.fun.Block.Lets {
			stackl := stack
			never, ret, stack = f.checkStackExprsSimple(c, stack, let.Exprs)
			if ret {
				panic(fmt.Sprintf("the let in fun '%s' does not have a valid stack", f.makeFunIdent(c)))
			}
			if never {
				return
			}
			err, nstack := stackPrefixSimple(c, stack, let.Typ)
			if err || stackl < nstack {
				panic(fmt.Sprintf("the let in fun '%s' does not have a valid stack", f.makeFunIdent(c)))
			}
			stack = nstack
		}
		never, _, stack = f.checkStackExprsSimple(c, stack, f.fun.Block.Exprs)
		if never {
			return
		}
	}
	err, stack := stackPrefixSimple(c, stack, f.fun.Outputs...)

	if err || stack != 0 {
		panic(fmt.Sprintf("the fun '%s' does not have a valid stack", f.makeFunIdent(c)))
	}
}

func (f *Fun) checkStackCall(c *Ctx, stack []parser.Typ, inputs []parser.Typ, outputs []parser.Typ) []parser.Typ {
	err, stack := c.stackPrefix(stack, inputs...)
	if err {
		panic(fmt.Sprintf("the fun '%s' does not have a valid stack", f.makeFunIdent(c)))
	}
	return append(stack, outputs...)
}

func (f *Fun) checkStackIfel(c *Ctx, stack []parser.Typ, ifel *parser.If) (bool, bool, []parser.Typ) {
	never, ret, stack := f.checkStackExprs(c, stack, ifel.Con)
	if ret {
		panic(fmt.Sprintf("the if in '%s' does not have a valid condition stack", f.makeFunIdent(c)))
	}
	if never {
		return true, false, []parser.Typ{}
	}
	err, stack := c.stackPrefix(stack, parser.BOOL)
	if err {
		panic(fmt.Sprintf("the if in '%s' does not have a valid condition stack", f.makeFunIdent(c)))
	}
	iNever, ret, iStack := f.checkStackExprs(c, stack, ifel.Exprs)
	if ret {
		return iNever, true, iStack
	}
	eNever, ret, eStack := f.checkStackExprs(c, stack, ifel.Else)
	if ret {
		return eNever, true, eStack
	}
	if iNever {
		if eNever {
			return true, false, []parser.Typ{}
		}
		return false, false, eStack
	}
	if eNever {
		return false, false, iStack
	}
	err, rstack := c.stackPrefix(iStack, eStack...)
	if err || len(rstack) != 0 {
		panic(fmt.Sprintf("the if in '%s' does not have a valid expression stack", f.makeFunIdent(c)))
	}
	return false, false, iStack
}

func (f *Fun) checkStackExprs(c *Ctx, stack []parser.Typ, exprs []parser.Expr) (bool, bool, []parser.Typ) {
	var r bool
	var never bool
	for i := 0; i < len(exprs); i++ {
		expr := exprs[i]
		ident := expr.AsIdent()
		call := expr.AsCall()
		number := expr.AsNumber()
		str := expr.AsString()
		ifel := expr.AsIf()
		unwrap := expr.AsUnwrap()
		wrap := expr.AsWrap()
		addr := expr.AsAddr()
		ret := expr.AsReturn()
		if ident != nil {
			ident := ident.Content
			let := f.info.lets[ident]
			if let == nil {
				let = c.lets[ident]
			}
			stack = append(stack, let.let.Typ)
		} else if call != nil {
			if containsNever(call.Outputs) {
				return true, false, []parser.Typ{}
			}
			stack = f.checkStackCall(c, stack, call.Inputs, call.Outputs)
		} else if number != nil {
			stack = append(stack, number.Typ)
		} else if str != nil {
			stack = append(stack, parser.STRING)
		} else if ifel != nil {
			never, r, stack = f.checkStackIfel(c, stack, ifel)
			if r {
				return never, true, stack
			}
			if never {
				return true, false, []parser.Typ{}
			}
		} else if unwrap != nil {
			if len(stack) != 0 {
				last := len(stack) - 1
				sub := stack[last].Sub(c.types)
				stack = append(stack[:last], sub...)
			} else {
				panic(fmt.Sprintf("can't unwrap empty stack in '%s'", f.makeFunIdent(c)))
			}
		} else if wrap != nil {
			err, nstack := c.stackPrefix(stack, wrap.Typ.Sub(c.types)...)
			if err {
				panic(fmt.Sprintf("can't wrap stack in '%s'", f.makeFunIdent(c)))
			}
			stack = append(nstack, wrap.Typ)
		} else if addr != nil {
			stack = append(stack, parser.U64)
		} else if ret != nil {
			return false, true, stack
		} else {
			panic("unreachable")
		}
	}
	return false, false, stack
}

func (c *Ctx) typsSize(typs []parser.Typ) int {
	size := 0
	for _, inp := range typs {
		size += inp.Size(c.types)
	}
	return size
}

func (f *Fun) checkStackCallSimple(c *Ctx, stack int, inputs []parser.Typ, outputs []parser.Typ) int {
	err, stack := stackPrefixSimple(c, stack, inputs...)
	if err {
		panic(fmt.Sprintf("the fun '%s' does not have a valid stack", f.makeFunIdent(c)))
	}
	return stack + c.typsSize(outputs)
}

func (f *Fun) checkStackIfelSimple(c *Ctx, stack int, ifel *parser.If) (bool, bool, int) {
	never, ret, stack := f.checkStackExprsSimple(c, stack, ifel.Con)
	if ret {
		panic(fmt.Sprintf("the if in '%s' does not have a valid condition stack", f.makeFunIdent(c)))
	}
	if never {
		return true, false, 0
	}
	err, stack := stackPrefixSimple(c, stack, parser.BOOL)
	if err {
		panic(fmt.Sprintf("the if in '%s' does not have a valid condition stack", f.makeFunIdent(c)))
	}
	iNever, ret, iStack := f.checkStackExprsSimple(c, stack, ifel.Exprs)
	if ret {
		return iNever, true, iStack
	}
	eNever, ret, eStack := f.checkStackExprsSimple(c, stack, ifel.Else)
	if ret {
		return eNever, true, eStack
	}
	if iNever {
		if eNever {
			return true, false, 0
		}
		return false, false, eStack
	}
	if eNever {
		return false, false, iStack
	}
	if iStack != eStack {
		panic(fmt.Sprintf("the if in '%s' does not have a valid expression stack", f.makeFunIdent(c)))
	}
	return false, false, iStack
}

func (f *Fun) checkStackExprsSimple(c *Ctx, stack int, exprs []parser.Expr) (bool, bool, int) {
	var r bool
	var never bool
	for i := 0; i < len(exprs); i++ {
		expr := exprs[i]
		ident := expr.AsIdent()
		call := expr.AsCall()
		number := expr.AsNumber()
		str := expr.AsString()
		ifel := expr.AsIf()
		unwrap := expr.AsUnwrap()
		wrap := expr.AsWrap()
		addr := expr.AsAddr()
		ret := expr.AsReturn()
		if ident != nil {
			ident := ident.Content
			let := f.info.lets[ident]
			if let == nil {
				let = c.lets[ident]
			}
			stack += let.let.Typ.Size(c.types)
		} else if call != nil {
			if containsNever(call.Outputs) {
				return true, false, 0
			}
			stack = f.checkStackCallSimple(c, stack, call.Inputs, call.Outputs)
		} else if number != nil {
			stack += number.Typ.Size(c.types)
		} else if str != nil {
			stack += parser.STRING.Size(c.types)
		} else if ifel != nil {
			never, r, stack = f.checkStackIfelSimple(c, stack, ifel)
			if r {
				return never, true, stack
			}
			if never {
				return true, false, 0
			}
		} else if unwrap != nil {
			panic(fmt.Sprintf("can't unwrap in simple type check fun '%s'", f.makeFunIdent(c)))
		} else if wrap != nil {
			panic(fmt.Sprintf("can't wrap in simple type check fun '%s'", f.makeFunIdent(c)))
		} else if addr != nil {
			stack += 8
		} else if ret != nil {
			return false, true, stack
		} else {
			panic("unreachable")
		}
	}
	return false, false, stack
}
