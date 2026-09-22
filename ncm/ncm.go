package ncm

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/songgao/packets/ethernet"
)

/*
NCM allows device and host to efficiently transfer one or more Ethernet frames
using a single USB transfer.
The USB transfer is formatted as a NCM Transfer Block (NTB).
*/
type ntbHeader struct {
	Signature   uint32
	HeaderLen   uint16
	SequenceNum uint16
	BlockLen    uint16
	NdpIndex    uint16
}

func (h ntbHeader) String() string {
	buf := make([]byte, 4)
	// Convert uint32 to bytes and store it in buf
	binary.LittleEndian.PutUint32(buf, h.Signature)
	return fmt.Sprintf("NTB-Header[sig:%s sighex:%x len=%d, seq=%d, blockLen=%d, NDPIndex= %d]", string(buf), h.Signature, h.HeaderLen, h.SequenceNum, h.BlockLen, h.NdpIndex)
}

type datagramPointerHeader struct {
	Signature    uint32
	Length       uint16
	NextNpdIndex uint16
}

const datagramPointerHeaderSignature = 0x304D434E

func (d datagramPointerHeader) IsValid() bool {
	return d.Signature == datagramPointerHeaderSignature
}

func (d datagramPointerHeader) String() string {
	return fmt.Sprintf("DatagramPointerHeader[len=%d, nextNdp=%d]", d.Length, d.NextNpdIndex)
}

type datagram struct {
	Index  uint16
	Length uint16
}

type NcmWrapper struct {
	targetReader io.Reader
	targetWriter io.Writer
	buf          *bytes.Buffer
	sequenceNum  uint16
	serial       string
}

const headerSignature = 0x484D434E

func NewWrapper(targetReader io.Reader, targetWriter io.Writer, serial string) *NcmWrapper {
	return &NcmWrapper{
		targetReader: targetReader,
		targetWriter: targetWriter,
		buf:          bytes.NewBuffer(nil),
		sequenceNum:  0,
		serial:       serial,
	}
}

const EtherHeaderLength = 14

func EthernetParser(datagram []byte) string {
	frame := ethernet.Frame(datagram)
	prot := ""
	if ethernet.IPv6 == frame.Ethertype() {
		prot = "(IPv6)"
	}
	return fmt.Sprintf("Ethernet(MAC) - dest:%x source:%x etherType:%x%s",
		frame.Destination(), frame.Source(), frame.Ethertype(), prot)

}

const UDP = 0x11

// https://en.wikipedia.org/wiki/List_of_IP_protocol_numbers
func iPv6Parser(packet []byte) string {
	length := binary.BigEndian.Uint16(packet[4:6])
	sourceAddressB := packet[8:24]
	destAddressB := packet[24:40]

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

	protocol := packet[6]
	prot := ""
	if protocol == UDP {
		prot = "UDP"
	} else {
		prot = fmt.Sprintf("PROTOCOL:%d", protocol)
	}
	return fmt.Sprintf("IP len:%d transport:%s source:%s dest:%s", length, prot, sourceIP, destIP)
}

