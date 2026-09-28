package ncm

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"
)

func TestReadDatagramsRejectsMalformedBlocksWithoutPanic(t *testing.T) {
	testCases := []struct {
		name  string
		block []byte
	}{
		{name: "short header", block: []byte{1, 2, 3}},
		{name: "header length below fixed", block: makeNTB(8, 12, 0, nil)},
		{name: "block length below header", block: makeNTB(16, 12, 0, nil)},
		{name: "NDP outside block", block: makeNTB(12, 12, 12, nil)},
		{name: "short NDP", block: makeNTB(12, 16, 12, make([]byte, 4))},
		{name: "bad NDP length", block: makeNTB(12, 24, 12, ndpBytes(8, 0, nil))},
		{name: "datagram outside block", block: makeNTB(12, 28, 12, ndpBytes(16, 0, [][2]uint16{{27, 10}, {0, 0}}))},
		{name: "short Ethernet frame", block: makeNTB(12, 39, 12, append(ndpBytes(16, 0, [][2]uint16{{28, 5}, {0, 0}}), make([]byte, 11)...))},
		{name: "cyclic NDP", block: makeNTB(12, 24, 12, ndpBytes(12, 12, [][2]uint16{{0, 0}}))},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			wrapper := NewWrapper(bytes.NewReader(testCase.block), io.Discard, "test")
			if _, err := wrapper.ReadDatagrams(); err == nil {
				t.Fatal("malformed block was accepted")
			}
		})
	}
}

func TestReadDatagramsSupportsBoundedNDPChain(t *testing.T) {
	frame := make([]byte, EtherHeaderLength)
	binary.BigEndian.PutUint16(frame[12:14], 0x0800)
	first := ndpBytes(12, 24, [][2]uint16{{0, 0}})
	second := ndpBytes(16, 0, [][2]uint16{{40, uint16(len(frame))}, {0, 0}})
	body := append(append(first, second...), frame...)
	block := makeNTB(12, uint16(12+len(body)), 12, body)

	wrapper := NewWrapper(bytes.NewReader(block), io.Discard, "test")
	frames, err := wrapper.ReadDatagrams()
	if err != nil {
		t.Fatalf("ReadDatagrams: %v", err)
	}
	if len(frames) != 1 {
		t.Fatalf("frame count = %d, want 1", len(frames))
	}
}

func TestReadDatagramsValidFrame(t *testing.T) {
	frame := make([]byte, EtherHeaderLength)
	binary.BigEndian.PutUint16(frame[12:14], 0x0800)
	ndp := ndpBytes(16, 0, [][2]uint16{{28, uint16(len(frame))}, {0, 0}})
	block := makeNTB(12, uint16(28+len(frame)), 12, append(ndp, frame...))

	wrapper := NewWrapper(bytes.NewReader(block), io.Discard, "test")
	frames, err := wrapper.ReadDatagrams()
	if err != nil {
		t.Fatalf("ReadDatagrams: %v", err)
	}
	if len(frames) != 1 || len(frames[0]) != len(frame) {
		t.Fatalf("frames = %#v, want one %d-byte frame", frames, len(frame))
	}
}

func FuzzReadDatagramsNeverPanics(f *testing.F) {
	f.Add([]byte{})
	f.Add(makeNTB(12, 12, 0, nil))
	f.Fuzz(func(t *testing.T, data []byte) {
		wrapper := NewWrapper(bytes.NewReader(data), io.Discard, "fuzz")
		_, _ = wrapper.ReadDatagrams()
	})
}

func makeNTB(headerLength, blockLength, ndpIndex uint16, body []byte) []byte {
	var buffer bytes.Buffer
	_ = binary.Write(&buffer, binary.LittleEndian, ntbHeader{
		Signature: headerSignature,
		HeaderLen: headerLength,
		BlockLen:  blockLength,
		NdpIndex:  ndpIndex,
	})
	buffer.Write(body)
	return buffer.Bytes()
}

func ndpBytes(length, next uint16, pointers [][2]uint16) []byte {
	var buffer bytes.Buffer
	_ = binary.Write(&buffer, binary.LittleEndian, datagramPointerHeader{
		Signature:    datagramPointerHeaderSignature,
		Length:       length,
		NextNpdIndex: next,
	})
	for _, pointer := range pointers {
		_ = binary.Write(&buffer, binary.LittleEndian, pointer[0])
		_ = binary.Write(&buffer, binary.LittleEndian, pointer[1])
	}
	return buffer.Bytes()
}
