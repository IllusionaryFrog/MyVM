package compiler

import (
	"bootstrap/parser"
	"fmt"
)

func (f *Fun) typeCheck(c *Ctx) {
	stackDiff := 0
	desiredSD := f.stackDiff()

	if f.info.asm {
		stackDiff += f.stackDiffAsm()
	} else {
		for _, let := range f.fun.Block.Lets {
			diff := f.stackDiffExprs(c, let.Exprs) - let.Typ.Size()
			if diff > 0 {
				panic(fmt.Sprintf("the let '%s' does not have a valid stack", let.Ident.Content))
			}
			stackDiff += diff
		}

		stackDiff += f.stackDiffExprs(c, f.fun.Block.Exprs)
	}

	if stackDiff != desiredSD {
		panic(fmt.Sprintf("the fun '%s' does not have a valid stack", f.makeFunIdent()))
	}
}

func (f *Fun) stackDiffExprs(c *Ctx, exprs []parser.Expr) int {
	stackDiff := 0
	for i := 0; i < len(exprs); i++ {
		expr := exprs[i]
		ident := expr.GetIdent()
		if ident != nil {
			ident := Ident(ident.Content)
			let := f.info.lets[ident]
			if let == nil {
				let = c.lets[ident]
			}
			stackDiff += int(let.info.size)
		}
		call := expr.GetCall()
		if call != nil {
			ident := makeFunIdent(call.Ident.Content, call.Inputs, call.Outputs)
			fun := c.funs[ident]
			fun.typeCheck(c)
			stackDiff += fun.stackDiff()
		}
		number := expr.GetNumber()
		if number != nil {
			stackDiff += number.Size
		}
		str := expr.GetString()
		if str != nil {
			stackDiff += 8 + 8
		}
		char := expr.GetChar()
		if char != nil {
			stackDiff += 1
		}
	}
	return stackDiff
}

func (f *Fun) stackDiff() int {
	inps := 0
	for i := 0; i < len(f.fun.Inputs); i++ {
		inps += f.fun.Inputs[i].Size()
	}
	outs := 0
	for i := 0; i < len(f.fun.Outputs); i++ {
		outs += f.fun.Outputs[i].Size()
	}
	return outs - inps
}
