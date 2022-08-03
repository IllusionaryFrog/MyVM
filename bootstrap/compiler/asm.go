package compiler

import (
	"bootstrap/parser"
	"fmt"
)

func (f *Fun) stackDiffAsm() int {
	stackDiff := 0

	for i := 0; i < len(f.fun.Block.Exprs); i++ {
		str := f.fun.Block.Exprs[i].AsString()
		if str == nil {
			panic(fmt.Sprintf("the asm fun '%s' can only contain strings", f.makeFunIdent()))
		}
		stackDiff += stackDiffInst(str.Content)
	}

	return stackDiff
}

func (f *Fun) compileAsm() []uint8 {
	bytes := []uint8{}

	for i := 0; i < len(f.fun.Block.Exprs); i++ {
		str := f.fun.Block.Exprs[i].AsString()
		if str == nil {
			panic(fmt.Sprintf("the asm fun '%s' can only contain strings", f.makeFunIdent()))
		}
		bytes = append(bytes, parseInst(str.Content))
	}

	return bytes
}

func parseInst(inst string) uint8 {
	switch inst {
	case "nop":
		return 0
	case "halt":
		return 1
	case "call":
		return 2
	case "return":
		return 3
	case "inter":
		return 4
	case "alloc":
		return 5
	case "read":
		return 6
	case "write":
		return 7
	case "readFile":
		return 8
	case "writeFile":
		return 9
	case "pushImm8":
		return 10
	case "pushImm16":
		return 11
	case "pushImm32":
		return 12
	case "pushImm64":
		return 13
	case "pushImm128":
		return 14
	case "popSp":
		return 15
	case "popCs":
		return 16
	case "popIh":
		return 17
	case "popIr":
		return 18
	case "pushIr":
		return 19
	case "drop8":
		return 20
	case "drop16":
		return 21
	case "drop32":
		return 22
	case "drop64":
		return 23
	case "drop128":
		return 24
	case "negate8":
		return 25
	case "negate16":
		return 26
	case "negate32":
		return 27
	case "negate64":
		return 28
	case "negate128":
		return 29
	case "swap8":
		return 30
	case "swap16":
		return 31
	case "swap32":
		return 32
	case "swap64":
		return 33
	case "swap128":
		return 34
	case "rotate8":
		return 35
	case "rotate16":
		return 36
	case "rotate32":
		return 37
	case "rotate64":
		return 38
	case "rotate128":
		return 39
	case "dup8":
		return 40
	case "dup16":
		return 41
	case "dup32":
		return 42
	case "dup64":
		return 43
	case "dup128":
		return 44
	case "over8":
		return 45
	case "over16":
		return 46
	case "over32":
		return 47
	case "over64":
		return 48
	case "over128":
		return 49
	case "and8":
		return 50
	case "and16":
		return 51
	case "and32":
		return 52
	case "and64":
		return 53
	case "and128":
		return 54
	case "or8":
		return 55
	case "or16":
		return 56
	case "or32":
		return 57
	case "or64":
		return 58
	case "or128":
		return 59
	case "shiftL8":
		return 60
	case "shiftL16":
		return 61
	case "shiftL32":
		return 62
	case "shiftL64":
		return 63
	case "shiftL128":
		return 64
	case "shiftR8":
		return 65
	case "shiftR16":
		return 66
	case "shiftR32":
		return 67
	case "shiftR64":
		return 68
	case "shiftR128":
		return 69
	case "rotateL8":
		return 70
	case "rotateL16":
		return 71
	case "rotateL32":
		return 72
	case "rotateL64":
		return 73
	case "rotateL128":
		return 74
	case "rotateR8":
		return 75
	case "rotateR16":
		return 76
	case "rotateR32":
		return 77
	case "rotateR64":
		return 78
	case "rotateR128":
		return 79
	case "equal8":
		return 80
	case "equal16":
		return 81
	case "equal32":
		return 82
	case "equal64":
		return 83
	case "equal128":
		return 84
	case "notEq8":
		return 85
	case "notEq16":
		return 86
	case "notEq32":
		return 87
	case "notEq64":
		return 88
	case "notEq128":
		return 89
	case "jump":
		return 90
	case "jumpF":
		return 91
	case "jumpB":
		return 92
	case "sleep":
		return 94
	case "branch":
		return 95
	case "branchF":
		return 96
	case "branchB":
		return 97
	case "addU8":
		return 100
	case "addU16":
		return 101
	case "addU32":
		return 102
	case "addU64":
		return 103
	case "addU128":
		return 104
	case "addI8":
		return 105
	case "addI16":
		return 106
	case "addI32":
		return 107
	case "addI64":
		return 108
	case "addI128":
		return 109
	case "subU8":
		return 110
	case "subU16":
		return 111
	case "subU32":
		return 112
	case "subU64":
		return 113
	case "subU128":
		return 114
	case "subI8":
		return 115
	case "subI16":
		return 116
	case "subI32":
		return 117
	case "subI64":
		return 118
	case "subI128":
		return 119
	case "mulU8":
		return 120
	case "mulU16":
		return 121
	case "mulU32":
		return 122
	case "mulU64":
		return 123
	case "mulU128":
		return 124
	case "mulI8":
		return 125
	case "mulI16":
		return 126
	case "mulI32":
		return 127
	case "mulI64":
		return 128
	case "mulI128":
		return 129
	case "divU8":
		return 130
	case "divU16":
		return 131
	case "divU32":
		return 132
	case "divU64":
		return 133
	case "divU128":
		return 134
	case "divI8":
		return 135
	case "divI16":
		return 136
	case "divI32":
		return 137
	case "divI64":
		return 138
	case "divI128":
		return 139
	case "modU8":
		return 140
	case "modU16":
		return 141
	case "modU32":
		return 142
	case "modU64":
		return 143
	case "modU128":
		return 144
	case "modI8":
		return 145
	case "modI16":
		return 146
	case "modI32":
		return 147
	case "modI64":
		return 148
	case "modI128":
		return 149
	case "lessU8":
		return 150
	case "lessU16":
		return 151
	case "lessU32":
		return 152
	case "lessU64":
		return 153
	case "lessU128":
		return 154
	case "lessI8":
		return 155
	case "lessI16":
		return 156
	case "lessI32":
		return 157
	case "lessI64":
		return 158
	case "lessI128":
		return 159
	case "lessEqU8":
		return 160
	case "lessEqU16":
		return 161
	case "lessEqU32":
		return 162
	case "lessEqU64":
		return 163
	case "lessEqU128":
		return 164
	case "lessEqI8":
		return 165
	case "lessEqI16":
		return 166
	case "lessEqI32":
		return 167
	case "lessEqI64":
		return 168
	case "lessEqI128":
		return 169
	case "greatU8":
		return 170
	case "greatU16":
		return 171
	case "greatU32":
		return 172
	case "greatU64":
		return 173
	case "greatU128":
		return 174
	case "greatI8":
		return 175
	case "greatI16":
		return 176
	case "greatI32":
		return 177
	case "greatI64":
		return 178
	case "greatI128":
		return 179
	case "greatEqU8":
		return 180
	case "greatEqU16":
		return 181
	case "greatEqU32":
		return 182
	case "greatEqU64":
		return 183
	case "greatEqU128":
		return 184
	case "greatEqI8":
		return 185
	case "greatEqI16":
		return 186
	case "greatEqI32":
		return 187
	case "greatEqI64":
		return 188
	case "greatEqI128":
		return 189
	case "8to16":
		return 190
	case "8to32":
		return 191
	case "8to64":
		return 192
	case "8to128":
		return 193
	case "16to8":
		return 194
	case "16to32":
		return 195
	case "16to64":
		return 196
	case "16to128":
		return 197
	case "32to8":
		return 198
	case "32to16":
		return 199
	case "32to64":
		return 200
	case "32to128":
		return 201
	case "64to8":
		return 202
	case "64to16":
		return 203
	case "64to32":
		return 204
	case "64to128":
		return 205
	case "128to8":
		return 206
	case "128to16":
		return 207
	case "128to32":
		return 208
	case "128to64":
		return 209
	case "load8":
		return 210
	case "load16":
		return 211
	case "load32":
		return 212
	case "load64":
		return 213
	case "load128":
		return 214
	case "store8":
		return 215
	case "store16":
		return 216
	case "store32":
		return 217
	case "store64":
		return 218
	case "store128":
		return 219
	case "jumpImm":
		return 220
	case "jumpImmF":
		return 221
	case "jumpImmB":
		return 222
	case "sleepImm":
		return 224
	case "branchImm":
		return 225
	case "branchImmF":
		return 226
	case "branchImmB":
		return 227
	case "callImm":
		return 229
	case "load8Imm":
		return 230
	case "load16Imm":
		return 231
	case "load32Imm":
		return 232
	case "load64Imm":
		return 233
	case "load128Imm":
		return 234
	case "store8Imm":
		return 235
	case "store16Imm":
		return 236
	case "store32Imm":
		return 237
	case "store64Imm":
		return 238
	case "store128Imm":
		return 239
	case "debug":
		return 250
	case "debug8":
		return 251
	case "debug16":
		return 252
	case "debug32":
		return 253
	case "debug64":
		return 254
	case "debug128":
		return 255
	default:
		panic(fmt.Sprintf("invalid asm instruction '%s'", inst))
	}
}

