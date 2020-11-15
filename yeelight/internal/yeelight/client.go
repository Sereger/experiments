package yeelight

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	discoverMSG = "M-SEARCH * HTTP/1.1\r\nHOST:239.255.255.250:1982\r\nMAN:\"ssdp:discover\"\r\nST:wifi_bulb\r\n"

	// timeout value for TCP and UDP commands
	timeout = time.Second * 2

	//SSDP discover address
	ssdpAddr = "239.255.255.250:1982"

	//CR-LF delimiter
	crlf = "\r\n"
)

type (
	//Command represents COMMAND request to Yeelight device
	Command struct {
		ID     int           `json:"id"`
		Method string        `json:"method"`
		Params []interface{} `json:"params"`
	}

	// CommandResult represents response from Yeelight device
	CommandResult struct {
		ID     int           `json:"id"`
		Result []interface{} `json:"result,omitempty"`
		Error  *Error        `json:"error,omitempty"`
	}

	// Notification represents notification response
	Notification struct {
		Method string            `json:"method"`
		Params map[string]string `json:"params"`
	}

	//Error struct represents error part of response
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	//Yeelight represents device
	Yeelight struct {
		Addr   string
		ID     string
		Power  string
		Name   string
		Bright int64
		rnd    *rand.Rand
	}
)

//Discover discovers device in local network via ssdp
func Discover() ([]*Yeelight, error) {
	var err error

	ssdp, err := net.ResolveUDPAddr("udp4", ssdpAddr)
	if err != nil {
		return nil, err
	}

	c, err := net.ListenPacket("udp4", ":0")
	if err != nil {
		return nil, err
	}
	defer c.Close()

	socket := c.(*net.UDPConn)
	_, err = socket.WriteToUDP([]byte(discoverMSG), ssdp)
	if err != nil {
		return nil, err
	}
	socket.SetReadDeadline(time.Now().Add(timeout))

	result := make([]*Yeelight, 0)
	rsBuf := make([]byte, 1024)
	undup := make(map[string]struct{})
	for {
		size, _, err := socket.ReadFromUDP(rsBuf)
		if err != nil {
			break
		}

		rs := rsBuf[0:size]
		y := parseDevice(string(rs))
		if y == nil {
			continue
		}

		log.Printf("Device [%+v] found", y)
		if _, ok := undup[y.Addr]; ok {
			continue
		}
		undup[y.Addr] = struct{}{}
		result = append(result, y)
	}

	return result, nil
}

//New creates new device instance for address provided
func New(addr string) *Yeelight {
	return &Yeelight{
		Addr: addr,
		rnd:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}

}

// Listen connects to device and listens for NOTIFICATION events
func (y *Yeelight) Listen() (<-chan *Notification, chan<- struct{}, error) {
	var err error
	notifCh := make(chan *Notification)
	done := make(chan struct{}, 1)

	conn, err := net.DialTimeout("tcp", y.Addr, time.Second*3)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot connect to %s. %s", y.Addr, err)
	}

	fmt.Println("Connection established")
	go func(c net.Conn) {
		//make sure connection is closed when method returns
		defer closeConnection(conn)

		connReader := bufio.NewReader(c)
		for {
			select {
			case <-done:
				return
			default:
				data, err := connReader.ReadString('\n')
				if nil == err {
					var rs Notification
					fmt.Println(data)
					json.Unmarshal([]byte(data), &rs)
					select {
					case notifCh <- &rs:
					default:
						fmt.Println("Channel is full")
					}
				}
			}

		}

	}(conn)

	return notifCh, done, nil
}

// GetProp method is used to retrieve current property of smart LED.
func (y *Yeelight) GetProp(values ...interface{}) ([]interface{}, error) {
	r, err := y.executeCommand("get_prop", values...)
	if nil != err {
		return nil, err
	}
	return r.Result, nil
}

//SetPower is used to switch on or off the smart LED (software managed on/off).
func (y *Yeelight) SetPower(on bool) error {
	var status string
	if on {
		status = "on"
	} else {
		status = "off"
	}
	_, err := y.executeCommand("set_power", status)
	return err
}

//SetPower is used to switch on or off the smart LED (software managed on/off).
func (y *Yeelight) SetBright(b int64) error {
	if b < 0 {
		b = 0
	}
	if b > 100 {
		b = 100
	}

	_, err := y.executeCommand("set_bright", b)
	return err
}

func (y *Yeelight) randID() int {
	i := y.rnd.Intn(100)
	return i
}

func (y *Yeelight) newCommand(name string, params []interface{}) *Command {
	return &Command{
		Method: name,
		ID:     y.randID(),
		Params: params,
	}
}

//executeCommand executes command with provided parameters
func (y *Yeelight) executeCommand(name string, params ...interface{}) (*CommandResult, error) {
	return y.execute(y.newCommand(name, params))
}

//executeCommand executes command
func (y *Yeelight) execute(cmd *Command) (*CommandResult, error) {

	conn, err := net.Dial("tcp", y.Addr)
	if nil != err {
		return nil, fmt.Errorf("cannot open connection to %s. %s", y.Addr, err)
	}
	time.Sleep(time.Second)
	conn.SetReadDeadline(time.Now().Add(timeout))

	//write request/command
	b, _ := json.Marshal(cmd)
	fmt.Fprint(conn, string(b)+crlf)

	//wait and read for response
	res, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("cannot read command result %s", err)
	}
	var rs CommandResult
	err = json.Unmarshal([]byte(res), &rs)
	if nil != err {
		return nil, fmt.Errorf("cannot parse command result %s", err)
	}
	if nil != rs.Error {
		return nil, fmt.Errorf("command execution error. Code: %d, Message: %s", rs.Error.Code, rs.Error.Message)
	}
	return &rs, nil
}

var respRegexp = regexp.MustCompile(`([\w_\-\d]+):\s*([^\n\r]*)\r?\n`)

//parseAddr parses address from ssdp response
func parseDevice(msg string) *Yeelight {
	result := &Yeelight{rnd: rand.New(rand.NewSource(time.Now().UnixNano()))}
	mtches := respRegexp.FindAllStringSubmatch(msg, -1)

	for _, m := range mtches {
		switch strings.ToLower(m[1]) {
		case "location":
			result.Addr = strings.TrimPrefix(m[2], "yeelight://")
		case "power":
			result.Power = m[2]
		case "bright":
			result.Bright, _ = strconv.ParseInt(m[2], 10, 64)
		case "name":
			result.Name = m[2]
		case "ID":
			result.ID = m[2]
		}
	}

	return result
}

//closeConnection closes network connection
func closeConnection(c net.Conn) {
	if nil != c {
		c.Close()
	}
}
