package dtx

import (
	"encoding/binary"

	"github.com/labstack/gommon/log"
)

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

	writeHeader(messageBytes, messageLength, Identifier, ChannelCode, ExpectsReply)
	writePayloadHeader(messageBytes[32:], MessageType, payloadLength, auxiliarySize)
	writeAuxHeader(messageBytes[48:], auxiliarySize)
	copy(messageBytes[64:], auxBytes)
	copy(messageBytes[64+auxiliarySize:], payloadBytes)

	//serializedMessage := make([]byte, message.)
	log.Infof("%x", messageBytes)
	return messageBytes, nil
}

func writeHeader(messageBytes []byte, messageLength uint32, Identifier int,
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
