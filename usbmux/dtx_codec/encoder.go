package dtx

import (
	"encoding/binary"
)

func BuildAckMessage(msg DtxMessage) []byte {
	response := make([]byte, 48)
	writeHeader(response, 16, msg.Identifier, msg.ChannelCode, msg.ConversationIndex+1, false)
	binary.LittleEndian.PutUint32(response[32:], Ack)
	binary.LittleEndian.PutUint32(response[36:], 0)
	binary.LittleEndian.PutUint32(response[40:], 0)
	binary.LittleEndian.PutUint32(response[44:], 0)
	return response
}

func Encode(
	Identifier int,
	ChannelCode int,
	ExpectsReply bool,
	MessageType int,
	payloadBytes []byte,
	Auxiliary DtxPrimitiveDictionary,
) ([]byte, error) {

	auxBytes, err := Auxiliary.ToBytes()
	if err != nil {
		return make([]byte, 0), err
	}

	payloadLength := len(payloadBytes)
	auxiliarySize := len(auxBytes)
	messageLength := 16 + 16 + uint32(auxiliarySize+payloadLength)
	messageBytes := make([]byte, 32+messageLength)

	writeHeader(messageBytes, messageLength, Identifier, 0, ChannelCode, ExpectsReply)
	writePayloadHeader(messageBytes[32:], MessageType, payloadLength, auxiliarySize)
	writeAuxHeader(messageBytes[48:], auxiliarySize)
	copy(messageBytes[64:], auxBytes)
	copy(messageBytes[64+auxiliarySize:], payloadBytes)

	//serializedMessage := make([]byte, message.)

	return messageBytes, nil
}

func writeHeader(messageBytes []byte, messageLength uint32, Identifier int, ConversationIndex int,
	ChannelCode int,
	ExpectsReply bool) {
	binary.BigEndian.PutUint32(messageBytes, DtxMessageMagic)
	binary.LittleEndian.PutUint32(messageBytes[4:], DtxHeaderLength)
	binary.LittleEndian.PutUint16(messageBytes[8:], 0)
	binary.LittleEndian.PutUint16(messageBytes[10:], 1)
	binary.LittleEndian.PutUint32(messageBytes[12:], uint32(messageLength))
	binary.LittleEndian.PutUint32(messageBytes[16:], uint32(Identifier))
	binary.LittleEndian.PutUint32(messageBytes[20:], 0)
	binary.LittleEndian.PutUint32(messageBytes[24:], uint32(ChannelCode))
	var expectsReplyUint32 uint32
	if ExpectsReply {
		expectsReplyUint32 = 1
	} else {
		expectsReplyUint32 = 0
	}
	binary.LittleEndian.PutUint32(messageBytes[28:], expectsReplyUint32)
}

func writePayloadHeader(messageBytes []byte, MessageType int, payloadLength int, auxLength int) {
	binary.LittleEndian.PutUint32(messageBytes, uint32(MessageType))
	auxLengthWithHeader := uint32(auxLength) + 16
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
