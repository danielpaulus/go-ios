package dtx_test

import (
	"encoding/binary"
	"testing"

	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
	"github.com/stretchr/testify/assert"
)

func TestAck(t *testing.T) {
	msg := dtx.Message{}
	ack := dtx.BuildAckMessage(msg)

	ackMessage, _, err := dtx.DecodeNonBlocking(ack)
	assert.NoError(t, err)
	assert.Equal(t, int(dtx.DtxMessageHeaderLength+dtx.DtxMessagePayloadHeaderLength), len(ack))
	assert.Equal(t, dtx.Ack, ackMessage.PayloadHeader.MessageType)
	assert.Equal(t, msg.ConversationIndex+1, ackMessage.ConversationIndex)
}

func payloadOnly() dtx.Message {
	return dtx.Message{FragmentIndex: 1, ConversationIndex: 2, Payload: []interface{}{"test"}, Auxiliary: dtx.NewPrimitiveDictionary()}
}

func auxOnly() dtx.Message {
	aux := dtx.NewPrimitiveDictionary()
	aux.AddInt32(5)
	return dtx.Message{FragmentIndex: 1, ConversationIndex: 2, Payload: []interface{}{}, Auxiliary: aux}
}

func TestEncoder(t *testing.T) {
	type test struct {
		msg         dtx.Message
		description string
	}
	reply := payloadOnly()
	reply.ExpectsReply = true
	tests := []test{
		{payloadOnly(), "Message with Payload"},
		{reply, "Message with ExpectsReply"},
		{auxOnly(), "Auxiliary message"},
	}

	for _, tc := range tests {
		msg := tc.msg
		payload, _ := nskeyedarchiver.ArchiveBin(msg.Payload)
		msgBytes, err := dtx.Encode(msg.Identifier, msg.ConversationIndex, msg.ChannelCode, msg.ExpectsReply, msg.PayloadHeader.MessageType, payload, msg.Auxiliary)
		assert.NoError(t, err)

		decodedMessage, _, err := dtx.DecodeNonBlocking(msgBytes)
		assert.NoError(t, err)
		assert.Equal(t, msg.ChannelCode, decodedMessage.ChannelCode)
		assert.Equal(t, msg.Identifier, decodedMessage.Identifier)
		assert.Equal(t, msg.ConversationIndex, decodedMessage.ConversationIndex)
		assert.Equal(t, msg.ExpectsReply, decodedMessage.ExpectsReply)
		assert.Equal(t, msg.PayloadHeader.MessageType, decodedMessage.PayloadHeader.MessageType)

	}
}

func TestPrimitiveDictionaryAddInt64(t *testing.T) {
	const value int64 = 1<<40 + 123

	aux := dtx.NewPrimitiveDictionary()
	aux.AddInt64(value)

	encoded, err := aux.ToBytes()
	assert.NoError(t, err)
	assert.Len(t, encoded, 16)
	assert.Equal(t, uint32(0x0A), binary.LittleEndian.Uint32(encoded[0:4]))
	assert.Equal(t, uint32(0x06), binary.LittleEndian.Uint32(encoded[4:8]))
	assert.Equal(t, value, int64(binary.LittleEndian.Uint64(encoded[8:16])))

	decoded := dtx.DecodeAuxiliary(encoded)
	assert.Equal(t, []interface{}{uint64(value)}, decoded.GetArguments())
}
