package sizeof

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	type (
		testStruct struct { // size 17
			A int64
			B int8
			c int64
		}
		recursiveStruct struct {
			child *recursiveStruct
		}
	)
	type tCase struct {
		name   string
		v      interface{}
		expect int64
	}
	recStruct := &recursiveStruct{
		child: &recursiveStruct{},
	}
	recStruct.child.child = recStruct

	tests := []tCase{
		{name: "int", v: int(10), expect: int64(unsafe.Sizeof(int(10)))},
		{name: "array", v: [3]int64{}, expect: 24},
		{name: "str", v: "abc", expect: 3 + 16}, // 3 - abc, 8 - prt to data, 8 - size of len
		{name: "str cyrillic", v: "абв", expect: 6 + 16},
		{name: "slice", v: make([]int64, 10, 20), expect: 184}, // 24 - struct slice, 10 * 8 - vals, 10*8 - reserved
		{name: "empty struct", v: struct{}{}, expect: 0},
		{name: "struct", v: struct {
			a int64
			b *testStruct
		}{}, expect: 16},
		{name: "struct with ptr to testStruct", v: struct {
			a int64
			b *testStruct
		}{b: new(testStruct)}, expect: 16 + 17},
		{name: "struct with string", v: struct {
			s string
		}{s: "abc"}, expect: 19},
		{name: "ptr to struct", v: new(testStruct), expect: 25}, // 17 + 8 (ptr)
		{name: "nil", v: nil, expect: 0},
		{name: "recursive", v: recStruct, expect: 16},
		{name: "empty map", v: map[int64]struct{}{}, expect: 8},
		{name: "map struct{}", v: map[int64]struct{}{0: {}, 1: {}}, expect: 24},
		{name: "map int:str", v: map[int64]string{0: "abc", 1: ""}, expect: 59},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := StructSize(test.v)
			require.Equal(t, test.expect, s)
		})
	}
}
