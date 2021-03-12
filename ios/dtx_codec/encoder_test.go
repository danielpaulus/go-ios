package dtx_test

import (
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

func TestEncoder(t *testing.T) {
	msg := dtx.Message{FragmentIndex: 1, ConversationIndex: 2}
	payload, _ := nskeyedarchiver.ArchiveBin("test")
	msgBytes, err := dtx.Encode(msg.Identifier, msg.ConversationIndex, msg.ChannelCode, msg.ExpectsReply, msg.PayloadHeader.MessageType, payload, dtx.NewPrimitiveDictionary())
	assert.NoError(t, err)

	decodedMessage, _, err := dtx.DecodeNonBlocking(msgBytes)
	assert.NoError(t, err)
	assert.Equal(t, msg.ChannelCode, decodedMessage.ChannelCode)
}
