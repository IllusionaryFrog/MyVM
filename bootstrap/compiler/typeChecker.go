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
		ident := expr.AsIdent()
		if ident != nil {
			ident := Ident(ident.Content)
			let := f.info.lets[ident]
			if let == nil {
				let = c.lets[ident]
			}
			stackDiff += int(let.info.size)
		}
		call := expr.AsCall()
		if call != nil {
			ident := makeFunIdent(call.Ident.Content, call.Inputs, call.Outputs)
			fun := c.funs[ident]
			fun.typeCheck(c)
			stackDiff += fun.stackDiff()
		}
		number := expr.AsNumber()
		if number != nil {
			stackDiff += number.Size
		}
		str := expr.AsString()
		if str != nil {
			stackDiff += 8 + 8
		}
		char := expr.AsChar()
		if char != nil {
			stackDiff += 1
		}
		ifel := expr.AsIf()
		if ifel != nil {
			stackDiff += f.stackDiffExprs(c, ifel.Con) - 1
			diff := f.stackDiffExprs(c, ifel.Exprs)
			if diff != f.stackDiffExprs(c, ifel.Else) {
				panic(fmt.Sprintf("invalid exprs in if in '%s'", f.fun.Ident.Content))
			}
			stackDiff += diff
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
