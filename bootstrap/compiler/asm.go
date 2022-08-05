package compiler

import (
	"bootstrap/parser"
	"fmt"
)

func (f *Fun) checkStackAsm(c *Ctx, stack []parser.Typ) []parser.Typ {
	for i := 0; i < len(f.fun.Block.Exprs); i++ {
		str := f.fun.Block.Exprs[i].AsString()
		if str == nil {
			panic(fmt.Sprintf("the asm fun '%s' can only contain strings", f.makeFunIdent(c)))
		}
		stack = f.checkStackInst(c, str.Content, stack)
	}
	return stack
}

func (f *Fun) checkStackInst(c *Ctx, inst string, stack []parser.Typ) []parser.Typ {
	inp, out := argsInst(inst)
	err, stack := c.stackPrefix(stack, inp...)
	if err {
		panic(fmt.Sprintf("the fun '%s' does not have a valid stack", f.makeFunIdent(c)))
	}
	return append(stack, out...)
}

func (f *Fun) checkStackAsmSimple(c *Ctx, stack int) int {
	for i := 0; i < len(f.fun.Block.Exprs); i++ {
		str := f.fun.Block.Exprs[i].AsString()
		if str == nil {
			panic(fmt.Sprintf("the asm fun '%s' can only contain strings", f.makeFunIdent(c)))
		}
		stack = f.checkStackInstSimple(c, str.Content, stack)
	}
	return stack
}

func (f *Fun) checkStackInstSimple(c *Ctx, inst string, stack int) int {
	inp, out := argsInst(inst)
	err, stack := stackPrefixSimple(c, stack, inp...)
	if err {
		panic(fmt.Sprintf("the fun '%s' does not have a valid stack", f.makeFunIdent(c)))
	}
	return stack + c.typsSize(out)
}

