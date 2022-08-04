package compiler

import (
	"bootstrap/parser"
	"fmt"
)

func stackPrefix(stack []parser.Typ, typs ...parser.Typ) (bool, []parser.Typ) {
	typs_len := len(typs)
	stack_len := len(stack)
	if typs_len > stack_len {
		return true, stack
	}
	for i := 1; i <= typs_len; i++ {
		if stack[stack_len-i].String() != typs[typs_len-i].String() {
			return true, stack
		}
	}
	return false, stack[:stack_len-typs_len]
}

func stackPrefixSimple(stack int, typs ...parser.Typ) (bool, int) {
	typs_len := typsSize(typs)
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
			stack = f.checkStackAsm(stack)
		} else {
			for _, let := range f.fun.Block.Lets {
				stackl := len(stack)
				stack = f.checkStackExprs(c, stack, let.Exprs)
				err, nstack := stackPrefix(stack, let.Typ)
				if err || stackl < len(nstack) {
					panic(fmt.Sprintf("the let in fun '%s' does not have a valid stack", f.makeFunIdent()))
				}
				stack = nstack
			}
			stack = f.checkStackExprs(c, stack, f.fun.Block.Exprs)
		}
		err, rest := stackPrefix(stack, f.fun.Outputs...)
		if err || len(rest) != 0 {
			panic(fmt.Sprintf("the fun '%s' does not have a valid stack", f.makeFunIdent()))
		}
	}
}

func (f *Fun) typeCheckSimple(c *Ctx) {
	stack := typsSize(f.fun.Inputs)

	if f.info.asm {
		stack = f.checkStackAsmSimple(stack)
	} else {
		for _, let := range f.fun.Block.Lets {
			stackl := stack
			stack = f.checkStackExprsSimple(c, stack, let.Exprs)
			err, nstack := stackPrefixSimple(stack, let.Typ)
			if err || stackl < nstack {
				panic(fmt.Sprintf("the let in fun '%s' does not have a valid stack", f.makeFunIdent()))
			}
			stack = nstack
		}
		stack = f.checkStackExprsSimple(c, stack, f.fun.Block.Exprs)
	}
	err, stack := stackPrefixSimple(stack, f.fun.Outputs...)

	if err || stack != 0 {
		panic(fmt.Sprintf("the fun '%s' does not have a valid stack", f.makeFunIdent()))
	}
}

func (f *Fun) checkStackCall(stack []parser.Typ, inputs []parser.Typ, outputs []parser.Typ) []parser.Typ {
	err, stack := stackPrefix(stack, inputs...)
	if err {
		panic(fmt.Sprintf("the fun '%s' does not have a valid stack", f.makeFunIdent()))
	}
	return append(stack, outputs...)
}

func (f *Fun) checkStackIfel(c *Ctx, stack []parser.Typ, ifel *parser.If) []parser.Typ {
	stack = f.checkStackExprs(c, stack, ifel.Con)
	err, stack := stackPrefix(stack, parser.U8)
	if err {
		panic(fmt.Sprintf("the if in '%s' does not have a valid condition stack", f.makeFunIdent()))
	}
	iStack := f.checkStackExprs(c, stack, ifel.Exprs)
	eStack := f.checkStackExprs(c, stack, ifel.Else)
	err, rstack := stackPrefix(iStack, eStack...)
	if err || len(rstack) != 0 {
		panic(fmt.Sprintf("the if in '%s' does not have a valid expression stack", f.makeFunIdent()))
	}
	return iStack
}

func (f *Fun) checkStackExprs(c *Ctx, stack []parser.Typ, exprs []parser.Expr) []parser.Typ {
	for i := 0; i < len(exprs); i++ {
		expr := exprs[i]
		ident := expr.AsIdent()
		if ident != nil {
			ident := Ident(ident.Content)
			let := f.info.lets[ident]
			if let == nil {
				let = c.lets[ident]
			}
			stack = append(stack, let.let.Typ)
		}
		call := expr.AsCall()
		if call != nil {
			stack = f.checkStackCall(stack, call.Inputs, call.Outputs)
		}
		number := expr.AsNumber()
		if number != nil {
			stack = append(stack, number.Typ)
		}
		str := expr.AsString()
		if str != nil {
			stack = append(stack, parser.STRING)
		}
		char := expr.AsChar()
		if char != nil {
			stack = append(stack, parser.CHAR)
		}
		ifel := expr.AsIf()
		if ifel != nil {
			stack = f.checkStackIfel(c, stack, ifel)
		}
		unwrap := expr.AsUnwrap()
		if unwrap != nil {
			if len(stack) != 0 {
				last := len(stack) - 1
				err, sub := stack[last].Sub()
				if err {
					panic(fmt.Sprintf("can't unwrap stack at this position in '%s'", f.makeFunIdent()))
				}
				stack = append(stack[:last], sub...)
			} else {
				panic(fmt.Sprintf("can't unwrap empty stack in '%s'", f.makeFunIdent()))
			}
		}
	}
	return stack
}

func typsSize(typs []parser.Typ) int {
	size := 0
	for _, inp := range typs {
		size += inp.Size()
	}
	return size
}

func (f *Fun) checkStackCallSimple(stack int, inputs []parser.Typ, outputs []parser.Typ) int {
	err, stack := stackPrefixSimple(stack, inputs...)
	if err {
		panic(fmt.Sprintf("the fun '%s' does not have a valid stack", f.makeFunIdent()))
	}
	return stack + typsSize(outputs)
}

func (f *Fun) checkStackIfelSimple(c *Ctx, stack int, ifel *parser.If) int {
	stack = f.checkStackExprsSimple(c, stack, ifel.Con)
	err, stack := stackPrefixSimple(stack, parser.U8)
	if err {
		panic(fmt.Sprintf("the if in '%s' does not have a valid condition stack", f.makeFunIdent()))
	}
	iStack := f.checkStackExprsSimple(c, stack, ifel.Exprs)
	eStack := f.checkStackExprsSimple(c, stack, ifel.Else)
	if err || iStack != eStack {
		panic(fmt.Sprintf("the if in '%s' does not have a valid expression stack", f.makeFunIdent()))
	}
	return iStack
}

func (f *Fun) checkStackExprsSimple(c *Ctx, stack int, exprs []parser.Expr) int {
	for i := 0; i < len(exprs); i++ {
		expr := exprs[i]
		ident := expr.AsIdent()
		if ident != nil {
			ident := Ident(ident.Content)
			let := f.info.lets[ident]
			if let == nil {
				let = c.lets[ident]
			}
			stack += let.let.Typ.Size()
		}
		call := expr.AsCall()
		if call != nil {
			stack = f.checkStackCallSimple(stack, call.Inputs, call.Outputs)
		}
		number := expr.AsNumber()
		if number != nil {
			stack += number.Typ.Size()
		}
		str := expr.AsString()
		if str != nil {
			stack += parser.STRING.Size()
		}
		char := expr.AsChar()
		if char != nil {
			stack += parser.CHAR.Size()
		}
		ifel := expr.AsIf()
		if ifel != nil {
			stack = f.checkStackIfelSimple(c, stack, ifel)
		}
		unwrap := expr.AsUnwrap()
		if unwrap != nil {
			panic(fmt.Sprintf("can't unwrap in simple type check fun '%s'", f.makeFunIdent()))
		}
	}
	return stack
}
