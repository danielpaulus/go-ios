package dtx_test

import (
	"bufio"
	"bytes"
	"os"
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

	dat, err = os.ReadFile("fixtures/notifyOfPublishedCapabilites")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < len(dat)-4; i++ {
		_, _, err = dtx.DecodeNonBlocking(dat[0 : 4+i])
		assert.True(t, dtx.IsIncomplete(err))
	}
}

func TestLZ4CompressedDtxMessage(t *testing.T) {
	dat, err := os.ReadFile("fixtures/instruments-metrics-dtx.bin")
	if err != nil {
		t.Fatal(err)
	}

	fixtureMsg, _, err := dtx.DecodeNonBlocking(dat)
	if err != nil {
		t.Fatal(err)
	}
	log.Infof("%v", fixtureMsg)
	assert.NoError(t, err)
}

func TestCodec2(t *testing.T) {
	dat, err := os.ReadFile("fixtures/requestChannelWithCodeIdentifier.bin")
	if err != nil {
		t.Fatal(err)
	}

	fixtureMsg, _, err := dtx.DecodeNonBlocking(dat)
	if err != nil {
		t.Fatal(err)
	}
	payloadBytes, err := nskeyedarchiver.ArchiveBin(fixtureMsg.Payload[0])
	assert.NoError(t, err)
	encodedBytes, err := dtx.Encode(fixtureMsg.Identifier, fixtureMsg.ConversationIndex, fixtureMsg.ChannelCode, fixtureMsg.ExpectsReply, fixtureMsg.PayloadHeader.MessageType, payloadBytes, fixtureMsg.Auxiliary)
	if assert.NoError(t, err) {
		msg, remainingBytes, err := dtx.DecodeNonBlocking(encodedBytes)
		assert.Equal(t, 0, len(remainingBytes))
		assert.NoError(t, err)
		assert.Equal(t, fixtureMsg.Payload, msg.Payload)
	}
}

func TestCodec(t *testing.T) {
	dat, err := os.ReadFile("fixtures/requestChannelWithCodeIdentifier.bin")
	if err != nil {
		t.Fatal(err)
	}
	var remainingBytes []byte
	remainingBytes = dat
	for len(remainingBytes) > 0 {
		msg, s, err := dtx.DecodeNonBlocking(remainingBytes)
		log.Info(msg.StringDebug())
		remainingBytes = s
		if !assert.NoError(t, err) {
			t.Fatal("whet", err)
		}
		bytes, err := dtx.Encode(3, 0, 0, true, 2, msg.RawBytes[303:], msg.Auxiliary)
		if assert.NoError(t, err) {
			assert.Equal(t, dat, bytes)
		}
	}
}

func TestAXDump(t *testing.T) {
	// dat, err := os.ReadFile("fixtures/broken-message-from-ax-1.bin")
	// dat, err := os.ReadFile("fixtures/nsmutablestring.bin")
	// dat, err := os.ReadFile("fixtures/nsnull.bin")
	dat, err := os.ReadFile("fixtures/dtactivitytapmessage.bin")
	if err != nil {
		t.Fatal(err)
	}
	var remainingBytes []byte
	remainingBytes = dat
	for len(remainingBytes) > 0 {
		msg, s, err := dtx.DecodeNonBlocking(remainingBytes)
		log.Info(msg)
		remainingBytes = s
		if !assert.NoError(t, err) {
			t.Fatal("whet", err)
		}
	}
}

func TestType1Message(t *testing.T) {
	// dat, err := os.ReadFile("fixtures/broken-message-from-ax-1.bin")
	// dat, err := os.ReadFile("fixtures/nsmutablestring.bin")
	// dat, err := os.ReadFile("fixtures/nsnull.bin")
	dat, err := os.ReadFile("fixtures/unknown-d-h-h-message.bin")
	if err != nil {
		t.Fatal(err)
	}
	var remainingBytes []byte
	remainingBytes = dat
	for len(remainingBytes) > 0 {
		msg, s, err := dtx.DecodeNonBlocking(remainingBytes)
		log.Info(msg)
		remainingBytes = s
		if !assert.NoError(t, err) {
			t.Fatal("whet", err)
		}
	}
}