func (r *NcmWrapper) ReadDatagrams() ([]ethernet.Frame, error) {
	var result []ethernet.Frame
	var h ntbHeader
	err := binary.Read(r.targetReader, binary.LittleEndian, &h)
	if err != nil {
		return result, fmt.Errorf("ReadDatagrams: reading header failed %w", err)
	}
	if h.Signature != headerSignature {
		return result, fmt.Errorf("ReadDatagrams: wrong header signature: %x", h.Signature)
	}
	const fixedHeaderSize = 12
	if h.HeaderLen < fixedHeaderSize {
		return result, fmt.Errorf("ReadDatagrams: header length %d is smaller than %d", h.HeaderLen, fixedHeaderSize)
	}
	if h.BlockLen < h.HeaderLen {
		return result, fmt.Errorf("ReadDatagrams: block length %d is smaller than header length %d", h.BlockLen, h.HeaderLen)
	}

	slog.Debug("read block", "ntbheader", h.String(), "length", h.BlockLen-h.HeaderLen)

	// Read the entire block after the fixed header. Keeping a zero-filled copy
	// of the header makes all protocol offsets relative to the start of the NTB.
	ncmTransferBlock := make([]byte, h.BlockLen)
	b, err := io.ReadFull(r.targetReader, ncmTransferBlock[fixedHeaderSize:])
	if err != nil {
		return result, fmt.Errorf("ReadDatagrams: reading block failed bytes read:%d err: %w", b, err)
	}
	usbReceiveBytes.WithLabelValues(r.serial).Add(float64(h.BlockLen))

	offset := int(h.NdpIndex)
	visited := make(map[int]struct{})
	for offset != 0 {
		if _, ok := visited[offset]; ok {
			return result, fmt.Errorf("ReadDatagrams: cyclic NDP chain at offset %d", offset)
		}
		visited[offset] = struct{}{}
		if offset < int(h.HeaderLen) || offset > len(ncmTransferBlock)-8 {
			return result, fmt.Errorf("ReadDatagrams: NDP offset %d is outside block length %d", offset, len(ncmTransferBlock))
		}

		var dh datagramPointerHeader
		if err := binary.Read(bytes.NewReader(ncmTransferBlock[offset:offset+8]), binary.LittleEndian, &dh); err != nil {
			return result, fmt.Errorf("ReadDatagrams: reading datagramPointerHeader failed %w", err)
		}
		if !dh.IsValid() {
			return result, fmt.Errorf("ReadDatagrams: datagrampointerheader invalid signature:%x", dh.Signature)
		}
		if dh.Length < 12 || (dh.Length-8)%4 != 0 {
			return result, fmt.Errorf("ReadDatagrams: invalid NDP length %d", dh.Length)
		}
		ndpEnd := offset + int(dh.Length)
		if ndpEnd < offset || ndpEnd > len(ncmTransferBlock) {
			return result, fmt.Errorf("ReadDatagrams: NDP at %d ends outside block at %d", offset, ndpEnd)
		}
		slog.Debug("datagramPointerHeader", "header", dh.String())

		terminated := false
		for pointer := offset + 8; pointer+4 <= ndpEnd; pointer += 4 {
			dgIndex := int(binary.LittleEndian.Uint16(ncmTransferBlock[pointer : pointer+2]))
			dgLen := int(binary.LittleEndian.Uint16(ncmTransferBlock[pointer+2 : pointer+4]))
			if dgIndex == 0 && dgLen == 0 {
				terminated = true
				break
			}
			if dgIndex == 0 || dgLen == 0 {
				return result, fmt.Errorf("ReadDatagrams: incomplete datagram pointer index=%d length=%d", dgIndex, dgLen)
			}
			dgEnd := dgIndex + dgLen
			if dgIndex < int(h.HeaderLen) || dgEnd < dgIndex || dgEnd > len(ncmTransferBlock) {
				return result, fmt.Errorf("ReadDatagrams: datagram range %d:%d is outside block length %d", dgIndex, dgEnd, len(ncmTransferBlock))
			}
			if dgLen < EtherHeaderLength {
				return result, fmt.Errorf("ReadDatagrams: datagram length %d is shorter than Ethernet header", dgLen)
			}

			frame := ethernet.Frame(ncmTransferBlock[dgIndex:dgEnd])
			ipv6 := ""
			if frame.Ethertype() == ethernet.IPv6 {
				if len(frame) < EtherHeaderLength+40 {
					return result, fmt.Errorf("ReadDatagrams: IPv6 datagram length %d is shorter than minimum frame", len(frame))
				}
				ipv6 = iPv6Parser(frame[EtherHeaderLength:])
			}
			slog.Debug("parse ethernet frame", "ipv6", ipv6, "ethernet", EthernetParser(frame))
			result = append(result, frame)
		}
		if !terminated {
			return result, fmt.Errorf("ReadDatagrams: NDP at %d has no terminating pointer", offset)
		}
		offset = int(dh.NextNpdIndex)
	}

	return result, nil
}

// this wants a complete ethernet.Frame on every write.
// also it's pretty inefficient atm as it packages one frame into one NTB
// it should work nevertheless, albeit a bit slower
func (r *NcmWrapper) Write(p []byte) (n int, err error) {
	blocklength := len(p) + 12 + 8 + 8 + 2
	block := make([]byte, blocklength)
	h := ntbHeader{
		Signature:   headerSignature,
		HeaderLen:   12,
		SequenceNum: r.sequenceNum,
		BlockLen:    uint16(blocklength),
		NdpIndex:    12,
	}

	dh := datagramPointerHeader{
		Signature:    datagramPointerHeaderSignature,
		Length:       16,
		NextNpdIndex: 0,
	}

	r.sequenceNum++

	buf := bytes.NewBuffer(block)
	buf.Reset()
	d := datagram{
		Index:  30,
		Length: uint16(len(p)),
	}
	d0 := datagram{
		Index:  0,
		Length: 0,
	}
	err = errors.Join(
		binary.Write(buf, binary.LittleEndian, h),
		binary.Write(buf, binary.LittleEndian, dh),
		binary.Write(buf, binary.LittleEndian, d),
		binary.Write(buf, binary.LittleEndian, d0),
		buf.WriteByte(0),
		buf.WriteByte(0),
	)
	if err != nil {
		return 0, fmt.Errorf("write: writing ncm packet to buffer failed %w", err)
	}
	buf.Write(p)
	block = buf.Bytes()
	usbSendBytes.WithLabelValues(r.serial).Add(float64(len(block)))
	n, err = r.targetWriter.Write(block)
	if err != nil {
		return n, fmt.Errorf("write: writing ncm packet to usb failed %w", err)
	}
	return n, nil
}
