package sizeof

import (
	"math/rand"
	"testing"
)

type (
	tStruct struct {
		X          int
		Y          int
		Primitives *primit
		Simple     simple
		Intef      interface{}
		I          int
		z          int
	}

	primit struct {
		bts      []byte
		ptrBts   *[]byte
		sl       []int
		slPtr    []*int
		ptrInt   *int
		int      int
		str      string
		mapStr   map[int]string
		mapPtr   map[string]*int
		mapSlice map[uint][]int
		arr      [4]string
	}

	simple struct {
		A, B string
	}
)

func testStruct() *tStruct {
	return &tStruct{
		X: 99, Y: 100,
		Primitives: makePrimitives(),
		Simple:     simple{"XXX", "YYYY"},
		Intef: &struct {
			X, U int
		}{99, 500},
		I: 100,
		z: -1,
	}
}
func makePrimitives() *primit {
	return &primit{
		bts:    []byte{11, 23, 99, 100},
		ptrBts: &[]byte{11, 23, 99, 100},
		sl:     []int{1, 2, 3, 4},
		slPtr:  []*int{intPointer(), intPointer(), intPointer(), intPointer()},
		ptrInt: intPointer(),
		int:    *intPointer(),
		str:    "xxxxx",
		mapStr: map[int]string{1: "x", 2: "U", 3: "axadwa"},
		mapPtr: map[string]*int{"a": intPointer(), "b": intPointer(), "rrr": intPointer()},
		mapSlice: map[uint][]int{
			1: {1, 2, 3},
			7: {4, 7, 3, 9},
			9: {7, 3, 9, 6, 1},
		},
		arr: [4]string{"a", "b", "c", "d"},
	}
}
func intPointer() *int {
	v := rand.Int()
	return &v
}

func BenchmarkReflectCalcSize(b *testing.B) {
	tSt := testStruct()

	for i := 0; i < b.N; i++ {
		StructSize(tSt)
	}
}
