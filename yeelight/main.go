package main

import (
	"fmt"
	"github.com/Sereger/experiments/yeelight/internal/yeelight"
	"github.com/kljensen/snowball"
	"strings"
)

func main1() {
	ys, err := yeelight.Discover()
	checkError(err)

	for _, y := range ys {
		fmt.Println(y.Name)
	}
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	stemmed, err := parseName("включи выключи свет в зале комнате ярчее светлее темнее светло сделай")
	if err == nil {
		fmt.Println(stemmed) // Prints "accumul"
	}
	fmt.Println(stemmed)
}

func parseName(name string) (tokens []string, err error) {
	words := strings.Split(name, " ")
	for i, w := range words {
		words[i], err = snowball.Stem(w, "russian", false)
		if err != nil {
			continue
		}
	}

	n := 0
	for _, t := range words {
		if len(t) < 3 {
			continue
		}
		switch t {
		case "свет", "комнат":

		}
		words[n] = t
		n++
	}

	return words[:n], nil
}
