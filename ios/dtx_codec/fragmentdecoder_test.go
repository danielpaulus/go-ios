package dtx_test

import (
	"encoding/binary"
	"testing"

	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/stretchr/testify/assert"
)

func payload() []byte {
	// bytes, _ := nskeyedarchiver.ArchiveBin("test")
	return []byte("payload")
}

func createHeader(fragmentIndex uint16, fragmentLength uint16, identifier uint32, length uint32, payload []byte) dtx.Message {
	payloadLength := len(payload)
	messageBytes := make([]byte, 32+payloadLength)
	binary.BigEndian.PutUint32(messageBytes, dtx.DtxMessageMagic)
	binary.LittleEndian.PutUint32(messageBytes[4:], dtx.DtxMessageHeaderLength)
	binary.LittleEndian.PutUint16(messageBytes[8:], fragmentIndex)
	binary.LittleEndian.PutUint16(messageBytes[10:], fragmentLength)
	binary.LittleEndian.PutUint32(messageBytes[12:], length)
	binary.LittleEndian.PutUint32(messageBytes[16:], identifier)
	binary.LittleEndian.PutUint32(messageBytes[20:], uint32(0))
	binary.LittleEndian.PutUint32(messageBytes[24:], uint32(0))
	if payloadLength != 0 {
		copy(messageBytes[32:], payload)
	}
	msg, _, err := dtx.DecodeNonBlocking(messageBytes)
	if err != nil {
		panic(err)
	}
	return msg
}

func createFragmentedMessage(identifier uint32) (dtx.Message, dtx.Message, dtx.Message, string) {
	payload := []byte("payload")
	firstFrag := createHeader(0, 3, identifier, uint32(len(payload)), make([]byte, 0))
	secondFrag := createHeader(1, 3, identifier, 4, payload[:4])
	thirdFrag := createHeader(2, 3, identifier, 3, payload[4:])
	return firstFrag, secondFrag, thirdFrag, string(payload)
}

func otherFrag() dtx.Message {
	return dtx.Message{FragmentIndex: 1, Fragments: 3, Identifier: 3}
}

func TestDefragmentation(t *testing.T) {
	identifier := 5
	firstFrag, secondFrag, thirdFrag, payload := createFragmentedMessage(uint32(identifier))
	decoder := dtx.NewFragmentDecoder(firstFrag)
	assert.False(t, decoder.AddFragment(otherFrag()))
	assert.False(t, decoder.HasFinished())

	assert.True(t, decoder.AddFragment(secondFrag))
	assert.False(t, decoder.HasFinished())

	assert.True(t, decoder.AddFragment(thirdFrag))
	assert.True(t, decoder.HasFinished())
	defragmentedMsg := decoder.Extract()
	// we cannot use regular message decoding here because the payload is kind of fake
	// usually it would start with a payloadheader. All these methods are private though
	// and I only want to test defragmenting properly anyway with this test
	// The defragmenter should correctly merge the payloads of all fragments and prepend
	// the header of the first message to it. Then it needs to set fragment index to 0 and length to
	// 1 so we get a de-fragmented message.
	defragmentedFragIndex := binary.LittleEndian.Uint16(defragmentedMsg[8:])
	defragmentedFragLength := binary.LittleEndian.Uint16(defragmentedMsg[10:])
	defragmentedLength := binary.LittleEndian.Uint32(defragmentedMsg[12:])
	defragmentedIdentifier := binary.LittleEndian.Uint32(defragmentedMsg[16:])

	assert.Equal(t, payload, string(defragmentedMsg[dtx.DtxMessageHeaderLength:]))
	assert.Equal(t, uint16(0), defragmentedFragIndex)
	assert.Equal(t, uint16(1), defragmentedFragLength)

	assert.Equal(t, uint32(len(payload)), defragmentedLength)
	assert.Equal(t, uint32(identifier), defragmentedIdentifier)
}

func TestFragmentPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	_, secondFrag, _, _ := createFragmentedMessage(4)

	dtx.NewFragmentDecoder(secondFrag)
}

func TestExtractPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	firstFrag, _, _, _ := createFragmentedMessage(3)
	decoder := dtx.NewFragmentDecoder(firstFrag)
	decoder.Extract()
}
