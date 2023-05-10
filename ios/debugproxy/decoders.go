package debugproxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	log "github.com/sirupsen/logrus"
)

type decoder interface {
	decode([]byte)
}

type dtxDecoder struct {
	jsonFilePath string
	binFilePath  string
	buffer       bytes.Buffer
	isBroken     bool
	log          *log.Entry
}

type MessageWithMetaInfo struct {
	DtxMessage   interface{}
	MessageType  string
	TimeReceived time.Time
	OffsetInDump int64
	Length       int
}

func NewDtxDecoder(jsonFilePath string, binFilePath string, log *log.Entry) decoder {
	return &dtxDecoder{jsonFilePath: jsonFilePath, binFilePath: binFilePath, buffer: bytes.Buffer{}, isBroken: false, log: log}
}

func (f *dtxDecoder) decode(data []byte) {
	file, err := os.OpenFile(f.binFilePath+".raw",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Println(err)
	}

	file.Write(data)
	file.Close()

	if f.isBroken {
		// when an error happens while decoding, this flag prevents from flooding the logs with errors
		// while still dumping binary to debug later
		return
	}
	f.buffer.Write(data)
	slice := f.buffer.Next(f.buffer.Len())
	written := 0
	for {
		msg, remainingbytes, err := dtx.DecodeNonBlocking(slice)
		if dtx.IsIncomplete(err) {
			f.buffer.Reset()
			f.buffer.Write(slice)
			break
		}
		if err != nil {
			f.log.Errorf("Failed decoding DTX:%s, continuing bindumping", err)
			f.log.Info(fmt.Sprintf("%x", slice))
			f.isBroken = true
		}
		slice = remainingbytes

		file, err := os.OpenFile(f.binFilePath,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			log.Println(err)
		}
		s, _ := file.Stat()
		offset := s.Size()
		file.Write(msg.RawBytes)
		file.Close()

		file, err = os.OpenFile(f.jsonFilePath,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			log.Println(err)
		}

		type Alias dtx.Message
		auxi := ""
		if msg.HasAuxiliary() {
			auxi = msg.Auxiliary.String()
		}
		aux := &struct {
			AuxiliaryContents string
			*Alias
		}{
			AuxiliaryContents: auxi,
			Alias:             (*Alias)(&msg),
		}
		aux.RawBytes = nil
		jsonMetaInfo := MessageWithMetaInfo{aux, "dtx", time.Now(), offset, len(msg.RawBytes)}

		mylog := f.log
		if strings.Contains(f.binFilePath, "from-device") {
			mylog = f.log.WithFields(log.Fields{"d": "in"})
		}
		if strings.Contains(f.binFilePath, "to-device") {
			mylog = f.log.WithFields(log.Fields{"d": "out"})
		}
		logDtxMessageNice(mylog, msg)
		jsonmsg, err := json.Marshal(jsonMetaInfo)
		file.Write(jsonmsg)
		io.WriteString(file, "\n")
		file.Close()

		written += len(msg.RawBytes)
	}
}

func logDtxMessageNice(log *log.Entry, msg dtx.Message) {
	if msg.PayloadHeader.MessageType == dtx.Methodinvocation {
		expectsReply := ""
		if msg.ExpectsReply {
			expectsReply = "e"
		}
		log.Infof("%d.%d%s c%d %s %s", msg.Identifier, msg.ConversationIndex, expectsReply, msg.ChannelCode, msg.Payload[0], msg.Auxiliary)
		return
	}
	if msg.PayloadHeader.MessageType == dtx.Ack {
		log.Infof("%d.%d c%d Ack", msg.Identifier, msg.ConversationIndex, msg.ChannelCode)
		return
	}
	if msg.PayloadHeader.MessageType == dtx.UnknownTypeOne {
		if len(msg.Payload) > 0 {
			log.Infof("type1 with payload: %x", msg.Payload[0])
			return
		}
		log.Infof("type1 without payload: %+v", msg)
		return
	}
	if msg.PayloadHeader.MessageType == dtx.ResponseWithReturnValueInPayload {
		log.Infof("%d.%d c%d response: %s", msg.Identifier, msg.ConversationIndex, msg.ChannelCode, msg.Payload[0])
		return
	}
	if msg.PayloadHeader.MessageType == dtx.DtxTypeError {
		log.Infof("%d.%d c%d error: %s", msg.Identifier, msg.ConversationIndex, msg.ChannelCode, msg.Payload[0])
		return
	}
	log.Infof("%+v", msg)
}

type binaryOnlyDumper struct {
	path string
}

// NewNoOpDecoder does nothing
func NewBinDumpOnly(jsonFilePath string, dumpFilePath string, log *log.Entry) decoder {
	return binaryOnlyDumper{dumpFilePath}
}

func (n binaryOnlyDumper) decode(bytes []byte) {
	writeBytes(n.path, bytes)
}

func writeBytes(filePath string, data []byte) {
	file, err := os.OpenFile(filePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		panic(fmt.Sprintf("Could not write to file error: %v path:'%s'", err, filePath))
	}

	file.Write(data)
	file.Close()
}
