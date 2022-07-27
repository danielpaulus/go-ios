package afc

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	Afc_magic                      uint64 = 0x4141504c36414643
	Afc_header_size                uint64 = 40
	Afc_operation_status           uint64 = 0x00000001
	Afc_operation_data             uint64 = 0x00000002
	Afc_operation_read_dir         uint64 = 0x00000003
	Afc_operation_remove_path      uint64 = 0x00000008
	Afc_operation_make_dir         uint64 = 0x00000009
	Afc_operation_file_info        uint64 = 0x0000000A
	Afc_operation_file_open        uint64 = 0x0000000D
	Afc_operation_file_close       uint64 = 0x00000014
	Afc_operation_file_write       uint64 = 0x00000010
	Afc_operation_file_open_result uint64 = 0x0000000E
	Afc_operation_file_read        uint64 = 0x0000000F
)

const (
	Afc_Mode_RDONLY   uint64 = 0x00000001
	Afc_Mode_RW       uint64 = 0x00000002
	Afc_Mode_WRONLY   uint64 = 0x00000003
	Afc_Mode_WR       uint64 = 0x00000004
	Afc_Mode_APPEND   uint64 = 0x00000005
	Afc_Mode_RDAPPEND uint64 = 0x00000006
)

const (
	Afc_Err_Success                = 0
	Afc_Err_UnknownError           = 1
	Afc_Err_OperationHeaderInvalid = 2
	Afc_Err_NoResources            = 3
	Afc_Err_ReadError              = 4
	Afc_Err_WriteError             = 5
	Afc_Err_UnknownPacketType      = 6
	Afc_Err_InvalidArgument        = 7
	Afc_Err_ObjectNotFound         = 8
	Afc_Err_ObjectIsDir            = 9
	Afc_Err_PermDenied             = 10
	Afc_Err_ServiceNotConnected    = 11
	Afc_Err_OperationTimeout       = 12
	Afc_Err_TooMuchData            = 13
	Afc_Err_EndOfData              = 14
	Afc_Err_OperationNotSupported  = 15
	Afc_Err_ObjectExists           = 16
	Afc_Err_ObjectBusy             = 17
	Afc_Err_NoSpaceLeft            = 18
	Afc_Err_OperationWouldBlock    = 19
	Afc_Err_IoError                = 20
	Afc_Err_OperationInterrupted   = 21
	Afc_Err_OperationInProgress    = 22
	Afc_Err_InternalError          = 23
	Afc_Err_MuxError               = 30
	Afc_Err_NoMemory               = 31
	Afc_Err_NotEnoughData          = 32
	Afc_Err_DirNotEmpty            = 33
)

func getError(errorCode uint64) error {
	switch errorCode {
	case Afc_Err_UnknownError:
		return errors.New("UnknownError")
	case Afc_Err_OperationHeaderInvalid:
		return errors.New("OperationHeaderInvalid")
	case Afc_Err_NoResources:
		return errors.New("NoResources")
	case Afc_Err_ReadError:
		return errors.New("ReadError")
	case Afc_Err_WriteError:
		return errors.New("WriteError")
	case Afc_Err_UnknownPacketType:
		return errors.New("UnknownPacketType")
	case Afc_Err_InvalidArgument:
		return errors.New("InvalidArgument")
	case Afc_Err_ObjectNotFound:
		return errors.New("ObjectNotFound")
	case Afc_Err_ObjectIsDir:
		return errors.New("ObjectIsDir")
	case Afc_Err_PermDenied:
		return errors.New("PermDenied")
	case Afc_Err_ServiceNotConnected:
		return errors.New("ServiceNotConnected")
	case Afc_Err_OperationTimeout:
		return errors.New("OperationTimeout")
	case Afc_Err_TooMuchData:
		return errors.New("TooMuchData")
	case Afc_Err_EndOfData:
		return errors.New("EndOfData")
	case Afc_Err_OperationNotSupported:
		return errors.New("OperationNotSupported")
	case Afc_Err_ObjectExists:
		return errors.New("ObjectExists")
	case Afc_Err_ObjectBusy:
		return errors.New("ObjectBusy")
	case Afc_Err_NoSpaceLeft:
		return errors.New("NoSpaceLeft")
	case Afc_Err_OperationWouldBlock:
		return errors.New("OperationWouldBlock")
	case Afc_Err_IoError:
		return errors.New("IoError")
	case Afc_Err_OperationInterrupted:
		return errors.New("OperationInterrupted")
	case Afc_Err_OperationInProgress:
		return errors.New("OperationInProgress")
	case Afc_Err_InternalError:
		return errors.New("InternalError")
	case Afc_Err_MuxError:
		return errors.New("MuxError")
	case Afc_Err_NoMemory:
		return errors.New("NoMemory")
	case Afc_Err_NotEnoughData:
		return errors.New("NotEnoughData")
	case Afc_Err_DirNotEmpty:
		return errors.New("DirNotEmpty")
	default:
		return nil
	}
}

type AfcPacketHeader struct {
	Magic         uint64
	Entire_length uint64
	This_length   uint64
	Packet_num    uint64
	Operation     uint64
}

type AfcPacket struct {
	Header        AfcPacketHeader
	HeaderPayload []byte
	Payload       []byte
}

func Decode(reader io.Reader) (AfcPacket, error) {
	var header AfcPacketHeader
	err := binary.Read(reader, binary.LittleEndian, &header)
	if err != nil {
		return AfcPacket{}, err
	}
	if header.Magic != Afc_magic {
		return AfcPacket{}, fmt.Errorf("Wrong magic:%x expected: %x", header.Magic, Afc_magic)
	}
	headerPayloadLength := header.This_length - Afc_header_size
	headerPayload := make([]byte, headerPayloadLength)
	_, err = io.ReadFull(reader, headerPayload)
	if err != nil {
		return AfcPacket{}, err
	}
	contentPayloadLength := header.Entire_length - header.This_length
	payload := make([]byte, contentPayloadLength)
	_, err = io.ReadFull(reader, payload)
	if err != nil {
		return AfcPacket{}, err
	}
	return AfcPacket{header, headerPayload, payload}, nil
}

func Encode(packet AfcPacket, writer io.Writer) error {
	err := binary.Write(writer, binary.LittleEndian, packet.Header)
	if err != nil {
		return err
	}
	_, err = writer.Write(packet.HeaderPayload)
	if err != nil {
		return err
	}

	_, err = writer.Write(packet.Payload)
	if err != nil {
		return err
	}
	return nil
}
