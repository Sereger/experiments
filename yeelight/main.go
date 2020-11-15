package main

import "fmt"

func main() {
	y, err := Discover()
	checkError(err)

	notyfy, _, err := y.Listen()
	checkError(err)
	for msg := range notyfy {
		fmt.Printf("%+v", msg)
	}
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
