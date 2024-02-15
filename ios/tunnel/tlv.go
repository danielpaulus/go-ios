package tunnel

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
)

// https://github.com/45clouds/WirelessCarPlay/blob/e7a2d3e8035de262b1867a90bdf5c52a039d8862/source/AccessorySDK/Support/PairingUtils.c#L175

/*
#define kTLVType_Method					0x00 // Pairing method to use.
	#define kTLVMethod_PairSetup			0 // Pair-setup.
	#define kTLVMethod_MFiPairSetup			1 // MFi pair-setup.
	#define kTLVMethod_Verify				2 // Pair-verify.
#define kTLVType_Identifier				0x01 // Identifier of the peer.
#define kTLVType_Salt					0x02 // 16+ bytes of random salt.
#define kTLVType_PublicKey				0x03 // Curve25519, SRP public key, or signed Ed25519 key.
#define kTLVType_Proof					0x04 // SRP proof.
#define kTLVType_EncryptedData			0x05 // Encrypted bytes. Use AuthTag to authenticate.
#define kTLVType_State					0x06 // State of the pairing process.
#define kTLVType_Error					0x07 // Error code. Missing means no error.
	#define kTLVError_Reserved0				0x00 // Must not be used in any TLV.
	#define kTLVError_Unknown				0x01 // Generic error to handle unexpected errors.
	#define kTLVError_Authentication		0x02 // Setup code or signature verification failed.
	#define kTLVError_Backoff				0x03 // Client must look at <RetryDelay> TLV item and wait before retrying.
	#define kTLVError_UnknownPeer			0x04 // Peer is not paired.
	#define kTLVError_MaxPeers				0x05 // Server cannot accept any more pairings.
	#define kTLVError_MaxTries				0x06 // Server reached its maximum number of authentication attempts
#define kTLVType_RetryDelay					0x08 // Seconds to delay until retrying setup.
#define kTLVType_Certificate			0x09 // X.509 Certificate.
#define kTLVType_Signature				0x0A // Ed25519 or MFi auth IC signature.
#define kTLVType_ReservedB				0x0B // Reserved.
#define kTLVType_FragmentData			0x0C // Non-last fragment of data. If length is 0, it's an ack.
#define kTLVType_FragmentLast			0x0D // Last fragment of data.
*/

type tlvType uint8
type pairingState uint8

const (
	pairStateStartRequest     = byte(0x01)
	pairStateStartResponse    = 0x02
	pairStateVerifyRequest    = 0x03
	pairStateVerifyResponse   = 0x04
	pairStateExchangeRequest  = 0x05
	pairStateExchangeResponse = 0x06
	pairStateDone             = 0x07
)

const (
	typeMethod        = tlvType(0x00)
	typeIdentifier    = tlvType(0x01)
	typeSalt          = tlvType(0x02)
	typePublicKey     = tlvType(0x03)
	typeProof         = tlvType(0x04)
	typeEncryptedData = tlvType(0x05)
	typeState         = tlvType(0x06)
	typeError         = tlvType(0x07)
	typeSignature     = tlvType(0x0A)
	typeInfo          = tlvType(0x11)
)

type tlvBuffer struct {
	buf *bytes.Buffer
}

func newTlvBuffer() tlvBuffer {
	return tlvBuffer{buf: bytes.NewBuffer(nil)}
}

func (b tlvBuffer) writeData(t tlvType, data []byte) {
	if len(data) > math.MaxUint8 {
		b.buf.WriteByte(byte(t))
		b.buf.WriteByte(byte(math.MaxUint8))
		b.buf.Write(data[:math.MaxUint8])
		b.writeData(t, data[math.MaxUint8:])
	} else {
		b.buf.WriteByte(byte(t))
		b.buf.WriteByte(byte(len(data)))
		b.buf.Write(data)
	}
}

func (b tlvBuffer) writeByte(t tlvType, v uint8) {
	b.writeData(t, []byte{v})
}

func (b tlvBuffer) bytes() []byte {
	return b.buf.Bytes()
}

type tlvReader []byte

func (r tlvReader) readCoalesced(t tlvType) ([]byte, error) {
	reader := bytes.NewReader(r)
	buf := bytes.NewBuffer(nil)

	for {
		chunkType, err := reader.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		l, _ := reader.ReadByte()
		if tlvType(chunkType) == t {
			_, err = io.CopyN(buf, reader, int64(l))
		} else {
			_, err = io.CopyN(io.Discard, reader, int64(l))
		}
		if err != nil {
			return nil, fmt.Errorf("readCoalesced: failed to read bytes of length %d: %w", l, err)
		}
	}

	return buf.Bytes(), nil
}

type tlvError byte

var errorNames = [...]string{"reserved0", "unknown", "authentication", "backoff", "unknownpeer", "maxpeers", "maxtries"}

func (e tlvError) Error() string {
	if int(e) >= 0 && int(e) < len(errorNames) {
		return errorNames[e]
	}
	return fmt.Sprintf("unknown error code '%d'", e)
}
