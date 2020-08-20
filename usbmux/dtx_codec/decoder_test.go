package dtx_test

import (
	"io/ioutil"
	"log"
	"testing"

	dtx "github.com/danielpaulus/go-ios/usbmux/dtx_codec"
	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {
	dat := make([]byte, 5)
	_, _, err := dtx.Decode(dat)
	assert.True(t, dtx.IsOutOfSync(err))

	dat = make([]byte, 2)
	_, _, err = dtx.Decode(dat)
	assert.True(t, dtx.IsIncomplete(err))

	dat, err = ioutil.ReadFile("fixtures/notifyOfPublishedCapabilites")
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < len(dat)-4; i++ {
		_, _, err = dtx.Decode(dat[0 : 4+i])
		assert.True(t, dtx.IsIncomplete(err))
	}

}

func TestDecoder(t *testing.T) {
	dat, err := ioutil.ReadFile("fixtures/notifyOfPublishedCapabilites")
	if err != nil {
		log.Fatal(err)
	}
	msg, remainingBytes, err := dtx.Decode(dat)
	if assert.NoError(t, err) {
		assert.Equal(t, 0, len(remainingBytes))
		assert.Equal(t, msg.Fragments, uint16(1))
		assert.Equal(t, msg.FragmentIndex, uint16(0))
		assert.Equal(t, msg.MessageLength, 612)
		assert.Equal(t, 0, msg.ChannelCode)
		assert.Equal(t, false, msg.ExpectsReply)
		assert.Equal(t, 2, msg.Identifier)
		assert.Equal(t, 0, msg.ChannelCode)

		assert.Equal(t, 2, msg.PayloadHeader.MessageType)
		assert.Equal(t, 425, msg.PayloadHeader.AuxiliaryLength)
		assert.Equal(t, 596, msg.PayloadHeader.TotalPayloadLength)
		assert.Equal(t, 0, msg.PayloadHeader.Flags)

	}

}
