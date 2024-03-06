package dtx

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	// DtxMessageMagic 4byte signature of every Message
	DtxMessageMagic uint32 = 0x795B3D1F
	// DtxMessageHeaderLength alwys 32 byte
	DtxMessageHeaderLength uint32 = 32
	// DtxMessagePayloadHeaderLength always 16 bytes
	DtxMessagePayloadHeaderLength uint32 = 16
	// DtxReservedBits are always 0x0
	DtxReservedBits uint32 = 0x0
)

/*
Message contains a decoded DtxMessage aka the most overcomplicated RPC protocol this planet has ever seen :-D

DTXMessages consist of a 32byte header that always starts with the DtxMessageMagic
It is followed by the 16 bytes PayloadHeader.
If there is an Auxiliary:

	Next is the 16 byte AuxiliaryHeader followed by the DtxPrimitiveDictionary containing auxiliary data
	Directly after the Auxiliary, the payload bytes are following.

	If there is no Auxiliary:
	The payload bytes follow directly after the PayloadHeader

	There is also support for fragmenting DTX messages into multiple messages, see fragmentdecoder.go for details
	how that works
*/
type Message struct {
	Fragments         uint16
	FragmentIndex     uint16
	MessageLength     int
	Identifier        int
	ConversationIndex int
	ChannelCode       int
	ExpectsReply      bool
	PayloadHeader     PayloadHeader
	Payload           []interface{}
	AuxiliaryHeader   AuxiliaryHeader
	Auxiliary         PrimitiveDictionary
	RawBytes          []byte
	fragmentBytes     []byte
}

// PayloadHeader contains the message type and Payload length
type PayloadHeader struct {
	MessageType        MessageType
	AuxiliaryLength    uint32
	TotalPayloadLength uint32
	Flags              uint32
}

func (p PayloadHeader) String() string {
	return fmt.Sprintf("[PayloadHeader MessageType: %s AuxilaryLength: %d, TotalPayloadLength: %d, Flags 0x%X]", p.MessageType, p.AuxiliaryLength, p.TotalPayloadLength, p.Flags)
}

// The AuxiliaryHeader can actually be completely ignored. We do not need to care about the buffer size
// And we already know the AuxiliarySize. The other two ints seem to be always 0 anyway. Could
// also be that Buffer and Aux Size are Uint64
type AuxiliaryHeader struct {
	BufferSize    uint32
	Unknown       uint32
	AuxiliarySize uint32
	Unknown2      uint32
}

type MessageType uint32

func (m MessageType) String() string {
	if s, ok := messageTypeLookup[m]; ok {
		return s
	}
	return fmt.Sprintf("unknown type %d", m)
}

// All the known MessageTypes
const (
	// Ack is the messagetype for a 16 byte long acknowleding DtxMessage.
	Ack MessageType = 0x0
	// Unknown
	UnknownTypeOne MessageType = 0x1
	// Methodinvocation is the messagetype for a remote procedure call style DtxMessage.
	Methodinvocation MessageType = 0x2
	// ResponseWithReturnValueInPayload is the response for a method call that has a return value
	ResponseWithReturnValueInPayload MessageType = 0x3
	// DtxTypeError is the messagetype for a DtxMessage containing an error
	DtxTypeError         MessageType = 0x4
	LZ4CompressedMessage MessageType = 0x0707
)

// This is only used for creating nice String() output
var messageTypeLookup = map[MessageType]string{
	ResponseWithReturnValueInPayload: `ResponseWithReturnValueInPayload`,
	Methodinvocation:                 `Methodinvocation`,
	Ack:                              `Ack`,
	LZ4CompressedMessage:             `LZ4Compressed`,
	UnknownTypeOne:                   `UnknownType1`,
	DtxTypeError:                     `Error`,
}

func (d Message) String() string {
	e := ""
	if d.ExpectsReply {
		e = "e"
	}
	msgtype := fmt.Sprintf("Unknown:%d", d.PayloadHeader.MessageType)
	if knowntype, ok := messageTypeLookup[d.PayloadHeader.MessageType]; ok {
		msgtype = knowntype
	}

	return fmt.Sprintf("i%d.%d%s c%d t:%s mlen:%d aux_len%d paylen%d", d.Identifier, d.ConversationIndex, e, d.ChannelCode, msgtype,
		d.MessageLength, d.PayloadHeader.AuxiliaryLength, d.PayloadLength())
}

// StringDebug prints the Message and its Payload/Auxiliary
func (d Message) StringDebug() string {
	blocks := make([]string, 0, 4)

	blocks = append(blocks, fmt.Sprintf("[MessageHeader ChannelCode: %d Identifier: %d ConversationIndex %d]", d.ChannelCode, d.Identifier, d.ConversationIndex))
	blocks = append(blocks, d.PayloadHeader.String())
	if d.HasPayload() {
		b, _ := json.Marshal(d.Payload[0])
		blocks = append(blocks, fmt.Sprintf("[Payload: %s]", string(b)))
	}
	if d.HasAuxiliary() {
		blocks = append(blocks, d.AuxiliaryHeader.String())
		blocks = append(blocks, d.Auxiliary.String())

	}
	return strings.Join(blocks, ", ")
}

// HasError returns true when the MessageType in this message's PayloadHeader equals 0x4 and false otherwise.
func (d Message) HasError() bool {
	return d.PayloadHeader.MessageType == DtxTypeError
}

func (a AuxiliaryHeader) String() string {
	return fmt.Sprintf("BufSiz:%d Unknown:%d AuxSiz:%d Unknown2:%d", a.BufferSize, a.Unknown, a.AuxiliarySize, a.Unknown2)
}