func (f *Fun) compileAsm(c *Ctx) []uint8 {
	bytes := []uint8{}

	for i := 0; i < len(f.fun.Block.Exprs); i++ {
		str := f.fun.Block.Exprs[i].AsString()
		if str == nil {
			panic(fmt.Sprintf("the asm fun '%s' can only contain strings", f.makeFunIdent(c)))
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
	case "read_file":
		return 8
	case "write_file":
		return 9
	case "push_imm_u8":
		return 10
	case "push_imm_u16":
		return 11
	case "push_imm_u32":
		return 12
	case "push_imm_u64":
		return 13
	case "push_imm_u128":
		return 14
	case "pop_sp":
		return 15
	case "pop_cs":
		return 16
	case "pop_ih":
		return 17
	case "pop_ir":
		return 18
	case "push_ir":
		return 19
	case "drop_u8":
		return 20
	case "drop_u16":
		return 21
	case "drop_u32":
		return 22
	case "drop_u64":
		return 23
	case "drop_u128":
		return 24
	case "negate_u8":
		return 25
	case "negate_u16":
		return 26
	case "negate_u32":
		return 27
	case "negate_u64":
		return 28
	case "negate_u128":
		return 29
	case "swap_u8":
		return 30
	case "swap_u16":
		return 31
	case "swap_u32":
		return 32
	case "swap_u64":
		return 33
	case "swap_u128":
		return 34
	case "rotate_u8":
		return 35
	case "rotate_u16":
		return 36
	case "rotate_u32":
		return 37
	case "rotate_u64":
		return 38
	case "rotate_u128":
		return 39
	case "dup_u8":
		return 40
	case "dup_u16":
		return 41
	case "dup_u32":
		return 42
	case "dup_u64":
		return 43
	case "dup_u128":
		return 44
	case "over_u8":
		return 45
	case "over_u16":
		return 46
	case "over_u32":
		return 47
	case "over_u64":
		return 48
	case "over_u128":
		return 49
	case "and_u8":
		return 50
	case "and_u16":
		return 51
	case "and_u32":
		return 52
	case "and_u64":
		return 53
	case "and_u128":
		return 54
	case "or_u8":
		return 55
	case "or_u16":
		return 56
	case "or_u32":
		return 57
	case "or_u64":
		return 58
	case "or_u128":
		return 59
	case "shift_l_u8":
		return 60
	case "shift_l_u16":
		return 61
	case "shift_l_u32":
		return 62
	case "shift_l_u64":
		return 63
	case "shift_l_u128":
		return 64
	case "shift_r_u8":
		return 65
	case "shift_r_u16":
		return 66
	case "shift_r_u32":
		return 67
	case "shift_r_u64":
		return 68
	case "shift_r_u128":
		return 69
	case "rotate_l_u8":
		return 70
	case "rotate_l_u16":
		return 71
	case "rotate_l_u32":
		return 72
	case "rotate_l_u64":
		return 73
	case "rotate_l_u128":
		return 74
	case "rotate_r_u8":
		return 75
	case "rotate_r_u16":
		return 76
	case "rotate_r_u32":
		return 77
	case "rotate_r_u64":
		return 78
	case "rotate_r_u128":
		return 79
	case "eq_u8":
		return 80
	case "eq_u16":
		return 81
	case "eq_u32":
		return 82
	case "eq_u64":
		return 83
	case "eq_u128":
		return 84
	case "not_eq_u8":
		return 85
	case "not_eq_u16":
		return 86
	case "not_eq_u32":
		return 87
	case "not_eq_u64":
		return 88
	case "not_eq_u128":
		return 89
	case "jump":
		return 90
	case "jump_f":
		return 91
	case "jump_b":
		return 92
	case "sleep":
		return 94
	case "branch":
		return 95
	case "branch_f":
		return 96
	case "branch_b":
		return 97
	case "add_u8":
		return 100
	case "add_u16":
		return 101
	case "add_u32":
		return 102
	case "add_u64":
		return 103
	case "add_u128":
		return 104
	case "add_i8":
		return 105
	case "add_i16":
		return 106
	case "add_i32":
		return 107
	case "add_i64":
		return 108
	case "add_i128":
		return 109
	case "sub_u8":
		return 110
	case "sub_u16":
		return 111
	case "sub_u32":
		return 112
	case "sub_u64":
		return 113
	case "sub_u128":
		return 114
	case "sub_i8":
		return 115
	case "sub_i16":
		return 116
	case "sub_i32":
		return 117
	case "sub_i64":
		return 118
	case "sub_i128":
		return 119
	case "mul_u8":
		return 120
	case "mul_u16":
		return 121
	case "mul_u32":
		return 122
	case "mul_u64":
		return 123
	case "mul_u128":
		return 124
	case "mul_i8":
		return 125
	case "mul_i16":
		return 126
	case "mul_i32":
		return 127
	case "mul_i64":
		return 128
	case "mul_i128":
		return 129
	case "div_u8":
		return 130
	case "div_u16":
		return 131
	case "div_u32":
		return 132
	case "div_u64":
		return 133
	case "div_u128":
		return 134
	case "div_i8":
		return 135
	case "div_i16":
		return 136
	case "div_i32":
		return 137
	case "div_i64":
		return 138
	case "div_i128":
		return 139
	case "mod_u8":
		return 140
	case "mod_u16":
		return 141
	case "mod_u32":
		return 142
	case "mod_u64":
		return 143
	case "mod_u128":
		return 144
	case "mod_i8":
		return 145
	case "mod_i16":
		return 146
	case "mod_i32":
		return 147
	case "mod_i64":
		return 148
	case "mod_i128":
		return 149
	case "less_u8":
		return 150
	case "less_u16":
		return 151
	case "less_u32":
		return 152
	case "less_u64":
		return 153
	case "less_u128":
		return 154
	case "less_i8":
		return 155
	case "less_i16":
		return 156
	case "less_i32":
		return 157
	case "less_i64":
		return 158
	case "less_i128":
		return 159
	case "less_eq_u8":
		return 160
	case "less_eq_u16":
		return 161
	case "less_eq_u32":
		return 162
	case "less_eq_u64":
		return 163
	case "less_eq_u128":
		return 164
	case "less_eq_i8":
		return 165
	case "less_eq_i16":
		return 166
	case "less_eq_i32":
		return 167
	case "less_eq_i64":
		return 168
	case "less_eq_i128":
		return 169
	case "great_u8":
		return 170
	case "great_u16":
		return 171
	case "great_u32":
		return 172
	case "great_u64":
		return 173
	case "great_u128":
		return 174
	case "great_i8":
		return 175
	case "great_i16":
		return 176
	case "great_i32":
		return 177
	case "great_i64":
		return 178
	case "great_i128":
		return 179
	case "great_eq_u8":
		return 180
	case "great_eq_u16":
		return 181
	case "great_eq_u32":
		return 182
	case "great_eq_u64":
		return 183
	case "great_eq_u128":
		return 184
	case "great_eq_i8":
		return 185
	case "great_eq_i16":
		return 186
	case "great_eq_i32":
		return 187
	case "great_eq_i64":
		return 188
	case "great_eq_i128":
		return 189
	case "u8_to_u16":
		return 190
	case "u8_to_u32":
		return 191
	case "u8_to_u64":
		return 192
	case "u8_to_u128":
		return 193
	case "u16_to_u8":
		return 194
	case "u16_to_u32":
		return 195
	case "u16_to_u64":
		return 196
	case "u16_to_u128":
		return 197
	case "u32_to_u8":
		return 198
	case "u32_to_u16":
		return 199
	case "u32_to_u64":
		return 200
	case "u32_to_u128":
		return 201
	case "u64_to_u8":
		return 202
	case "u64_to_u16":
		return 203
	case "u64_to_u32":
		return 204
	case "u64_to_u128":
		return 205
	case "u128_to_u8":
		return 206
	case "u128_to_u16":
		return 207
	case "u128_to_u32":
		return 208
	case "u128_to_u64":
		return 209
	case "load_u8":
		return 210
	case "load_u16":
		return 211
	case "load_u32":
		return 212
	case "load_u64":
		return 213
	case "load_u128":
		return 214
	case "store_u8":
		return 215
	case "store_u16":
		return 216
	case "store_u32":
		return 217
	case "store_u64":
		return 218
	case "store_u128":
		return 219
	case "jump_imm":
		return 220
	case "jump_imm_f":
		return 221
	case "jump_imm_b":
		return 222
	case "sleep_imm":
		return 224
	case "branch_imm":
		return 225
	case "branch_imm_f":
		return 226
	case "branch_imm_b":
		return 227
	case "call_imm":
		return 229
	case "load_imm_u8":
		return 230
	case "load_imm_u16":
		return 231
	case "load_imm_u32":
		return 232
	case "load_imm_u64":
		return 233
	case "load_imm_u128":
		return 234
	case "store_imm_u8":
		return 235
	case "store_imm_u16":
		return 236
	case "store_imm_u32":
		return 237
	case "store_imm_u64":
		return 238
	case "store_imm_u128":
		return 239
	case "debug":
		return 250
	case "debug_u8":
		return 251
	case "debug_u16":
		return 252
	case "debug_u32":
		return 253
	case "debug_u64":
		return 254
	case "debug_u128":
		return 255
	default:
		panic(fmt.Sprintf("invalid asm instruction '%s'", inst))
	}
}

type Args struct {
	typs []parser.Builtin
}

func args(args ...parser.Builtin) []parser.Typ {
	res := []parser.Typ{}
	for _, arg := range args {
		res = append(res, &arg)
	}
	return res
}

func argsInst(inst string) ([]parser.Typ, []parser.Typ) {
	switch inst {
	// 000
	case "nop":
		return args(), args()
	// 001
	case "halt":
		return args(), args()
	// 002
	case "call":
		return args(parser.U64), args()
	// 003
	case "return":
		return args(), args()
	// 004
	case "inter":
		return args(), args()
	// 005
	case "alloc":
		return args(parser.U64), args(parser.U64)
	// 006
	case "read":
		return args(parser.STRING), args(parser.U64)
	// 007
	case "write":
		return args(parser.STRING), args(parser.U64)
	// 008
	case "read_file":
		return args(parser.STRING, parser.STRING), args(parser.U64)
	// 009
	case "write_file":
		return args(parser.STRING, parser.STRING), args(parser.U64)
	// 010
	case "push_imm_u8":
		return args(), args(parser.U8)
	// 011
	case "push_imm_u16":
		return args(), args(parser.U16)
	// 012
	case "push_imm_u32":
		return args(), args(parser.U32)
	// 013
	case "push_imm_u64":
		return args(), args(parser.U64)
	// 014
	case "push_imm_u128":
		return args(), args(parser.U128)
	// 015
	case "pop_sp":
		return args(parser.U64), args()
	// 016
	case "pop_cs":
		return args(parser.U64), args()
	// 017
	case "pop_ih":
		return args(parser.U64), args()
	// 018
	case "pop_ir":
		return args(parser.I8), args()
	// 019
	case "push_ir":
		return args(), args(parser.I8)
	// 020
	case "drop_u8":
		return args(parser.U8), args()
	// 021
	case "drop_u16":
		return args(parser.U16), args()
	// 022
	case "drop_u32":
		return args(parser.U32), args()
	// 023
	case "drop_u64":
		return args(parser.U64), args()
	// 024
	case "drop_u128":
		return args(parser.U128), args()
	// 025
	case "negate_u8":
		return args(parser.U8), args(parser.I8)
	// 026
	case "negate_u16":
		return args(parser.U16), args(parser.I16)
	// 027
	case "negate_u32":
		return args(parser.U32), args(parser.I32)
	// 028
	case "negate_u64":
		return args(parser.U64), args(parser.I64)
	// 029
	case "negate_u128":
		return args(parser.U128), args(parser.I128)
	// 030
	case "swap_u8":
		return args(parser.U8, parser.U8), args(parser.U8, parser.U8)
	// 031
	case "swap_u16":
		return args(parser.U16, parser.U16), args(parser.U16, parser.U16)
	// 032
	case "swap_u32":
		return args(parser.U32, parser.U32), args(parser.U32, parser.U32)
	// 033
	case "swap_u64":
		return args(parser.U64, parser.U64), args(parser.U64, parser.U64)
	// 034
	case "swap_u128":
		return args(parser.U128, parser.U128), args(parser.U128, parser.U128)
	// 035
	case "rotate_u8":
		return args(parser.U8, parser.U8, parser.U8), args(parser.U8, parser.U8, parser.U8)
	// 036
	case "rotate_u16":
		return args(parser.U16, parser.U16, parser.U16), args(parser.U16, parser.U16, parser.U16)
	// 037
	case "rotate_u32":
		return args(parser.U32, parser.U32, parser.U32), args(parser.U32, parser.U32, parser.U32)
	// 038
	case "rotate_u64":
		return args(parser.U64, parser.U64, parser.U64), args(parser.U64, parser.U64, parser.U64)
	// 039
	case "rotate_u128":
		return args(parser.U128, parser.U128, parser.U128), args(parser.U128, parser.U128, parser.U128)
	// 040
	case "dup_u8":
		return args(parser.U8), args(parser.U8, parser.U8)
	// 041
	case "dup_u16":
		return args(parser.U16), args(parser.U16, parser.U16)
	// 042
	case "dup_u32":
		return args(parser.U32), args(parser.U32, parser.U32)
	// 043
	case "dup_u64":
		return args(parser.U64), args(parser.U64, parser.U64)
	// 044
	case "dup_u128":
		return args(parser.U128), args(parser.U128, parser.U128)
	// 045
	case "over_u8":
		return args(parser.U8, parser.U8), args(parser.U8, parser.U8, parser.U8)
	// 046
	case "over_u16":
		return args(parser.U16, parser.U16), args(parser.U16, parser.U16, parser.U16)
	// 047
	case "over_u32":
		return args(parser.U32, parser.U32), args(parser.U32, parser.U32, parser.U32)
	// 048
	case "over_u64":
		return args(parser.U64, parser.U64), args(parser.U64, parser.U64, parser.U64)
	// 049
	case "over_u128":
		return args(parser.U128, parser.U128), args(parser.U128, parser.U128, parser.U128)
	// 050
	case "and_u8":
		return args(parser.U8, parser.U8), args(parser.U8)
	// 051
	case "and_u16":
		return args(parser.U16, parser.U16), args(parser.U16)
	// 052
	case "and_u32":
		return args(parser.U32, parser.U32), args(parser.U32)
	// 053
	case "and_u64":
		return args(parser.U64, parser.U64), args(parser.U64)
	// 054
	case "and_u128":
		return args(parser.U128, parser.U128), args(parser.U128)
	// 055
	case "or_u8":
		return args(parser.U8, parser.U8), args(parser.U8)
	// 056
	case "or_u16":
		return args(parser.U16, parser.U16), args(parser.U16)
	// 057
	case "or_u32":
		return args(parser.U32, parser.U32), args(parser.U32)
	// 058
	case "or_u64":
		return args(parser.U64, parser.U64), args(parser.U64)
	// 059
	case "or_u128":
		return args(parser.U128, parser.U128), args(parser.U128)
	// 060
	case "shift_l_u8":
		return args(parser.U8, parser.U8), args(parser.U8)
	// 061
	case "shift_l_u16":
		return args(parser.U16, parser.U8), args(parser.U16)
	// 062
	case "shift_l_u32":
		return args(parser.U32, parser.U8), args(parser.U32)
	// 063
	case "shift_l_u64":
		return args(parser.U64, parser.U8), args(parser.U64)
	// 064
	case "shift_l_u128":
		return args(parser.U128, parser.U8), args(parser.U128)
	// 065
	case "shift_r_u8":
		return args(parser.U8, parser.U8), args(parser.U8)
	// 066
	case "shift_r_u16":
		return args(parser.U16, parser.U8), args(parser.U16)
	// 067
	case "shift_r_u32":
		return args(parser.U32, parser.U8), args(parser.U32)
	// 068
	case "shift_r_u64":
		return args(parser.U64, parser.U8), args(parser.U64)
	// 069
	case "shift_r_u128":
		return args(parser.U128, parser.U8), args(parser.U128)
	// 070
	case "rotate_l_u8":
		return args(parser.U8, parser.U8), args(parser.U8)
	// 071
	case "rotate_l_u16":
		return args(parser.U16, parser.U8), args(parser.U16)
	// 072
	case "rotate_l_u32":
		return args(parser.U32, parser.U8), args(parser.U32)
	// 073
	case "rotate_l_u64":
		return args(parser.U64, parser.U8), args(parser.U64)
	// 074
	case "rotate_l_u128":
		return args(parser.U128, parser.U8), args(parser.U128)
	// 075
	case "rotate_r_u8":
		return args(parser.U8, parser.U8), args(parser.U8)
	// 076
	case "rotate_r_u16":
		return args(parser.U16, parser.U8), args(parser.U16)
	// 077
	case "rotate_r_u32":
		return args(parser.U32, parser.U8), args(parser.U32)
	// 078
	case "rotate_r_u64":
		return args(parser.U64, parser.U8), args(parser.U64)
	// 079
	case "rotate_r_u128":
		return args(parser.U128, parser.U8), args(parser.U128)
	// 080
	case "eq_u8":
		return args(parser.U8, parser.U8), args(parser.U8)
	// 081
	case "eq_u16":
		return args(parser.U16, parser.U16), args(parser.U8)
	// 082
	case "eq_u32":
		return args(parser.U32, parser.U32), args(parser.U8)
	// 083
	case "eq_u64":
		return args(parser.U64, parser.U64), args(parser.U8)
	// 084
	case "eq_u128":
		return args(parser.U128, parser.U128), args(parser.U8)
	// 085
	case "not_eq_u8":
		return args(parser.U8, parser.U8), args(parser.U8)
	// 086
	case "not_eq_u16":
		return args(parser.U16, parser.U16), args(parser.U8)
	// 087
	case "not_eq_u32":
		return args(parser.U32, parser.U32), args(parser.U8)
	// 088
	case "not_eq_u64":
		return args(parser.U64, parser.U64), args(parser.U8)
	// 089
	case "not_eq_u128":
		return args(parser.U128, parser.U128), args(parser.U8)
	// 090
	case "jump":
		return args(parser.U64), args()
	// 091
	case "jump_f":
		return args(parser.U64), args()
	// 092
	case "jump_b":
		return args(parser.U64), args()
	// 094
	case "sleep":
		return args(parser.U64), args()
	// 095
	case "branch":
		return args(parser.U64, parser.U8), args()
	// 096
	case "branch_f":
		return args(parser.U64, parser.U8), args()
	// 097
	case "branch_b":
		return args(parser.U64, parser.U8), args()
	// 100
	case "add_u8":
		return args(parser.U8, parser.U8), args(parser.U8)
	// 101
	case "add_u16":
		return args(parser.U16, parser.U16), args(parser.U16)
	// 102
	case "add_u32":
		return args(parser.U32, parser.U32), args(parser.U32)
	// 103
	case "add_u64":
		return args(parser.U64, parser.U64), args(parser.U64)
	// 104
	case "add_u128":
		return args(parser.U128, parser.U128), args(parser.U128)
	// 105
	case "add_i8":
		return args(parser.I8, parser.I8), args(parser.I8)
	// 106
	case "add_i16":
		return args(parser.I16, parser.I16), args(parser.I16)
	// 107
	case "add_i32":
		return args(parser.I32, parser.I32), args(parser.I32)
	// 108
	case "add_i64":
		return args(parser.I64, parser.I64), args(parser.I64)
	// 109
	case "add_i128":
		return args(parser.I128, parser.I128), args(parser.I128)
	// 110
	case "sub_u8":
		return args(parser.U8, parser.U8), args(parser.U8)
	// 111
	case "sub_u16":
		return args(parser.U16, parser.U16), args(parser.U16)
	// 112
	case "sub_u32":
		return args(parser.U32, parser.U32), args(parser.U32)
	// 113
	case "sub_u64":
		return args(parser.U64, parser.U64), args(parser.U64)
	// 114
	case "sub_u128":
		return args(parser.U128, parser.U128), args(parser.U128)
	// 115
	case "sub_i8":
		return args(parser.I8, parser.I8), args(parser.I8)
	// 116
	case "sub_i16":
		return args(parser.I16, parser.I16), args(parser.I16)
	// 117
	case "sub_i32":
		return args(parser.I32, parser.I32), args(parser.I32)
	// 118
	case "sub_i64":
		return args(parser.I64, parser.I64), args(parser.I64)
	// 119
	case "sub_i128":
		return args(parser.I128, parser.I128), args(parser.I128)
	// 120
	case "mul_u8":
		return args(parser.U8, parser.U8), args(parser.U8)
	// 121
	case "mul_u16":
		return args(parser.U16, parser.U16), args(parser.U16)
	// 122
	case "mul_u32":
		return args(parser.U32, parser.U32), args(parser.U32)
	// 123
	case "mul_u64":
		return args(parser.U64, parser.U64), args(parser.U64)
	// 124
	case "mul_u128":
		return args(parser.U128, parser.U128), args(parser.U128)
	// 125
	case "mul_i8":
		return args(parser.I8, parser.I8), args(parser.I8)
	// 126
	case "mul_i16":
		return args(parser.I16, parser.I16), args(parser.I16)
	// 127
	case "mul_i32":
		return args(parser.I32, parser.I32), args(parser.I32)
	// 128
	case "mul_i64":
		return args(parser.I64, parser.I64), args(parser.I64)
	// 129
	case "mul_i128":
		return args(parser.I128, parser.I128), args(parser.I128)
	// 130
	case "div_u8":
		return args(parser.U8, parser.U8), args(parser.U8)
	// 131
	case "div_u16":
		return args(parser.U16, parser.U16), args(parser.U16)
	// 132
	case "div_u32":
		return args(parser.U32, parser.U32), args(parser.U32)
	// 133
	case "div_u64":
		return args(parser.U64, parser.U64), args(parser.U64)
	// 134
	case "div_u128":
		return args(parser.U128, parser.U128), args(parser.U128)
	// 135
	case "div_i8":
		return args(parser.I8, parser.I8), args(parser.I8)
	// 136
	case "div_i16":
		return args(parser.I16, parser.I16), args(parser.I16)
	// 137
	case "div_i32":
		return args(parser.I32, parser.I32), args(parser.I32)
	// 138
	case "div_i64":
		return args(parser.I64, parser.I64), args(parser.I64)
	// 139
	case "div_i128":
		return args(parser.I128, parser.I128), args(parser.I128)
	// 140
	case "mod_u8":
		return args(parser.U8, parser.U8), args(parser.U8)
	// 141
	case "mod_u16":
		return args(parser.U16, parser.U16), args(parser.U16)
	// 142
	case "mod_u32":
		return args(parser.U32, parser.U32), args(parser.U32)
	// 143
	case "mod_u64":
		return args(parser.U64, parser.U64), args(parser.U64)
	// 144
	case "mod_u128":
		return args(parser.U128, parser.U128), args(parser.U128)
	// 145
	case "mod_i8":
		return args(parser.I8, parser.I8), args(parser.I8)
	// 146
	case "mod_i16":
		return args(parser.I16, parser.I16), args(parser.I16)
	// 147
	case "mod_i32":
		return args(parser.I32, parser.I32), args(parser.I32)
	// 148
	case "mod_i64":
		return args(parser.I64, parser.I64), args(parser.I64)
	// 149
	case "mod_i128":
		return args(parser.I128, parser.I128), args(parser.I128)
	// 150
	case "less_u8":
		return args(parser.U8, parser.U8), args(parser.U8)
	// 151
	case "less_u16":
		return args(parser.U16, parser.U16), args(parser.U8)
	// 152
	case "less_u32":
		return args(parser.U32, parser.U32), args(parser.U8)
	// 153
	case "less_u64":
		return args(parser.U64, parser.U64), args(parser.U8)
	// 154
	case "less_u128":
		return args(parser.U128, parser.U128), args(parser.U8)
	// 155
	case "less_i8":
		return args(parser.I8, parser.I8), args(parser.U8)
	// 156
	case "less_i16":
		return args(parser.I16, parser.I16), args(parser.U8)
	// 157
	case "less_i32":
		return args(parser.I32, parser.I32), args(parser.U8)
	// 158
	case "less_i64":
		return args(parser.I64, parser.I64), args(parser.U8)
	// 159
	case "less_i128":
		return args(parser.I128, parser.I128), args(parser.U8)
	// 160
	case "less_eq_u8":
		return args(parser.U8, parser.U8), args(parser.U8)
	// 161
	case "less_eq_u16":
		return args(parser.U16, parser.U16), args(parser.U8)
	// 162
	case "less_eq_u32":
		return args(parser.U32, parser.U32), args(parser.U8)
	// 163
	case "less_eq_u64":
		return args(parser.U64, parser.U64), args(parser.U8)
	// 164
	case "less_eq_u128":
		return args(parser.U128, parser.U128), args(parser.U8)
	// 165
	case "less_eq_i8":
		return args(parser.I8, parser.I8), args(parser.U8)
	// 166
	case "less_eq_i16":
		return args(parser.I16, parser.I16), args(parser.U8)
	// 167
	case "less_eq_i32":
		return args(parser.I32, parser.I32), args(parser.U8)
	// 168
	case "less_eq_i64":
		return args(parser.I64, parser.I64), args(parser.U8)
	// 169
	case "less_eq_i128":
		return args(parser.I128, parser.I128), args(parser.U8)
	// 170
	case "great_u8":
		return args(parser.U8, parser.U8), args(parser.U8)
	// 171
	case "great_u16":
		return args(parser.U16, parser.U16), args(parser.U8)
	// 172
	case "great_u32":
		return args(parser.U32, parser.U32), args(parser.U8)
	// 173
	case "great_u64":
		return args(parser.U64, parser.U64), args(parser.U8)
	// 174
	case "great_u128":
		return args(parser.U128, parser.U128), args(parser.U8)
	// 175
	case "great_i8":
		return args(parser.I8, parser.I8), args(parser.U8)
	// 176
	case "great_i16":
		return args(parser.I16, parser.I16), args(parser.U8)
	// 177
	case "great_i32":
		return args(parser.I32, parser.I32), args(parser.U8)
	// 178
	case "great_i64":
		return args(parser.I64, parser.I64), args(parser.U8)
	// 179
	case "great_i128":
		return args(parser.I128, parser.I128), args(parser.U8)
	// 180
	case "great_eq_u8":
		return args(parser.U8, parser.U8), args(parser.U8)
	// 181
	case "great_eq_u16":
		return args(parser.U16, parser.U16), args(parser.U8)
	// 182
	case "great_eq_u32":
		return args(parser.U32, parser.U32), args(parser.U8)
	// 183
	case "great_eq_u64":
		return args(parser.U64, parser.U64), args(parser.U8)
	// 184
	case "great_eq_u128":
		return args(parser.U128, parser.U128), args(parser.U8)
	// 185
	case "great_eq_i8":
		return args(parser.I8, parser.I8), args(parser.U8)
	// 186
	case "great_eq_i16":
		return args(parser.I16, parser.I16), args(parser.U8)
	// 187
	case "great_eq_i32":
		return args(parser.I32, parser.I32), args(parser.U8)
	// 188
	case "great_eq_i64":
		return args(parser.I64, parser.I64), args(parser.U8)
	// 189
	case "great_eq_i128":
		return args(parser.I128, parser.I128), args(parser.U8)
	// 190
	case "u8_to_u16":
		return args(parser.U8), args(parser.U16)
	// 191
	case "u8_to_u32":
		return args(parser.U8), args(parser.U32)
	// 192
	case "u8_to_u64":
		return args(parser.U8), args(parser.U64)
	// 193
	case "u8_to_u128":
		return args(parser.U8), args(parser.U128)
	// 194
	case "u16_to_u8":
		return args(parser.U16), args(parser.U8)
	// 195
	case "u16_to_u32":
		return args(parser.U16), args(parser.U32)
	// 196
	case "u16_to_u64":
		return args(parser.U16), args(parser.U64)
	// 197
	case "u16_to_u128":
		return args(parser.U16), args(parser.U128)
	// 198
	case "u32_to_u8":
		return args(parser.U32), args(parser.U8)
	// 199
	case "u32_to_u16":
		return args(parser.U32), args(parser.U16)
	// 200
	case "u32_to_u64":
		return args(parser.U32), args(parser.U64)
	// 201
	case "u32_to_u128":
		return args(parser.U32), args(parser.U128)
	// 202
	case "u64_to_u8":
		return args(parser.U64), args(parser.U8)
	// 203
	case "u64_to_u16":
		return args(parser.U64), args(parser.U16)
	// 204
	case "u64_to_u32":
		return args(parser.U64), args(parser.U32)
	// 205
	case "u64_to_u128":
		return args(parser.U64), args(parser.U128)
	// 206
	case "u128_to_u8":
		return args(parser.U128), args(parser.U8)
	// 207
	case "u128_to_u16":
		return args(parser.U128), args(parser.U16)
	// 208
	case "u128_to_u32":
		return args(parser.U128), args(parser.U32)
	// 209
	case "u128_to_u64":
		return args(parser.U128), args(parser.U64)
	// 210
	case "load_u8":
		return args(parser.U64), args(parser.U8)
	// 211
	case "load_u16":
		return args(parser.U64), args(parser.U16)
	// 212
	case "load_u32":
		return args(parser.U64), args(parser.U32)
	// 213
	case "load_u64":
		return args(parser.U64), args(parser.U64)
	// 214
	case "load_u128":
		return args(parser.U64), args(parser.U128)
	// 215
	case "store_u8":
		return args(parser.U64, parser.U8), args()
	// 216
	case "store_u16":
		return args(parser.U64, parser.U16), args()
	// 217
	case "store_u32":
		return args(parser.U64, parser.U32), args()
	// 218
	case "store_u64":
		return args(parser.U64, parser.U64), args()
	// 219
	case "store_u128":
		return args(parser.U64, parser.U128), args()
	// 220
	case "jump_imm":
		return args(), args()
	// 221
	case "jump_imm_f":
		return args(), args()
	// 222
	case "jump_imm_b":
		return args(), args()
	// 224
	case "sleep_imm":
		return args(), args()
	// 225
	case "branch_imm":
		return args(parser.U8), args()
	// 226
	case "branch_imm_f":
		return args(parser.U8), args()
	// 227
	case "branch_imm_b":
		return args(parser.U8), args()
	// 229
	case "call_imm":
		return args(), args()
	// 230
	case "load_imm_u8":
		return args(), args(parser.U8)
	// 231
	case "load_imm_u16":
		return args(), args(parser.U16)
	// 232
	case "load_imm_u32":
		return args(), args(parser.U32)
	// 233
	case "load_imm_u64":
		return args(), args(parser.U64)
	// 234
	case "load_imm_u128":
		return args(), args(parser.U128)
	// 235
	case "store_imm_u8":
		return args(parser.U8), args()
	// 236
	case "store_imm_u16":
		return args(parser.U16), args()
	// 237
	case "store_imm_u32":
		return args(parser.U32), args()
	// 238
	case "store_imm_u64":
		return args(parser.U64), args()
	// 239
	case "store_imm_u128":
		return args(parser.U128), args()
	// 250
	case "debug":
		return args(), args()
	// 251
	case "debug_u8":
		return args(parser.U8), args()
	// 252
	case "debug_u16":
		return args(parser.U16), args()
	// 253
	case "debug_u32":
		return args(parser.U32), args()
	// 254
	case "debug_u64":
		return args(parser.U64), args()
	// 255
	case "debug_u128":
		return args(parser.U128), args()
	default:
		panic(fmt.Sprintf("invalid asm instruction '%s'", inst))
	}
}