func TestFragmentedMessage(t *testing.T) {
	dat, err := os.ReadFile("fixtures/fragmentedmessage.bin")
	if err != nil {
		t.Fatal(err)
	}

	// test the non blocking decoder first
	msg, remainingBytes, err := dtx.DecodeNonBlocking(dat)
	if assert.NoError(t, err) {
		assert.Equal(t, 79707, len(remainingBytes))
		assert.Equal(t, uint16(3), msg.Fragments)
		assert.Equal(t, uint16(0), msg.FragmentIndex)
		assert.Equal(t, false, msg.HasPayload())
	}
	defragmenter := dtx.NewFragmentDecoder(msg)
	msg, remainingBytes, err = dtx.DecodeNonBlocking(remainingBytes)
	if assert.NoError(t, err) {
		assert.Equal(t, 14171, len(remainingBytes))
		assert.Equal(t, uint16(3), msg.Fragments)
		assert.Equal(t, uint16(1), msg.FragmentIndex)
		assert.Equal(t, false, msg.HasPayload())
	}
	defragmenter.AddFragment(msg)
	msg, remainingBytes, err = dtx.DecodeNonBlocking(remainingBytes)
	if assert.NoError(t, err) {
		assert.Equal(t, 0, len(remainingBytes))
		assert.Equal(t, uint16(3), msg.Fragments)
		assert.Equal(t, uint16(2), msg.FragmentIndex)
		assert.Equal(t, false, msg.HasPayload())
	}
	defragmenter.AddFragment(msg)
	assert.True(t, defragmenter.HasFinished())
	nonblockingFullMessage := defragmenter.Extract()

	// now test that the blocking decoder creates the same message and that it is decodeable
	dtxReader := bufio.NewReader(bytes.NewReader(dat))
	msg, err = dtx.ReadMessage(dtxReader)
	if assert.NoError(t, err) {
		assert.Equal(t, uint16(3), msg.Fragments)
		assert.Equal(t, uint16(0), msg.FragmentIndex)
		assert.Equal(t, false, msg.HasPayload())
		assert.Equal(t, true, msg.IsFirstFragment())
	}
	defragmenter = dtx.NewFragmentDecoder(msg)
	msg, err = dtx.ReadMessage(dtxReader)
	if assert.NoError(t, err) {
		assert.Equal(t, uint16(3), msg.Fragments)
		assert.Equal(t, uint16(1), msg.FragmentIndex)
		assert.Equal(t, true, msg.IsFragment())
	}
	defragmenter.AddFragment(msg)
	msg, err = dtx.ReadMessage(dtxReader)
	if assert.NoError(t, err) {
		assert.Equal(t, uint16(3), msg.Fragments)
		assert.Equal(t, uint16(2), msg.FragmentIndex)
		assert.Equal(t, true, msg.IsLastFragment())
	}
	defragmenter.AddFragment(msg)
	assert.Equal(t, true, defragmenter.HasFinished())
	defraggedMessage := defragmenter.Extract()
	assert.Equal(t, defraggedMessage, nonblockingFullMessage)
	dtxReader = bufio.NewReader(bytes.NewReader(defraggedMessage))
	_, err = dtx.ReadMessage(dtxReader)
	assert.NoError(t, err)
}

func TestDecoder(t *testing.T) {
	dat, err := os.ReadFile("fixtures/notifyOfPublishedCapabilites")
	if err != nil {
		t.Fatal(err)
	}

	reader := bufio.NewReader(bytes.NewReader(append(dat[:], dat[:]...)))
	for i := 0; i < 2; i++ {
		msg, err := dtx.ReadMessage(reader)
		if assert.NoError(t, err) {
			assert.Equal(t, msg.Fragments, uint16(1))
			assert.Equal(t, msg.FragmentIndex, uint16(0))
			assert.Equal(t, msg.MessageLength, 612)
			assert.Equal(t, 0, msg.ChannelCode)
			assert.Equal(t, false, msg.ExpectsReply)
			assert.Equal(t, 2, msg.Identifier)
			assert.Equal(t, 0, msg.ChannelCode)

			assert.Equal(t, dtx.MessageType(2), msg.PayloadHeader.MessageType)
			assert.Equal(t, uint32(425), msg.PayloadHeader.AuxiliaryLength)
			assert.Equal(t, uint32(596), msg.PayloadHeader.TotalPayloadLength)
			assert.Equal(t, uint32(0), msg.PayloadHeader.Flags)
		}
	}
}
