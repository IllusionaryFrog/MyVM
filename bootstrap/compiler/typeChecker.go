package compiler

import "fmt"

func (f *Fun) typeCheck(c *Ctx) {
	stackDiff := 0
	desiredSD := f.stackDiff()

	if f.info.asm {
		stackDiff += f.stackDiffAsm()
	} else {
		if len(f.fun.Block.Lets) != 0 {
			panic("unimplemented")
		}

		for i := 0; i < len(f.fun.Block.Exprs); i++ {
			expr := f.fun.Block.Exprs[i]
			ident := expr.GetIdent()
			if ident != nil {
				ident := Ident(ident.Content)
				let := c.lets[ident]
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
	}

	if stackDiff != desiredSD {
		panic(fmt.Sprintf("the fun '%s' does not have a valid stack", f.makeFunIdent()))
	}
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
