package ncm

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/songgao/packets/ethernet"
)

/*
NCM allows device and host to efficiently transfer one or more Ethernet frames
using a single USB trans- fer.
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
}

const headerSignature = 0x484D434E

func NewWrapper(targetReader io.Reader, targetWriter io.Writer) *NcmWrapper {
	return &NcmWrapper{
		targetReader: targetReader,
		targetWriter: targetWriter,
		buf:          bytes.NewBuffer(nil),
		sequenceNum:  0,
	}
}

const EtherHeaderLength = 14
const IPv6 = 0x86DD

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
func IPv6Parser(packet []byte) string {
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
		return result, err
	}
	if h.Signature != headerSignature {
		fmt.Printf("%x%x%x%x%x\n", h.Signature, h.HeaderLen, h.SequenceNum, h.BlockLen, h.NdpIndex)
		test := make([]byte, 4000)
		n, err := r.targetReader.Read(test)

		if err == nil {
			fmt.Printf("%x\n", test)
		} else {
			println(err)
		}
		test = make([]byte, 4000)
		n, err = r.targetReader.Read(test)

		if err == nil {
			fmt.Printf("%x\n", test)
		} else {
			println(err)
		}
		return result, fmt.Errorf("wrong header signature: %x, read %d additional", h.Signature, n)
	}
	fmt.Printf("%s, read block: %d\n", h.String(), h.BlockLen-h.HeaderLen)

	//read the entire block, minus the header
	ncmTransferBlock := make([]byte, h.BlockLen)

	//later we need many indexes, so we pad the header length with 0s for easier calculations
	b, err := io.ReadFull(r.targetReader, ncmTransferBlock[h.HeaderLen:])
	if err != nil {
		return result, fmt.Errorf("reading block failed bytes read:%d err: %w", b, err)
	}
	//fmt.Printf("block: %x\n", ncmTransferBlock)

	offset := h.NdpIndex
	var dh datagramPointerHeader
	err = binary.Read(bytes.NewReader(ncmTransferBlock[offset:]), binary.LittleEndian, &dh)
	if err != nil {
		return result, err
	}
	if !dh.IsValid() {
		return result, fmt.Errorf("datagrampointerheader invalid signature:%x", dh.Signature)
	}
	fmt.Printf("datagramPointerHeader: %s\n", dh.String())
	if dh.NextNpdIndex != 0 {
		//if this happens, we gotta create a loop here to extract all dhs, starting with the next index until
		//nextndpindex==0
		panic("not implemented :-)")
	}
	datagramPointers := ncmTransferBlock[offset+8:]
	pointer := 0
	for {
		dgIndex := binary.LittleEndian.Uint16(datagramPointers[pointer:])
		dgLen := binary.LittleEndian.Uint16(datagramPointers[pointer+2:])
		if dgLen == 0 {
			break
		}
		slog.Debug("datagram", "index", dgIndex, "length", dgLen)
		datagram := ncmTransferBlock[dgIndex : dgIndex+dgLen]
		fmt.Printf("%s\n%s \n", IPv6Parser(datagram[EtherHeaderLength:]), EthernetParser(datagram))
		result = append(result, ethernet.Frame(datagram))
		pointer += 4
		if pointer > int(dh.Length-8) {
			slog.Error("datagramheaderpointer out of bounds")
			break
		}
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

	binary.Write(buf, binary.LittleEndian, h)
	binary.Write(buf, binary.LittleEndian, dh)

	d := datagram{
		Index:  30,
		Length: uint16(len(p)),
	}
	d0 := datagram{
		Index:  0,
		Length: 0,
	}
	binary.Write(buf, binary.LittleEndian, d)
	binary.Write(buf, binary.LittleEndian, d0)
	buf.WriteByte(0)
	buf.WriteByte(0)

	buf.Write(p)
	block = buf.Bytes()
	n, err = r.targetWriter.Write(block)
	fmt.Printf("%x\n", block)
	/*
		//just for debugging
		var tw = NewWrapper(bytes.NewReader(buf.Bytes()), io.Discard)
		frame, err := tw.ReadDatagrams()
		if err != nil {
			slog.Error("failed aprsing ", "err", err)
		} else {
			fmt.Print("sending:")
			fmt.Print(EthernetParser(frame[0]))
		}*/

	return n, err
}
