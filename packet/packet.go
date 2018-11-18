package packet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
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
	p.FromAddr = fromAddr
	toAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", net.IP(next(4)), u16(next(2))))
	// If toAddr is loopback, it should be as same as the host of fromAddr.
	if toAddr.IP.IsLoopback() {
		toAddr.IP = fromAddr.IP
	}
	p.ToAddr = toAddr
	p.Payload = data[curr:]
	return &p, err
}
