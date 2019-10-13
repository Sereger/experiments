package main

import (
	"fmt"
	"go.starlark.net/starlark"
	"log"
	"time"
)

const script = `
print("abc")
def x():
	print("abc2")
	print(xxx)
	if xxx > 100:
		return "UUU"
	
	return "PPPPP"

y = x()
`

func main() {
	// The Thread defines the behavior of the built-in 'print' function.
	thread := &starlark.Thread{
		Name:  "example",
		Print: func(_ *starlark.Thread, msg string) { fmt.Println(msg) },
	}

	s := time.Now()
	_, p, err := starlark.SourceProgram("test", script, func(s string) bool {
		if s == "xxx" {
			return true
		}
		return false
	})
	ts := time.Since(s)
	if err != nil {
		if evalErr, ok := err.(*starlark.EvalError); ok {
			log.Fatal(evalErr.Backtrace())
		}
		log.Fatal(err)
	}

	// This dictionary defines the pre-declared environment.
	predeclared := starlark.StringDict{
		"xxx": starlark.MakeInt(100),
	}

	s = time.Now()
	globals, err := p.Init(thread, predeclared)
	ts1 := time.Since(s)
	if err != nil {
		if evalErr, ok := err.(*starlark.EvalError); ok {
			log.Fatal(evalErr.Backtrace())
		}
		log.Fatal(err)
	}

	// Print the global environment.
	fmt.Println("\nGlobals:")
	for _, name := range globals.Keys() {
		v := globals[name]
		fmt.Printf("%s (%s) = %s\n", name, v.Type(), v.String())
	}
	fmt.Printf("duration: %s %s\n", ts, ts1)
}
