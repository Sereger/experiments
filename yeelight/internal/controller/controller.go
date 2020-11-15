package controller

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/Sereger/experiments/yeelight/internal/session"
	"github.com/Sereger/experiments/yeelight/internal/yeelight"
	"github.com/kljensen/snowball"
)

var names = map[string]string{
	"192.168.0.19": "спальня",
}

type Controller struct{}

func (c *Controller) ExecuteCommand(tokens []string, sess *session.Session) (error, bool) {
	devices, err := c.resolveDevices(sess)
	if err != nil {
		return err, false
	}
	deviceTokens := c.devicesTokens(devices, sess)
	if len(deviceTokens) == 0 {
		return fmt.Errorf("Мне не удалось найти устройства"), false
	}

	var command, value string
	targets := make(map[*yeelight.Yeelight]struct{})
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
		case "яркост":
			command = "set_bright"
		case "максимальн":
			value = "100"
		case "половин", "наполовин":
			value = "50"
		case "минимальн":
			value = "20"
		}

		devs, ok := deviceTokens[token]
		if !ok {
			continue
		}

		for _, d := range devs {
			targets[d] = struct{}{}
		}
	}

	if len(targets) == 0 && len(sess.LastTargets) == 0 {
		return fmt.Errorf("Мне не удалось найти подходящее устройство"), false
	} else if len(targets) == 0 {
		targets = sess.LastTargets
	}
	sess.LastTargets = targets

	if command == "" || value == "" {
		return fmt.Errorf("Мне не удалось распознать команду"), true
	}

	var continueSession bool
	for device := range targets {
		switch command {
		case "set_power":
			device.SetPower(value == "on")
		case "set_bright":
			continueSession = true
			switch value {
			case "+":
				device.SetBright(device.Bright + 20)
			case "-":
				device.SetBright(device.Bright - 20)
			default:
				v, _ := strconv.ParseInt(value, 10, 64)
				if v > 0 {
					device.SetBright(v)
				}
			}
		}
	}

	return nil, continueSession
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

func (c *Controller) resolveDevices(sess *session.Session) ([]*yeelight.Yeelight, error) {
	if len(sess.Devices) > 0 {
		return sess.Devices, nil
	}

	devices, err := yeelight.Discover()
	if err != nil || len(devices) == 0 {
		return nil, fmt.Errorf("Мне не удалось найти устройства")
	}

	sess.Devices = devices
	return devices, nil
}

func (c *Controller) devicesTokens(devices []*yeelight.Yeelight, sess *session.Session) map[string][]*yeelight.Yeelight {
	if len(sess.DeviceTokens) > 0 {
		return sess.DeviceTokens
	}

	deviceTokens := make(map[string][]*yeelight.Yeelight, len(devices)*2)
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
	sess.DeviceTokens = deviceTokens

	return sess.DeviceTokens
}
