package dtx_test

import (
	"bytes"
	"io/ioutil"

	"testing"

	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {
	dat := make([]byte, 5)
	_, _, err := dtx.DecodeNonBlocking(dat)
	assert.True(t, dtx.IsOutOfSync(err))

	dat = make([]byte, 2)
	_, _, err = dtx.DecodeNonBlocking(dat)
	assert.True(t, dtx.IsIncomplete(err))

	dat, err = ioutil.ReadFile("fixtures/notifyOfPublishedCapabilites")
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < len(dat)-4; i++ {
		_, _, err = dtx.DecodeNonBlocking(dat[0 : 4+i])
		assert.True(t, dtx.IsIncomplete(err))
	}

}

func TestCodec2(t *testing.T) {
	dat, err := ioutil.ReadFile("fixtures/requestChannelWithCodeIdentifier.bin")

	if err != nil {
		log.Fatal(err)
	}

	fixtureMsg, _, err := dtx.DecodeNonBlocking(dat)
	if err != nil {
		log.Fatal(err)
	}
	payloadBytes, err := nskeyedarchiver.ArchiveBin(fixtureMsg.Payload[0])
	assert.NoError(t, err)
	encodedBytes, err := dtx.Encode(fixtureMsg.Identifier, fixtureMsg.ChannelCode, fixtureMsg.ExpectsReply, fixtureMsg.PayloadHeader.MessageType, payloadBytes, fixtureMsg.Auxiliary)
	if assert.NoError(t, err) {
		msg, remainingBytes, err := dtx.DecodeNonBlocking(encodedBytes)
		assert.Equal(t, 0, len(remainingBytes))
		assert.NoError(t, err)
		assert.Equal(t, fixtureMsg.Payload, msg.Payload)
	}
}

func TestCodec(t *testing.T) {

	dat, err := ioutil.ReadFile("fixtures/requestChannelWithCodeIdentifier.bin")

	if err != nil {
		log.Fatal(err)
	}
	var remainingBytes []byte
	remainingBytes = dat
	for len(remainingBytes) > 0 {
		msg, s, err := dtx.DecodeNonBlocking(remainingBytes)
		log.Info(msg.StringDebug())
		remainingBytes = s
		if !assert.NoError(t, err) {
			log.Fatal("whet", err)
		}
		bytes, err := dtx.Encode(3, 0, true, 2, msg.RawBytes[303:], msg.Auxiliary)
		if assert.NoError(t, err) {
			assert.Equal(t, dat, bytes)
		}
	}

}

func TestAXDump(t *testing.T) {

	//dat, err := ioutil.ReadFile("fixtures/broken-message-from-ax-1.bin")
	//dat, err := ioutil.ReadFile("fixtures/nsmutablestring.bin")
	//dat, err := ioutil.ReadFile("fixtures/nsnull.bin")
	dat, err := ioutil.ReadFile("fixtures/dtactivitytapmessage.bin")

	if err != nil {
		log.Fatal(err)
	}
	var remainingBytes []byte
	remainingBytes = dat
	for len(remainingBytes) > 0 {
		msg, s, err := dtx.DecodeNonBlocking(remainingBytes)
		log.Info(msg)
		remainingBytes = s
		if !assert.NoError(t, err) {
			log.Fatal("whet", err)
		}
	}

}

func TestType1Message(t *testing.T) {

	//dat, err := ioutil.ReadFile("fixtures/broken-message-from-ax-1.bin")
	//dat, err := ioutil.ReadFile("fixtures/nsmutablestring.bin")
	//dat, err := ioutil.ReadFile("fixtures/nsnull.bin")
	dat, err := ioutil.ReadFile("fixtures/unknown-d-h-h-message.bin")

	if err != nil {
		log.Fatal(err)
	}
	var remainingBytes []byte
	remainingBytes = dat
	for len(remainingBytes) > 0 {
		msg, s, err := dtx.DecodeNonBlocking(remainingBytes)
		log.Info(msg)
		remainingBytes = s
		if !assert.NoError(t, err) {
			log.Fatal("whet", err)
		}
	}

}

func TestDecoder(t *testing.T) {
	dat, err := ioutil.ReadFile("fixtures/notifyOfPublishedCapabilites")
	if err != nil {
		log.Fatal(err)
	}
	msg, remainingBytes, err := dtx.DecodeNonBlocking(dat)
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

	msg, err = dtx.ReadMessage(bytes.NewReader(dat))
	if assert.NoError(t, err) {

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
