package dtx

import (
	"encoding/binary"
)

// BuildAckMessage creates a 32+ 16 byte long message that can be used as a response for a message
// the had the ExpectsReply flag set.
func BuildAckMessage(msg Message) []byte {
	response := make([]byte, 48)
	writeHeader(response, 16, msg.Identifier, msg.ConversationIndex+1, msg.ChannelCode, false)
	binary.LittleEndian.PutUint32(response[32:], uint32(Ack))
	binary.LittleEndian.PutUint32(response[36:], 0)
	binary.LittleEndian.PutUint32(response[40:], 0)
	binary.LittleEndian.PutUint32(response[44:], 0)
	return response
}

// Encode encodes the given parameters to a DtxMessage that can be sent to the device.
func Encode(
	Identifier int,
	ConversationIndex int,
	ChannelCode int,
	ExpectsReply bool,
	MessageType MessageType,
	payloadBytes []byte,
	Auxiliary PrimitiveDictionary,
) ([]byte, error) {
	auxBytes, err := Auxiliary.ToBytes()
	if err != nil {
		return make([]byte, 0), err
	}

	payloadLength := len(payloadBytes)
	auxiliarySize := len(auxBytes)
	auxHeaderSize := uint32(16)
	messageLength := 16 + uint32(auxiliarySize+payloadLength)
	if auxiliarySize > 0 {
		messageLength += auxHeaderSize
	}
	messageBytes := make([]byte, 32+messageLength)

	writeHeader(messageBytes, messageLength, Identifier, ConversationIndex, ChannelCode, ExpectsReply)
	writePayloadHeader(messageBytes[32:], MessageType, payloadLength, auxiliarySize)
	if auxiliarySize == 0 {
		copy(messageBytes[48:], payloadBytes)
	} else {
		writeAuxHeader(messageBytes[48:], auxiliarySize)
		copy(messageBytes[64:], auxBytes)
		copy(messageBytes[64+auxiliarySize:], payloadBytes)
	}

	return messageBytes, nil
}

func writeHeader(messageBytes []byte, messageLength uint32, Identifier int, ConversationIndex int,
	ChannelCode int,
	ExpectsReply bool,
) {
	binary.BigEndian.PutUint32(messageBytes, DtxMessageMagic)
	binary.LittleEndian.PutUint32(messageBytes[4:], DtxMessageHeaderLength)
	binary.LittleEndian.PutUint16(messageBytes[8:], 0)
	binary.LittleEndian.PutUint16(messageBytes[10:], 1)
	binary.LittleEndian.PutUint32(messageBytes[12:], uint32(messageLength))
	binary.LittleEndian.PutUint32(messageBytes[16:], uint32(Identifier))
	binary.LittleEndian.PutUint32(messageBytes[20:], uint32(ConversationIndex))
	binary.LittleEndian.PutUint32(messageBytes[24:], uint32(ChannelCode))
	var expectsReplyUint32 uint32
	if ExpectsReply {
		expectsReplyUint32 = 1
	} else {
		expectsReplyUint32 = 0
	}
	binary.LittleEndian.PutUint32(messageBytes[28:], expectsReplyUint32)
}

func writePayloadHeader(messageBytes []byte, messageType MessageType, payloadLength int, auxLength int) {
	binary.LittleEndian.PutUint32(messageBytes, uint32(messageType))
	auxLengthWithHeader := uint32(auxLength)
	if auxLength > 0 {
		auxLengthWithHeader += 16
	}
	binary.LittleEndian.PutUint32(messageBytes[4:], auxLengthWithHeader)
	binary.LittleEndian.PutUint32(messageBytes[8:], uint32(payloadLength)+auxLengthWithHeader)
	binary.LittleEndian.PutUint32(messageBytes[12:], 0)
}

func writeAuxHeader(messageBytes []byte, auxiliarySize int) {
	binary.LittleEndian.PutUint32(messageBytes, uint32(496))
	binary.LittleEndian.PutUint32(messageBytes[4:], 0)
	binary.LittleEndian.PutUint32(messageBytes[8:], uint32(auxiliarySize))
	binary.LittleEndian.PutUint32(messageBytes[12:], 0)
}