func stackDiffInst(inst string) int {
	inputs, outputs := argsInst(inst)
	inps := 0
	for i := 0; i < len(inputs.typs); i++ {
		inps += inputs.typs[i].Size()
	}
	outs := 0
	for i := 0; i < len(outputs.typs); i++ {
		outs += outputs.typs[i].Size()
	}
	return outs - inps
}

type Args struct {
	typs []parser.Builtin
}

func args(args ...parser.Builtin) Args {
	return Args{typs: []parser.Builtin(args)}
}

func argsInst(inst string) (Args, Args) {
	switch inst {
	case "nop":
		return args(), args()
	case "halt":
		return args(), args()
	case "call":
		return args(parser.U64), args()
	case "return":
		return args(), args()
	case "inter":
		return args(), args()
	case "alloc":
		return args(parser.U64), args(parser.U64)
	case "read":
		return args(parser.U64, parser.U64), args(parser.U64)
	case "write":
		return args(parser.U64, parser.U64), args(parser.U64)
	case "readFile":
		return args(parser.U64, parser.U64, parser.U64, parser.U64), args(parser.U64)
	case "writeFile":
		return args(parser.U64, parser.U64, parser.U64, parser.U64), args(parser.U64)
	case "pushImm8":
		return args(), args(parser.U8)
	case "pushImm16":
		return args(), args(parser.U16)
	case "pushImm32":
		return args(), args(parser.U32)
	case "pushImm64":
		return args(), args(parser.U64)
	case "pushImm128":
		return args(), args(parser.U128)
	case "popSp":
		return args(parser.U64), args()
	case "popCs":
		return args(parser.U64), args()
	case "popIh":
		return args(parser.U64), args()
	case "popIr":
		return args(parser.I8), args()
	case "pushIr":
		return args(), args(parser.I8)
	case "drop8":
		return args(parser.U8), args()
	case "drop16":
		return args(parser.U16), args()
	case "drop32":
		return args(parser.U32), args()
	case "drop64":
		return args(parser.U64), args()
	case "drop128":
		return args(parser.U128), args()
	case "negate8":
		return args(parser.U8), args(parser.U8)
	case "negate16":
		return args(parser.U16), args(parser.U16)
	case "negate32":
		return args(parser.U32), args(parser.U32)
	case "negate64":
		return args(parser.U64), args(parser.U64)
	case "negate128":
		return args(parser.U128), args(parser.U128)
	case "swap8":
		return args(), args()
	case "swap16":
		return args(), args()
	case "swap32":
		return args(), args()
	case "swap64":
		return args(), args()
	case "swap128":
		return args(), args()
	case "rotate8":
		return args(), args()
	case "rotate16":
		return args(), args()
	case "rotate32":
		return args(), args()
	case "rotate64":
		return args(), args()
	case "rotate128":
		return args(), args()
	case "dup8":
		return args(), args(parser.U8)
	case "dup16":
		return args(), args(parser.U16)
	case "dup32":
		return args(), args(parser.U32)
	case "dup64":
		return args(), args(parser.U64)
	case "dup128":
		return args(), args(parser.U128)
	case "over8":
		return args(), args(parser.U8)
	case "over16":
		return args(), args(parser.U16)
	case "over32":
		return args(), args(parser.U32)
	case "over64":
		return args(), args(parser.U64)
	case "over128":
		return args(), args(parser.U128)
	case "and8":
		return args(parser.U8, parser.U8), args(parser.U8)
	case "and16":
		return args(parser.U16, parser.U16), args(parser.U16)
	case "and32":
		return args(parser.U32, parser.U32), args(parser.U32)
	case "and64":
		return args(parser.U64, parser.U64), args(parser.U64)
	case "and128":
		return args(parser.U128, parser.U128), args(parser.U128)
	case "or8":
		return args(parser.U8, parser.U8), args(parser.U8)
	case "or16":
		return args(parser.U16, parser.U16), args(parser.U16)
	case "or32":
		return args(parser.U32, parser.U32), args(parser.U32)
	case "or64":
		return args(parser.U64, parser.U64), args(parser.U64)
	case "or128":
		return args(parser.U128, parser.U128), args(parser.U128)
	case "shiftL8":
		return args(parser.U8, parser.U8), args(parser.U8)
	case "shiftL16":
		return args(parser.U16, parser.U8), args(parser.U16)
	case "shiftL32":
		return args(parser.U32, parser.U8), args(parser.U32)
	case "shiftL64":
		return args(parser.U64, parser.U8), args(parser.U64)
	case "shiftL128":
		return args(parser.U128, parser.U8), args(parser.U128)
	case "shiftR8":
		return args(parser.U8, parser.U8), args(parser.U8)
	case "shiftR16":
		return args(parser.U16, parser.U8), args(parser.U16)
	case "shiftR32":
		return args(parser.U32, parser.U8), args(parser.U32)
	case "shiftR64":
		return args(parser.U64, parser.U8), args(parser.U64)
	case "shiftR128":
		return args(parser.U128, parser.U8), args(parser.U128)
	case "rotateL8":
		return args(parser.U8, parser.U8), args(parser.U8)
	case "rotateL16":
		return args(parser.U16, parser.U8), args(parser.U16)
	case "rotateL32":
		return args(parser.U32, parser.U8), args(parser.U32)
	case "rotateL64":
		return args(parser.U64, parser.U8), args(parser.U64)
	case "rotateL128":
		return args(parser.U128, parser.U8), args(parser.U128)
	case "rotateR8":
		return args(parser.U8, parser.U8), args(parser.U8)
	case "rotateR16":
		return args(parser.U16, parser.U8), args(parser.U16)
	case "rotateR32":
		return args(parser.U32, parser.U8), args(parser.U32)
	case "rotateR64":
		return args(parser.U64, parser.U8), args(parser.U64)
	case "rotateR128":
		return args(parser.U128, parser.U8), args(parser.U128)
	case "equal8":
		return args(parser.U8, parser.U8), args(parser.U8)
	case "equal16":
		return args(parser.U16, parser.U16), args(parser.U8)
	case "equal32":
		return args(parser.U32, parser.U32), args(parser.U8)
	case "equal64":
		return args(parser.U64, parser.U64), args(parser.U8)
	case "equal128":
		return args(parser.U128, parser.U128), args(parser.U8)
	case "notEq8":
		return args(parser.U8, parser.U8), args(parser.U8)
	case "notEq16":
		return args(parser.U16, parser.U16), args(parser.U8)
	case "notEq32":
		return args(parser.U32, parser.U32), args(parser.U8)
	case "notEq64":
		return args(parser.U64, parser.U64), args(parser.U8)
	case "notEq128":
		return args(parser.U128, parser.U128), args(parser.U8)
	case "jump":
		return args(parser.U64), args()
	case "jumpF":
		return args(parser.U64), args()
	case "jumpB":
		return args(parser.U64), args()
	case "sleep":
		return args(parser.U64), args()
	case "branch":
		return args(parser.U64, parser.U8), args()
	case "branchF":
		return args(parser.U64, parser.U8), args()
	case "branchB":
		return args(parser.U64, parser.U8), args()
	case "addU8":
		return args(parser.U8, parser.U8), args(parser.U8)
	case "addU16":
		return args(parser.U16, parser.U16), args(parser.U16)
	case "addU32":
		return args(parser.U32, parser.U32), args(parser.U32)
	case "addU64":
		return args(parser.U64, parser.U64), args(parser.U64)
	case "addU128":
		return args(parser.U128, parser.U128), args(parser.U128)
	case "addI8":
		return args(parser.I8, parser.I8), args(parser.I8)
	case "addI16":
		return args(parser.I16, parser.I16), args(parser.I16)
	case "addI32":
		return args(parser.I32, parser.I32), args(parser.I32)
	case "addI64":
		return args(parser.I64, parser.I64), args(parser.I64)
	case "addI128":
		return args(parser.I128, parser.I128), args(parser.I128)
	case "subU8":
		return args(parser.U8, parser.U8), args(parser.U8)
	case "subU16":
		return args(parser.U16, parser.U16), args(parser.U16)
	case "subU32":
		return args(parser.U32, parser.U32), args(parser.U32)
	case "subU64":
		return args(parser.U64, parser.U64), args(parser.U64)
	case "subU128":
		return args(parser.U128, parser.U128), args(parser.U128)
	case "subI8":
		return args(parser.I8, parser.I8), args(parser.I8)
	case "subI16":
		return args(parser.I16, parser.I16), args(parser.I16)
	case "subI32":
		return args(parser.I32, parser.I32), args(parser.I32)
	case "subI64":
		return args(parser.I64, parser.I64), args(parser.I64)
	case "subI128":
		return args(parser.I128, parser.I128), args(parser.I128)
	case "mulU8":
		return args(parser.U8, parser.U8), args(parser.U8)
	case "mulU16":
		return args(parser.U16, parser.U16), args(parser.U16)
	case "mulU32":
		return args(parser.U32, parser.U32), args(parser.U32)
	case "mulU64":
		return args(parser.U64, parser.U64), args(parser.U64)
	case "mulU128":
		return args(parser.U128, parser.U128), args(parser.U128)
	case "mulI8":
		return args(parser.I8, parser.I8), args(parser.I8)
	case "mulI16":
		return args(parser.I16, parser.I16), args(parser.I16)
	case "mulI32":
		return args(parser.I32, parser.I32), args(parser.I32)
	case "mulI64":
		return args(parser.I64, parser.I64), args(parser.I64)
	case "mulI128":
		return args(parser.I128, parser.I128), args(parser.I128)
	case "divU8":
		return args(parser.U8, parser.U8), args(parser.U8)
	case "divU16":
		return args(parser.U16, parser.U16), args(parser.U16)
	case "divU32":
		return args(parser.U32, parser.U32), args(parser.U32)
	case "divU64":
		return args(parser.U64, parser.U64), args(parser.U64)
	case "divU128":
		return args(parser.U128, parser.U128), args(parser.U128)
	case "divI8":
		return args(parser.I8, parser.I8), args(parser.I8)
	case "divI16":
		return args(parser.I16, parser.I16), args(parser.I16)
	case "divI32":
		return args(parser.I32, parser.I32), args(parser.I32)
	case "divI64":
		return args(parser.I64, parser.I64), args(parser.I64)
	case "divI128":
		return args(parser.I128, parser.I128), args(parser.I128)
	case "modU8":
		return args(parser.U8, parser.U8), args(parser.U8)
	case "modU16":
		return args(parser.U16, parser.U16), args(parser.U16)
	case "modU32":
		return args(parser.U32, parser.U32), args(parser.U32)
	case "modU64":
		return args(parser.U64, parser.U64), args(parser.U64)
	case "modU128":
		return args(parser.U128, parser.U128), args(parser.U128)
	case "modI8":
		return args(parser.I8, parser.I8), args(parser.I8)
	case "modI16":
		return args(parser.I16, parser.I16), args(parser.I16)
	case "modI32":
		return args(parser.I32, parser.I32), args(parser.I32)
	case "modI64":
		return args(parser.I64, parser.I64), args(parser.I64)
	case "modI128":
		return args(parser.I128, parser.I128), args(parser.I128)
	case "lessU8":
		return args(parser.U8, parser.U8), args(parser.U8)
	case "lessU16":
		return args(parser.U16, parser.U16), args(parser.U8)
	case "lessU32":
		return args(parser.U32, parser.U32), args(parser.U8)
	case "lessU64":
		return args(parser.U64, parser.U64), args(parser.U8)
	case "lessU128":
		return args(parser.U128, parser.U128), args(parser.U8)
	case "lessI8":
		return args(parser.I8, parser.I8), args(parser.U8)
	case "lessI16":
		return args(parser.I16, parser.I16), args(parser.U8)
	case "lessI32":
		return args(parser.I32, parser.I32), args(parser.U8)
	case "lessI64":
		return args(parser.I64, parser.I64), args(parser.U8)
	case "lessI128":
		return args(parser.I128, parser.I128), args(parser.U8)
	case "lessEqU8":
		return args(parser.U8, parser.U8), args(parser.U8)
	case "lessEqU16":
		return args(parser.U16, parser.U16), args(parser.U8)
	case "lessEqU32":
		return args(parser.U32, parser.U32), args(parser.U8)
	case "lessEqU64":
		return args(parser.U64, parser.U64), args(parser.U8)
	case "lessEqU128":
		return args(parser.U128, parser.U128), args(parser.U8)
	case "lessEqI8":
		return args(parser.I8, parser.I8), args(parser.U8)
	case "lessEqI16":
		return args(parser.I16, parser.I16), args(parser.U8)
	case "lessEqI32":
		return args(parser.I32, parser.I32), args(parser.U8)
	case "lessEqI64":
		return args(parser.I64, parser.I64), args(parser.U8)
	case "lessEqI128":
		return args(parser.I128, parser.I128), args(parser.U8)
	case "greatU8":
		return args(parser.U8, parser.U8), args(parser.U8)
	case "greatU16":
		return args(parser.U16, parser.U16), args(parser.U8)
	case "greatU32":
		return args(parser.U32, parser.U32), args(parser.U8)
	case "greatU64":
		return args(parser.U64, parser.U64), args(parser.U8)
	case "greatU128":
		return args(parser.U128, parser.U128), args(parser.U8)
	case "greatI8":
		return args(parser.I8, parser.I8), args(parser.U8)
	case "greatI16":
		return args(parser.I16, parser.I16), args(parser.U8)
	case "greatI32":
		return args(parser.I32, parser.I32), args(parser.U8)
	case "greatI64":
		return args(parser.I64, parser.I64), args(parser.U8)
	case "greatI128":
		return args(parser.I128, parser.I128), args(parser.U8)
	case "greatEqU8":
		return args(parser.U8, parser.U8), args(parser.U8)
	case "greatEqU16":
		return args(parser.U16, parser.U16), args(parser.U8)
	case "greatEqU32":
		return args(parser.U32, parser.U32), args(parser.U8)
	case "greatEqU64":
		return args(parser.U64, parser.U64), args(parser.U8)
	case "greatEqU128":
		return args(parser.U128, parser.U128), args(parser.U8)
	case "greatEqI8":
		return args(parser.I8, parser.I8), args(parser.U8)
	case "greatEqI16":
		return args(parser.I16, parser.I16), args(parser.U8)
	case "greatEqI32":
		return args(parser.I32, parser.I32), args(parser.U8)
	case "greatEqI64":
		return args(parser.I64, parser.I64), args(parser.U8)
	case "greatEqI128":
		return args(parser.I128, parser.I128), args(parser.U8)
	case "8to16":
		return args(parser.U8), args(parser.U16)
	case "8to32":
		return args(parser.U8), args(parser.U32)
	case "8to64":
		return args(parser.U8), args(parser.U64)
	case "8to128":
		return args(parser.U8), args(parser.U128)
	case "16to8":
		return args(parser.U16), args(parser.U8)
	case "16to32":
		return args(parser.U16), args(parser.U32)
	case "16to64":
		return args(parser.U16), args(parser.U64)
	case "16to128":
		return args(parser.U16), args(parser.U128)
	case "32to8":
		return args(parser.U32), args(parser.U8)
	case "32to16":
		return args(parser.U32), args(parser.U16)
	case "32to64":
		return args(parser.U32), args(parser.U64)
	case "32to128":
		return args(parser.U32), args(parser.U128)
	case "64to8":
		return args(parser.U64), args(parser.U8)
	case "64to16":
		return args(parser.U64), args(parser.U16)
	case "64to32":
		return args(parser.U64), args(parser.U32)
	case "64to128":
		return args(parser.U64), args(parser.U128)
	case "128to8":
		return args(parser.U128), args(parser.U8)
	case "128to16":
		return args(parser.U128), args(parser.U16)
	case "128to32":
		return args(parser.U128), args(parser.U32)
	case "128to64":
		return args(parser.U128), args(parser.U64)
	case "load8":
		return args(parser.U64), args(parser.U8)
	case "load16":
		return args(parser.U64), args(parser.U16)
	case "load32":
		return args(parser.U64), args(parser.U32)
	case "load64":
		return args(parser.U64), args(parser.U64)
	case "load128":
		return args(parser.U64), args(parser.U128)
	case "store8":
		return args(parser.U64, parser.U8), args()
	case "store16":
		return args(parser.U64, parser.U16), args()
	case "store32":
		return args(parser.U64, parser.U32), args()
	case "store64":
		return args(parser.U64, parser.U64), args()
	case "store128":
		return args(parser.U64, parser.U128), args()
	case "jumpImm":
		return args(), args()
	case "jumpImmF":
		return args(), args()
	case "jumpImmB":
		return args(), args()
	case "sleepImm":
		return args(), args()
	case "branchImm":
		return args(parser.U8), args()
	case "branchImmF":
		return args(parser.U8), args()
	case "branchImmB":
		return args(parser.U8), args()
	case "callImm":
		return args(), args()
	case "load8Imm":
		return args(), args(parser.U8)
	case "load16Imm":
		return args(), args(parser.U16)
	case "load32Imm":
		return args(), args(parser.U32)
	case "load64Imm":
		return args(), args(parser.U64)
	case "load128Imm":
		return args(), args(parser.U128)
	case "store8Imm":
		return args(parser.U8), args()
	case "store16Imm":
		return args(parser.U16), args()
	case "store32Imm":
		return args(parser.U32), args()
	case "store64Imm":
		return args(parser.U64), args()
	case "store128Imm":
		return args(parser.U128), args()
	case "debug":
		return args(), args()
	case "debug8":
		return args(parser.U8), args()
	case "debug16":
		return args(parser.U16), args()
	case "debug32":
		return args(parser.U32), args()
	case "debug64":
		return args(parser.U64), args()
	case "debug128":
		return args(parser.U128), args()
	default:
		panic(fmt.Sprintf("invalid asm instruction '%s'", inst))
	}
}
