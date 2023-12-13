package tunnel

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/chacha20poly1305"
	"testing"
)

func TestWriteEvent(t *testing.T) {
	conn := new(xpcConnMock)

	pd := pairingData{
		data:            make([]byte, 16),
		kind:            "kind",
		sendingHost:     "host",
		startNewSession: true,
	}

	rw := newControlChannelReadWriter(conn)

	expected := map[string]interface{}{
		"mangledTypeName": "RemotePairing.ControlChannelMessageEnvelope",
		"value": map[string]interface{}{
			"message": map[string]interface{}{
				"plain": map[string]interface{}{
					"_0": map[string]interface{}{
						"event": map[string]interface{}{
							"_0": map[string]interface{}{
								"pairingData": map[string]interface{}{
									"_0": map[string]interface{}{
										"data":            make([]byte, 16),
										"kind":            "kind",
										"sendingHost":     "host",
										"startNewSession": true,
									}},
							},
						},
					},
				},
			},
			"originatedBy":   "host",
			"sequenceNumber": uint64(1),
		},
	}

	conn.On("Send", expected, []uint32(nil)).Return(nil)

	err := rw.writeEvent(&pd)
	assert.NoError(t, err)

	conn.On("ReceiveOnClientServerStream").Return(expected, nil)

	var pdResponse pairingData
	err = rw.readEvent(&pdResponse)
	assert.NoError(t, err)
}

func TestReadWriteRequest(t *testing.T) {
	conn := new(xpcConnMock)
	rw := newControlChannelReadWriter(conn)

	expected := map[string]interface{}{
		"mangledTypeName": "RemotePairing.ControlChannelMessageEnvelope",
		"value": map[string]interface{}{
			"message": map[string]interface{}{
				"plain": map[string]interface{}{
					"_0": map[string]interface{}{
						"request": map[string]interface{}{
							"_0": map[string]interface{}{
								"handshake": map[string]interface{}{
									"_0": map[string]interface{}{
										"hostOptions": map[string]interface{}{
											"attemptPairVerify": true,
										},
										"wireProtocolVersion": int64(19),
									}},
							},
						},
					},
				},
			},
			"originatedBy":   "host",
			"sequenceNumber": uint64(1),
		},
	}

	conn.On("Send", expected, []uint32(nil)).Return(nil)

	err := rw.writeRequest(map[string]interface{}{
		"handshake": map[string]interface{}{
			"_0": map[string]interface{}{
				"hostOptions": map[string]interface{}{
					"attemptPairVerify": true,
				},
				"wireProtocolVersion": int64(19),
			},
		},
	})

	assert.NoError(t, err)
}

func TestWriteIncrementsSequenceNumber(t *testing.T) {
	conn := new(xpcConnMock)
	rw := newControlChannelReadWriter(conn)

	conn.On("Send", mock.Anything, []uint32(nil)).Return(nil)

	event := pairingData{}

	err := rw.writeEvent(&event)
	require.NoError(t, err)
	assert.Equal(t, uint64(2), rw.seqNr)
	err = rw.writeEvent(&event)
	require.NoError(t, err)
	assert.Equal(t, uint64(3), rw.seqNr)
}

func TestReadWriteCipher(t *testing.T) {
	conn := new(xpcConnMock)
	controlChannel := newControlChannelReadWriter(conn)

	key := make([]byte, 32)
	cipher, err := chacha20poly1305.New(key)
	require.NoError(t, err)

	cs := newCipherStream(controlChannel, cipher, cipher)

	req := map[string]interface{}{
		"test": "value",
	}

	j, err := json.Marshal(req)
	require.NoError(t, err)

	nonce := make([]byte, cipher.NonceSize())

	expected := map[string]interface{}{
		"mangledTypeName": "RemotePairing.ControlChannelMessageEnvelope",
		"value": map[string]interface{}{
			"message": map[string]interface{}{
				"streamEncrypted": map[string]interface{}{
					"_0": cipher.Seal(nil, nonce, j, nil),
				},
			},
			"originatedBy":   "host",
			"sequenceNumber": uint64(1),
		},
	}

	conn.On("Send", expected, []uint32(nil)).Return(nil)
	err = cs.write(req)
	require.NoError(t, err)

	conn.On("ReceiveOnClientServerStream").Return(expected, nil)
	var res map[string]interface{}
	err = cs.read(&res)
	require.NoError(t, err)

	assert.Equal(t, req, res)
}

func TestCipherJsonEncode(t *testing.T) {
	conn := new(xpcConnMock)
	controlChannel := newControlChannelReadWriter(conn)

	key := make([]byte, 32)
	cipher, err := chacha20poly1305.New(key)
	require.NoError(t, err)

	cs := newCipherStream(controlChannel, cipher, cipher)

	conn.On("Send", mock.MatchedBy(func(m map[string]interface{}) bool {
		if encrypted, err := getChildMap(m, "value", "message", "streamEncrypted"); err == nil {
			if enc, ok := encrypted["_0"].([]byte); ok {
				plain, err := cipher.Open(nil, cs.nonce, enc, nil)
				require.NoError(t, err)
				var decoded map[string]interface{}
				err = json.Unmarshal(plain, &decoded)
				require.NoError(t, err)
				assert.Equal(t, map[string]interface{}{"test": "value"}, decoded)
			}
		}
		return true
	}), []uint32(nil)).Return(nil)
	err = cs.write(map[string]interface{}{
		"test": "value",
	})
	require.NoError(t, err)
}

type xpcConnMock struct {
	mock.Mock
}

func (x *xpcConnMock) Send(data map[string]interface{}, flags ...uint32) error {
	args := x.Called(data, flags)
	return args.Error(0)
}

func (x *xpcConnMock) ReceiveOnClientServerStream() (map[string]interface{}, error) {
	args := x.Called()
	return args.Get(0).(map[string]interface{}), args.Error(1)
}
