package videohub

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Videohub struct {
	ip              string
	conn            net.Conn
	logger          *log.Logger
	readerThread    *sync.WaitGroup
	protocolVersion string // Videohub Ethernet Protocol Version (ex. '2.7')
	model           string // Model of Videohub (ex. 'Blackmagic Smart Videohub 20 x 20')
	uniqueID        string // Generated unique identifier for each Videohub, persists across boots and network changes. (ex. '7C2E0DA4BFC0' )
	inputs          int    // Number of Video Inputs (sources)
	outputs         int    // Number of Video Outputs (destinations)
	inputLabels     []string
	outputLabels    []string
	routing         []int
}

func NewVideohub(ip string) *Videohub {
	vh := &Videohub{
		ip:     ip,
		logger: log.New(os.Stderr, "", log.LstdFlags),
	}
	vh.connect()
	vh.readerThread = &sync.WaitGroup{}
	vh.readerThread.Add(1)
	go vh.reader()
	return vh
}

func (vh *Videohub) connect() {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:9990", vh.ip))
	if err != nil {
		vh.logger.Fatalf("Failed to connect to Videohub: %v", err)
	}
	vh.conn = conn
}

func (vh *Videohub) reader() {
	defer vh.readerThread.Done()
	reader := bufio.NewReader(vh.conn)
	for {
		message, err := reader.ReadBytes('\n')
		if err != nil {
			vh.logger.Printf("Error reading from Videohub: %v", err)
			vh.reconnect()
			continue
		}
		messageStr := string(message)
		if strings.HasSuffix(messageStr, ":\n") {
			message, err = reader.ReadBytes('\n')
			if err != nil {
				vh.logger.Printf("Error reading from Videohub: %v", err)
				vh.reconnect()
				continue
			}
			vh.decodeMessage(append(message[:len(message)-1], message...))
		} else {
			vh.decodeResponse(message[:len(message)-1])
		}
	}
}

func (vh *Videohub) reconnect() {
	vh.logger.Println("Reconnecting to Videohub...")
	vh.conn.Close()
	vh.connect()
}

func (vh *Videohub) send(command string) {
	vh.logger.Printf("Sending Message: [%s]", strings.ReplaceAll(command, "\n", "-"))
	_, err := vh.conn.Write([]byte(command + "\n\n"))
	if err != nil {
		vh.logger.Printf("Error sending command to Videohub: %v", err)
		vh.reconnect()
	}
}

func (vh *Videohub) decodeMessage(message []byte) {
	msg := strings.TrimSuffix(string(message), "\n\n")
	vh.logger.Printf("Received Message: [%s]", strings.ReplaceAll(msg, "\n", "//"))
	lines := strings.Split(msg, "\n")
	vh.responseProcessor(lines)
}

func (vh *Videohub) decodeResponse(message []byte) {
	response := string(message)
	vh.logger.Printf("Received Response: [%s]", strings.ReplaceAll(response, "\n", "//"))
}

func (vh *Videohub) responseProcessor(message []string) {
	messageType := strings.TrimSuffix(message[0], ":")
	contents := message[1:]
	switch messageType {
	case "PROTOCOL PREAMBLE":
		vh.processProtocolPreamble(contents)
	case "VIDEOHUB DEVICE":
		vh.processVideohubDevice(contents)
	case "INPUT LABELS":
		vh.processInputLabels(contents)
	case "OUTPUT LABELS":
		vh.processOutputLabels(contents)
	case "VIDEO OUTPUT LOCKS":
		// Do nothing
	case "VIDEO OUTPUT ROUTING":
		vh.processOutputRouting(contents)
	case "CONFIGURATION":
		// Do nothing
	}
}

func (vh *Videohub) processProtocolPreamble(contents []string) {
	for _, item := range contents {
		parts := strings.Split(item, ": ")
		if len(parts) == 2 {
			key, value := parts[0], parts[1]
			if key == "Version" {
				vh.protocolVersion = value
			}
		}
	}
}

func (vh *Videohub) processVideohubDevice(contents []string) {
	for _, item := range contents {
		parts := strings.Split(item, ": ")
		if len(parts) == 2 {
			key, value := parts[0], parts[1]
			switch key {
			case "Model name":
				vh.model = value
			case "Unique ID":
				vh.uniqueID = value
			case "Video inputs":
				vh.inputs = parseInt(value)
				vh.inputLabels = make([]string, vh.inputs)
			case "Video outputs":
				vh.outputs = parseInt(value)
				vh.outputLabels = make([]string, vh.outputs)
				vh.routing = make([]int, vh.outputs)
				for i := range vh.routing {
					vh.routing[i] = -1
				}
			}
		}
	}
}

func (vh *Videohub) processInputLabels(contents []string) {
	for _, item := range contents {
		parts := strings.SplitN(item, " ", 2)
		if len(parts) == 2 {
			i, label := parseInt(parts[0]), parts[1]
			vh.inputLabels[i] = label
		}
	}
}

func (vh *Videohub) processOutputLabels(contents []string) {
	for _, item := range contents {
		parts := strings.SplitN(item, " ", 2)
		if len(parts) == 2 {
			o, label := parseInt(parts[0]), parts[1]
			vh.outputLabels[o] = label
		}
	}
}

func (vh *Videohub) processOutputRouting(contents []string) {
	for _, item := range contents {
		parts := strings.Split(item, " ")
		if len(parts) == 2 {
			destination, source := parseInt(parts[0]), parseInt(parts[1])
			vh.routing[destination] = source
		}
	}
}

func (vh *Videohub) Route(destination, source int) {
	vh.send(fmt.Sprintf("VIDEO OUTPUT ROUTING:\n%d %d", destination, source))
}

func (vh *Videohub) BulkRoute(routes [][2]int) {
	command := "VIDEO OUTPUT ROUTING:"
	for _, route := range routes {
		command += fmt.Sprintf("\n%d %d", route[0], route[1])
	}
	vh.send(command)
}

func (vh *Videohub) InputLabel(source int, label string) {
	vh.send(fmt.Sprintf("INPUT LABELS:\n%d %s", source, label))
}

func (vh *Videohub) OutputLabel(destination int, label string) {
	vh.send(fmt.Sprintf("OUTPUT LABELS:\n%d %s", destination, label))
}

func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
