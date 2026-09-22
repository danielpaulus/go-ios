package pcap

import (
	"encoding/binary"
	"math"
	"testing"
)

func TestGetPacketRejectsTruncatedHeader(t *testing.T) {
	data := make([]byte, PacketHeaderSize-1)
	if _, _, err := getPacket(data); err == nil {
		t.Fatal("truncated fixed header was accepted")
	}
}

func TestGetPacketRejectsInvalidHeaderSizes(t *testing.T) {
	for _, headerSize := range []uint32{PacketHeaderSize - 1, PacketHeaderSize + 1, math.MaxUint32} {
		data := make([]byte, PacketHeaderSize)
		binary.BigEndian.PutUint32(data, headerSize)
		if _, _, err := getPacket(data); err == nil {
			t.Errorf("header size %d was accepted for %d-byte input", headerSize, len(data))
		}
	}
}

func TestGetPacketSkipsBoundedExtendedHeader(t *testing.T) {
	const extensionSize = 8
	data := make([]byte, int(PacketHeaderSize)+extensionSize+3)
	binary.BigEndian.PutUint32(data, PacketHeaderSize+extensionSize)
	copy(data[int(PacketHeaderSize)+extensionSize:], []byte{1, 2, 3})

	_, packet, err := getPacket(data)
	if err != nil {
		t.Fatalf("getPacket: %v", err)
	}
	// FramePreLength defaults to zero, so getPacket prepends a 14-byte Ethernet header.
	if got := packet[len(packet)-3:]; got[0] != 1 || got[1] != 2 || got[2] != 3 {
		t.Fatalf("packet suffix = %v, want [1 2 3]", got)
	}
}
