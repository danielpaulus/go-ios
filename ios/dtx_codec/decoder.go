package dtx

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	log "github.com/sirupsen/logrus"

	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
)

// ReadMessage uses the reader to fully read a Message from it in blocking mode.
func ReadMessage(reader io.Reader) (Message, error) {
	header := make([]byte, 32)
	_, err := io.ReadFull(reader, header)
	if err != nil {
		return Message{}, err
	}
	if binary.BigEndian.Uint32(header) != DtxMessageMagic {
		return Message{}, NewOutOfSync(fmt.Sprintf("Wrong Magic: %x", header[0:4]))
	}
	result := readHeader(header)

	if result.IsFragment() {
		// the first part of a fragmented message is only a header indicating the total length of
		// the defragmented message
		if result.IsFirstFragment() {
			// put in the header as bytes here
			result.fragmentBytes = header
			return result, nil
		}
		// 32 offset is correct, the binary starts with a payload header
		messageBytes := make([]byte, result.MessageLength)
		_, err := io.ReadFull(reader, messageBytes)
		if err != nil {
			return Message{}, err
		}
		result.fragmentBytes = messageBytes
		return result, nil
	}

	payloadHeaderBytes := make([]byte, 16)
	_, err = io.ReadFull(reader, payloadHeaderBytes)
	if err != nil {
		return Message{}, err
	}

	ph, err := parsePayloadHeader(payloadHeaderBytes)
	if err != nil {
		return Message{}, err
	}
	result.PayloadHeader = ph

	if result.HasAuxiliary() {
		auxHeaderBytes := make([]byte, 16)
		_, err = io.ReadFull(reader, auxHeaderBytes)
		if err != nil {
			return Message{}, err
		}

		header, err := parseAuxiliaryHeader(auxHeaderBytes)
		if err != nil {
			return Message{}, err
		}
		result.AuxiliaryHeader = header
		auxBytes := make([]byte, result.AuxiliaryHeader.AuxiliarySize)
		_, err = io.ReadFull(reader, auxBytes)
		if err != nil {
			return Message{}, err
		}
		result.Auxiliary = DecodeAuxiliary(auxBytes)
	}

	result.RawBytes = make([]byte, 0)
	if result.HasPayload() {
		payloadBytes := make([]byte, result.PayloadLength())
		_, err := io.ReadFull(reader, payloadBytes)
		if err != nil {
			return Message{}, err
		}

		payload, err := nskeyedarchiver.Unarchive(payloadBytes)
		if err != nil {
			return Message{}, err
		}
		result.Payload = payload
	}

	return result, nil
}

// DecodeNonBlocking should only be used for the debug proxy to on the fly decode DtxMessages.
// It is used because if the Decoder encounters an error, we can still keep reading and forwarding the raw bytes.
// This ensures that the debug proxy keeps working and the byte dump can be used to fix the DtxDecoder
func DecodeNonBlocking(messageBytes []byte) (Message, []byte, error) {
	if len(messageBytes) < 4 {
		return Message{}, make([]byte, 0), NewIncomplete("Less than 4 bytes")
	}

	if binary.BigEndian.Uint32(messageBytes) != DtxMessageMagic {
		return Message{}, make([]byte, 0), NewOutOfSync(fmt.Sprintf("Wrong Magic: %x", messageBytes[0:4]))
	}

	if len(messageBytes) < 32 {
		return Message{}, make([]byte, 0), NewIncomplete("Less than 32 bytes")
	}

	if binary.LittleEndian.Uint32(messageBytes[4:]) != DtxMessageHeaderLength {
		return Message{}, make([]byte, 0), fmt.Errorf("Incorrect Header length, should be 32: %x", messageBytes[4:8])
	}

	result := readHeader(messageBytes)
	if result.IsFirstFragment() {
		result.fragmentBytes = messageBytes[:32]
		return result, messageBytes[32:], nil
	}
	if result.IsFragment() {
		// 32 offset is correct, the binary starts with a payload header
		if len(messageBytes) < result.MessageLength+32 {
			return Message{}, make([]byte, 0), NewIncomplete("Fragment lacks bytes")
		}
		result.fragmentBytes = messageBytes[32 : result.MessageLength+32]
		return result, messageBytes[result.MessageLength+32:], nil
	}

	if len(messageBytes) < 48 {
		return Message{}, make([]byte, 0), NewIncomplete("Payload Header missing")
	}

	ph, err := parsePayloadHeader(messageBytes[32:48])
	if err != nil {
		return Message{}, make([]byte, 0), err
	}
	result.PayloadHeader = ph

	if result.HasAuxiliary() {
		if len(messageBytes) < 64 {
			return Message{}, make([]byte, 0), NewIncomplete("Aux Header missing")
		}
		header, err := parseAuxiliaryHeader(messageBytes[48:64])
		if err != nil {
			return Message{}, make([]byte, 0), err
		}
		result.AuxiliaryHeader = header
		if len(messageBytes) < int(48+result.PayloadHeader.AuxiliaryLength) {
			return Message{}, make([]byte, 0), NewIncomplete("Aux Payload missing")
		}
		auxBytes := messageBytes[64 : 48+result.PayloadHeader.AuxiliaryLength]
		result.Auxiliary = DecodeAuxiliary(auxBytes)
	}

	totalMessageLength := result.MessageLength + int(DtxMessageHeaderLength)
	if len(messageBytes) < totalMessageLength {
		return Message{}, make([]byte, 0), NewIncomplete("Payload missing")
	}
	result.RawBytes = messageBytes[:totalMessageLength]

	if result.HasPayload() {
		payload, err := result.parsePayloadBytes(result.RawBytes)
		if err != nil {
			return Message{}, make([]byte, 0), err
		}
		result.Payload = payload
	}

	remainingBytes := messageBytes[totalMessageLength:]
	return result, remainingBytes, nil
}

