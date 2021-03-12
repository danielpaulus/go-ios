package dtx_test

import (
	"testing"

	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/stretchr/testify/assert"
)

func TestAck(t *testing.T) {
	msg := dtx.Message{}
	ack := dtx.BuildAckMessage(msg)

	ackMessage, _, err := dtx.DecodeNonBlocking(ack)
	assert.NoError(t, err)
	assert.Equal(t, int(dtx.DtxMessageHeaderLength+dtx.DtxMessagePayloadHeaderLength), len(ack))
	assert.Equal(t, dtx.Ack, ackMessage.PayloadHeader.MessageType)
}
