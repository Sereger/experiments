package controller

import (
	"fmt"
	"github.com/Sereger/experiments/yeelight/internal/yeelight"
	"github.com/kljensen/snowball"
	"log"
	"strings"
)

var names = map[string]string{
	"192.168.0.19": "спальня",
}

type Controller struct{}

func (c *Controller) ExecuteCommand(tokens []string) error {
	devices, err := yeelight.Discover()
	if err != nil || len(devices) == 0 {
		return fmt.Errorf("Мне не удалось найти устройства")
	}

	deviceTokens := make(map[string][]*yeelight.Yeelight)
	for _, y := range devices {
		delim := strings.IndexByte(y.Addr, ':')
		name, ok := names[y.Addr[:delim]]
		if !ok {
			continue
		}

		nameTokens, err := parseName(name)
		if err != nil {
			log.Print(err)
			continue
		}
		for _, t := range nameTokens {
			deviceTokens[t] = append(deviceTokens[t], y)
		}
	}
	if len(deviceTokens) == 0 {
		return fmt.Errorf("Мне не удалось найти устройства")
	}

	var command, value string
	var target *yeelight.Yeelight
	for _, w := range tokens {
		token, err := snowball.Stem(w, "russian", false)
		if err != nil {
			continue
		}

		switch token {
		case "включ":
			command, value = "set_power", "on"
		case "выключ":
			command, value = "set_power", "off"
		case "ярч", "светл":
			command, value = "set_bright", "+"
		case "темн":
			command, value = "set_bright", "-"
		}

		devs, ok := deviceTokens[token]
		if !ok || len(devs) > 1 {
			continue
		}
		target = devs[0]
	}

	if target == nil {
		return fmt.Errorf("Мне не удалось найти подходящее устройство")
	}
	if command == "" || value == "" {
		return fmt.Errorf("Мне не удалось распознать команду")
	}

	switch command {
	case "set_power":
		target.SetPower(value == "on")
	case "set_bright":
		switch value {
		case "+":
			target.SetBright(target.Bright + 20)
		case "-":
			target.SetBright(target.Bright - 20)
		}
	}

	return nil
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
