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

func (f *Fun) typeCheck(c *Ctx, simple bool) {
	if simple {
		f.typeCheckSimple(c)
	} else {
		stack := f.fun.Inputs
		if f.info.asm {
			stack = f.checkStackAsm(c, stack)
		} else {
			for _, let := range f.fun.Block.Lets {
				stackl := len(stack)
				stack = f.checkStackExprs(c, stack, let.Exprs)
				err, nstack := c.stackPrefix(stack, let.Typ)
				if err || stackl < len(nstack) {
					panic(fmt.Sprintf("the let in fun '%s' does not have a valid stack", f.makeFunIdent(c)))
				}
				stack = nstack
			}
			stack = f.checkStackExprs(c, stack, f.fun.Block.Exprs)
		}
		err, rest := c.stackPrefix(stack, f.fun.Outputs...)
		if err || len(rest) != 0 {
			panic(fmt.Sprintf("the fun '%s' does not have a valid stack", f.makeFunIdent(c)))
		}
	}
}

func (f *Fun) typeCheckSimple(c *Ctx) {
	stack := c.typsSize(f.fun.Inputs)

	if f.info.asm {
		stack = f.checkStackAsmSimple(c, stack)
	} else {
		for _, let := range f.fun.Block.Lets {
			stackl := stack
			stack = f.checkStackExprsSimple(c, stack, let.Exprs)
			err, nstack := stackPrefixSimple(c, stack, let.Typ)
			if err || stackl < nstack {
				panic(fmt.Sprintf("the let in fun '%s' does not have a valid stack", f.makeFunIdent(c)))
			}
			stack = nstack
		}
		stack = f.checkStackExprsSimple(c, stack, f.fun.Block.Exprs)
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

func (f *Fun) checkStackIfel(c *Ctx, stack []parser.Typ, ifel *parser.If) []parser.Typ {
	stack = f.checkStackExprs(c, stack, ifel.Con)
	err, stack := c.stackPrefix(stack, parser.U8)
	if err {
		panic(fmt.Sprintf("the if in '%s' does not have a valid condition stack", f.makeFunIdent(c)))
	}
	iStack := f.checkStackExprs(c, stack, ifel.Exprs)
	eStack := f.checkStackExprs(c, stack, ifel.Else)
	err, rstack := c.stackPrefix(iStack, eStack...)
	if err || len(rstack) != 0 {
		panic(fmt.Sprintf("the if in '%s' does not have a valid expression stack", f.makeFunIdent(c)))
	}
	return iStack
}

func (f *Fun) checkStackExprs(c *Ctx, stack []parser.Typ, exprs []parser.Expr) []parser.Typ {
	for i := 0; i < len(exprs); i++ {
		expr := exprs[i]
		ident := expr.AsIdent()
		call := expr.AsCall()
		number := expr.AsNumber()
		str := expr.AsString()
		char := expr.AsChar()
		ifel := expr.AsIf()
		unwrap := expr.AsUnwrap()
		wrap := expr.AsWrap()
		if ident != nil {
			ident := ident.Content
			let := f.info.lets[ident]
			if let == nil {
				let = c.lets[ident]
			}
			stack = append(stack, let.let.Typ)
		} else if call != nil {
			stack = f.checkStackCall(c, stack, call.Inputs, call.Outputs)
		} else if number != nil {
			stack = append(stack, number.Typ)
		} else if str != nil {
			stack = append(stack, parser.STRING)
		} else if char != nil {
			stack = append(stack, parser.CHAR)
		} else if ifel != nil {
			stack = f.checkStackIfel(c, stack, ifel)
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
		} else {
			panic("unreachable")
		}
	}
	return stack
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

func (f *Fun) checkStackIfelSimple(c *Ctx, stack int, ifel *parser.If) int {
	stack = f.checkStackExprsSimple(c, stack, ifel.Con)
	err, stack := stackPrefixSimple(c, stack, parser.U8)
	if err {
		panic(fmt.Sprintf("the if in '%s' does not have a valid condition stack", f.makeFunIdent(c)))
	}
	iStack := f.checkStackExprsSimple(c, stack, ifel.Exprs)
	eStack := f.checkStackExprsSimple(c, stack, ifel.Else)
	if err || iStack != eStack {
		panic(fmt.Sprintf("the if in '%s' does not have a valid expression stack", f.makeFunIdent(c)))
	}
	return iStack
}

func (f *Fun) checkStackExprsSimple(c *Ctx, stack int, exprs []parser.Expr) int {
	for i := 0; i < len(exprs); i++ {
		expr := exprs[i]
		ident := expr.AsIdent()
		call := expr.AsCall()
		number := expr.AsNumber()
		str := expr.AsString()
		char := expr.AsChar()
		ifel := expr.AsIf()
		unwrap := expr.AsUnwrap()
		wrap := expr.AsWrap()
		if ident != nil {
			ident := ident.Content
			let := f.info.lets[ident]
			if let == nil {
				let = c.lets[ident]
			}
			stack += let.let.Typ.Size(c.types)
		} else if call != nil {
			stack = f.checkStackCallSimple(c, stack, call.Inputs, call.Outputs)
		} else if number != nil {
			stack += number.Typ.Size(c.types)
		} else if str != nil {
			stack += parser.STRING.Size(c.types)
		} else if char != nil {
			stack += parser.CHAR.Size(c.types)
		} else if ifel != nil {
			stack = f.checkStackIfelSimple(c, stack, ifel)
		} else if unwrap != nil {
			panic(fmt.Sprintf("can't unwrap in simple type check fun '%s'", f.makeFunIdent(c)))
		} else if wrap != nil {
			panic(fmt.Sprintf("can't wrap in simple type check fun '%s'", f.makeFunIdent(c)))
		} else {
			panic("unreachable")
		}
	}
	return stack
}