func readHeader(messageBytes []byte) Message {
	result := Message{}
	result.FragmentIndex = binary.LittleEndian.Uint16(messageBytes[8:])
	result.Fragments = binary.LittleEndian.Uint16(messageBytes[10:])
	result.MessageLength = int(binary.LittleEndian.Uint32(messageBytes[12:]))
	result.Identifier = int(binary.LittleEndian.Uint32(messageBytes[16:]))
	result.ConversationIndex = int(binary.LittleEndian.Uint32(messageBytes[20:]))
	result.ChannelCode = int(binary.LittleEndian.Uint32(messageBytes[24:]))

	result.ExpectsReply = binary.LittleEndian.Uint32(messageBytes[28:]) == uint32(1)
	return result
}

func parseAuxiliaryHeader(headerBytes []byte) (AuxiliaryHeader, error) {
	r := bytes.NewReader(headerBytes)
	var result AuxiliaryHeader
	err := binary.Read(r, binary.LittleEndian, &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func parsePayloadHeader(messageBytes []byte) (PayloadHeader, error) {
	result := PayloadHeader{}
	result.MessageType = MessageType(binary.LittleEndian.Uint32(messageBytes))
	result.AuxiliaryLength = binary.LittleEndian.Uint32(messageBytes[4:])
	result.TotalPayloadLength = binary.LittleEndian.Uint32(messageBytes[8:])
	result.Flags = binary.LittleEndian.Uint32(messageBytes[12:])

	return result, nil
}

func (d Message) parsePayloadBytes(messageBytes []byte) ([]interface{}, error) {
	offset := uint32(0)
	if d.HasAuxiliary() && d.HasPayload() {
		offset = 48 + d.PayloadHeader.AuxiliaryLength
	}
	if !d.HasAuxiliary() && d.HasPayload() {
		offset = 48
	}
	if d.PayloadHeader.MessageType == UnknownTypeOne {
		return []interface{}{messageBytes[offset:]}, nil
	}
	if d.PayloadHeader.MessageType == LZ4CompressedMessage {
		uncompressed, err := Decompress(messageBytes[offset:])
		if err == nil {
			log.Infof("lz4 compressed %d bytes/ %d uncompressed ", len(messageBytes[offset:]), len(uncompressed))
		} else {
			log.Infof("skipping lz4 compressed msg with %d bytes, decompression error %v", len(messageBytes[offset:]), err)
		}
		return []interface{}{messageBytes[offset:]}, nil
	}
	return nskeyedarchiver.Unarchive(messageBytes[offset:])
}

// PayloadLength equals PayloadHeader.TotalPayloadLength - d.PayloadHeader.AuxiliaryLength so it is the Payload without the Auxiliary
func (d Message) PayloadLength() uint32 {
	return d.PayloadHeader.TotalPayloadLength - d.PayloadHeader.AuxiliaryLength
}

// HasAuxiliary returns PayloadHeader.AuxiliaryLength > 0
func (d Message) HasAuxiliary() bool {
	return d.PayloadHeader.AuxiliaryLength > 0
}

// HasPayload returns PayloadLength() > 0, it is true if the Message has payload bytes
func (d Message) HasPayload() bool {
	return d.PayloadLength() > 0
}

// IsFirstFragment returns true if the message is the first of a series of fragments.IsFirstFragment
// The first fragment message is only 32 bytes long
func (d Message) IsFirstFragment() bool {
	return d.Fragments > 1 && d.FragmentIndex == 0
}

// IsLastFragment returns true if this message is the last fragment
func (d Message) IsLastFragment() bool {
	return d.Fragments > 1 && d.Fragments-d.FragmentIndex == 1
}

// IsFragment returns true if the Message is a fragment
func (d Message) IsFragment() bool {
	return d.Fragments > 1
}

// MessageIsFirstFragmentFor indicates whether the message you call this on, is the first part of a fragmented message, and if otherMessage is a subsequent fragment
func (d Message) MessageIsFirstFragmentFor(otherMessage Message) bool {
	if !d.IsFirstFragment() {
		return false
	}
	return d.Identifier == otherMessage.Identifier && d.Fragments == otherMessage.Fragments && otherMessage.FragmentIndex > 0
}
