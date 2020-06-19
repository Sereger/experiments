### Description

This package calculates full size of go struct including size of the nested entities (structs, slices, maps, etc). For example:
```go
type x struct {
    sl []int64 // int64 - 8 bytes 
}

xVal := x{
    sl: make([]int64, 10, 100),
}

fmt.Println(runtime.SizeOf(xVal), StructSize(xVal))
// Output: 24 824
// 24 bytes - it just size of slice without data array
// and StructSize calculate size with allocated memory for data 
```

But, it now so fast:
```shell script
$ make bench
pkg: github.com/Sereger/experiments/sizeof
BenchmarkReflectCalcSize-8        346278              2987 ns/op             950 B/op         22 allocs/op
```

### Usage

```go
package main

import "github.com/Sereger/experiments/sizeof"

func main()  {
	myVar := myType{}
	size := sizeof.StructSize(myVar)
}
```