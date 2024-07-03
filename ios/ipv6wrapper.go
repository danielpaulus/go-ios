package ios

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

type ipv6wrapper struct {
	conn io.ReadWriteCloser
}

func (i ipv6wrapper) Close() error {
	return i.conn.Close()
}
func (i ipv6wrapper) Read(p []byte) (n int, err error) {
	payload, err := iPv6Parser(i.conn)
	if err != nil {
		return 0, err
	}
	n = copy(p, payload)
	return n, nil
}

func (i ipv6wrapper) Write(p []byte) (n int, err error) {
	srcAddr := [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}  // ::1
	destAddr := [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1} // ::1
	packet := wrapIntoIPv6Packet(p, srcAddr, destAddr)
	n, err = i.conn.Write(packet)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func wrapipv6(rawConn io.ReadWriteCloser) io.ReadWriteCloser {
	w := ipv6wrapper{conn: rawConn}
	return w
}

func wrapIntoIPv6Packet(payload []byte, srcAddr, destAddr [16]byte) []byte {
	packet := make([]byte, 40+len(payload)) // IPv6 header is 40 bytes

	// Version (4 bits), Traffic Class (8 bits), and Flow Label (20 bits)
	packet[0] = 0x60 // Version is 6, Traffic Class and Flow Label are 0

	// Payload Length (16 bits)
	binary.BigEndian.PutUint16(packet[4:6], uint16(len(payload)))

	// Next Header (8 bits), set to 59 (No Next Header)
	packet[6] = 59

	// Hop Limit (8 bits), set to 64
	packet[7] = 64

	// Source Address (128 bits)
	copy(packet[8:24], srcAddr[:])

	// Destination Address (128 bits)
	copy(packet[24:40], destAddr[:])

	// Payload
	copy(packet[40:], payload)

	return packet
}

const UDP = 0x11

// https://en.wikipedia.org/wiki/List_of_IP_protocol_numbers
func iPv6Parser(stream io.Reader) ([]byte, error) {
	buf := make([]byte, 66000)
	// magic header and flags
	stream.Read(buf[:40])
	fmt.Printf("magic header and flags:%x\n", buf[:4])
	if (buf[0] & 0xF0) == 4 {
		print("dropping ipv4")
		length := binary.BigEndian.Uint16(buf[2:4])
		print(length)
		print("\n")
		stream.Read(buf[:length-4])
	}
	// length

	length := binary.BigEndian.Uint16(buf[4:6])

	// Combine the bytes into a single 32-bit value.
	combined := uint32(buf[0])<<24 | uint32(buf[1])<<16 | uint32(buf[2])<<8 | uint32(buf[3])
	trafficClass := (combined >> 20) & 0xFF

	fmt.Printf("Traffic Class: %X\n", trafficClass)
	// Mask out the first 12 bits (version and traffic class) and keep the last 20 bits (flow label).
	flowLabel := combined & 0x000FFFFF

	fmt.Printf("Flow Label: %X\n", flowLabel)
	//protocol, like TCP 0x06
	fmt.Printf("next header %x\n", buf[6])

	// TTL, can be anything
	fmt.Printf("hop limit %x\n", buf[7])
	// next header, hop limit

	sourceAddressB := buf[8:24]
	destAddressB := buf[24:40]

	var hexStrings []string
	for _, b := range sourceAddressB {
		hexStrings = append(hexStrings, fmt.Sprintf("%02X", b))
	}

	sourceIP := strings.Join(hexStrings, ":")

	var hexStrings1 []string
	for _, b := range destAddressB {
		hexStrings1 = append(hexStrings1, fmt.Sprintf("%02X", b))
	}
	destIP := strings.Join(hexStrings1, ":")

	protocol := buf[6]
	prot := ""
	if protocol == UDP {
		prot = "UDP"
	} else {
		prot = fmt.Sprintf("PROTOCOL:%d", protocol)
	}
	stream.Read(buf[:length])
	fmt.Sprintf("IP len:%d transport:%s source:%s dest:%s", length, prot, sourceIP, destIP)
	return buf[:length], nil
}
