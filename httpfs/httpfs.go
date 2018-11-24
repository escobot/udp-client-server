package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	minLen = 11
	maxLen = 1024
)

// Packet represents a simulated network packet.
type Packet struct {
	// Type is the type of the packet which is either ACK or DATA (1 byte).
	Type uint8
	// SeqNum is the sequence number of the packet. It's 4 bytes in BigEndian format.
	SeqNum uint32
	// ToAddr is the destination address of the packet.
	// It include 4 bytes for IPv6 and 2 bytes in BigEndian for port number.
	ToAddr *net.UDPAddr
	// FromAddr is the address of the sender. It's not included in the raw data.
	// It's inferred from the recvFrom method.
	FromAddr *net.UDPAddr
	// Payload is the real data of the packet.
	Payload []byte
}

// Raw returns the raw representation of the packet is to be sent in BigEndian.
func (p Packet) Raw() []byte {
	var buf bytes.Buffer
	append := func(data interface{}) {
		binary.Write(&buf, binary.BigEndian, data)
	}
	append(p.Type)
	append(p.SeqNum)

	// Swap the peer value from ToAddr to FromAddr; and uses 4bytes version.
	append(p.FromAddr.IP.To4())
	append(uint16(p.FromAddr.Port))

	append(p.Payload)
	return buf.Bytes()
}

func (p Packet) String() string {
	return fmt.Sprintf("#%d, %s -> %s, sz=%d", p.SeqNum, p.FromAddr, p.ToAddr, len(p.Payload))
}

// parsePacket extracts, validates and creates a packet from a slice of bytes.
func parsePacket(fromAddr *net.UDPAddr, data []byte) (*Packet, error) {
	if len(data) < minLen {
		return nil, fmt.Errorf("packet is too short: %d bytes", len(data))
	}
	if len(data) > maxLen {
		return nil, fmt.Errorf("packet is exceeded max length: %d bytes", len(data))
	}
	curr := 0
	next := func(n int) []byte {
		bs := data[curr : curr+n]
		curr += n
		return bs
	}
	u16, u32 := binary.BigEndian.Uint16, binary.BigEndian.Uint32
	p := Packet{}
	p.Type = next(1)[0]
	p.SeqNum = u32(next(4))
	toAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", net.IP(next(4)), u16(next(2))))
	// If toAddr is loopback, it should be as same as the host of fromAddr.
	if toAddr.IP.IsLoopback() {
		toAddr.IP = fromAddr.IP
	}
	p.ToAddr = fromAddr
	p.FromAddr = toAddr
	p.Payload = data[curr:]
	return &p, err
}

// send sends the packet the associated destination of the packet.
func send(conn *net.UDPConn, p Packet) {
	if _, err := conn.WriteToUDP(p.Raw(), p.ToAddr); err != nil {
		logger.Printf("failed to deliver %s: %v\n", p, err)
		return
	}
	logger.Printf("packet %s is delivered\n", p)
}

var (
	verbose   = flag.Bool("v", false, "Prints debugging messages.")
	directory = flag.String("d", ".", "Specifies the directory that the server will use to read/write requested files.")
	port      = flag.Int("p", 8080, "Specifies the port number that the server will listen and serve at.")
)

var logger *log.Logger

func init() {
	logf, err := os.OpenFile("httpfs.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open log file: %v", err)
		panic(err)
	}
	logger = log.New(io.MultiWriter(logf, os.Stderr), "", log.Ltime|log.Lmicroseconds)
}

func main() {
	flag.Parse()

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", *port))
	if err != nil {
		logger.Fatalln("failed to resolve address:", err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		logger.Fatalln("failed to listen on port", err)
	}
	defer conn.Close()
	logger.Println("httpfs is listening at", addr)

	if *verbose {
		logger.Println("httpfs is running in verbose mode")
	}

	for {
		buf := make([]byte, 2048)
		n, fromAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			logger.Println("failed to receive message:", err)
			continue
		}
		go func(conn *net.UDPConn, fromAddr *net.UDPAddr, buf []byte, n int) {
			p, err := parsePacket(fromAddr, buf[:n])
			if err != nil {
				logger.Println("invalid packet:", err)
			}
			// ACK
			if p.Type == 0 {
				req, err := ParseRequest(string(p.Payload))
				if err != nil {
					logger.Println("Error parsing the request:", err)
				}
				resp := createResponse(req)
				if resp == nil {
					logger.Println("Error creating the response")
				}
				p.Payload = resp
				send(conn, *p)

				// syn
			} else if p.Type == 1 {
				p.Type = 2
				fmt.Println(p.Type)
				send(conn, *p)
			}
		}(conn, fromAddr, buf, n)
	}
}

func createResponse(req *Request) (payload []byte) {
	res := NewResponse()
	filepath := *directory + req.Path

	if strings.Contains(filepath, "/..") {
		return res.Send(400, "BAD REQUEST: do not have access to other directories\r\n", "")
	}
	if req.Method == "GET" {
		if req.Path == "/" {
			files, err := readDirectory(*directory)
			if err != nil {
				log.Printf("Error could not read directory: %v", err)
				return
			}
			return res.Send(200, strings.Join(files, "\r\n")+"\r\n", "")
		}
		return res.Send(200, "", filepath)
	} else if req.Method == "POST" {
		if req.Path == "/" {
			return res.Send(400, "BAD REQUEST: need to pick filename\r\n", "")
		}

		if _, val := req.Headers["Content-Length"]; !val {
			return res.Send(400, "Content-Length header is required", "")
		}

		l, err := strconv.Atoi(req.Headers["Content-Length"])
		if err != nil {
			log.Printf("Error could not read content-length: %v. value: %v", err, req.Headers["Content-Length"])
			return
		}

		f, err := os.Create(filepath)
		if err != nil {
			log.Printf("Error could not open file %s for writing: %v", req.Path[1:], err)
			return
		}
		defer f.Close()

		var r io.Reader
		if *verbose {
			r = io.TeeReader(req, os.Stdout)
		} else {
			r = req
		}

		if _, err = io.CopyN(f, r, int64(l)); err != nil {
			log.Printf("Error writing to file: %v", err)
			return
		}

		return res.SendStatus(200)
	}
	return nil
}

func readDirectory(d string) ([]string, error) {
	files, err := ioutil.ReadDir(d)
	if err != nil {
		return nil, fmt.Errorf("Error reading the directory %v", err)
	}
	fileList := []string{}
	for _, file := range files {
		if !file.IsDir() {
			fileList = append(fileList, file.Name())
		}
	}
	return fileList, nil
}
