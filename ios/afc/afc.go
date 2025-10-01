package afc

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	Afc_magic                              uint64 = 0x4141504c36414643
	Afc_header_size                        uint64 = 40
	Afc_operation_status                   uint64 = 0x00000001
	Afc_operation_data                     uint64 = 0x00000002
	Afc_operation_read_dir                 uint64 = 0x00000003
	Afc_operation_remove_path              uint64 = 0x00000008
	Afc_operation_make_dir                 uint64 = 0x00000009
	Afc_operation_file_info                uint64 = 0x0000000A
	Afc_operation_device_info              uint64 = 0x0000000B
	Afc_operation_file_open                uint64 = 0x0000000D
	Afc_operation_file_close               uint64 = 0x00000014
	Afc_operation_file_write               uint64 = 0x00000010
	Afc_operation_file_open_result         uint64 = 0x0000000E
	Afc_operation_file_read                uint64 = 0x0000000F
	Afc_operation_remove_path_and_contents uint64 = 0x00000022
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
	errSuccess                = 0
	errUnknown                = 1
	errOperationHeaderInvalid = 2
	errNoResources            = 3
	errReadError              = 4
	errWriteError             = 5
	errUnknownPacketType      = 6
	errInvalidArgument        = 7
	errObjectNotFound         = 8
	errObjectIsDir            = 9
	errPermDenied             = 10
	errServiceNotConnected    = 11
	errOperationTimeout       = 12
	errTooMuchData            = 13
	errEndOfData              = 14
	errOperationNotSupported  = 15
	errObjectExists           = 16
	errObjectBusy             = 17
	errNoSpaceLeft            = 18
	errOperationWouldBlock    = 19
	errIoError                = 20
	errOperationInterrupted   = 21
	errOperationInProgress    = 22
	errInternalError          = 23
	errMuxError               = 30
	errNoMemory               = 31
	errNotEnoughData          = 32
	errDirNotEmpty            = 33
)

type DeviceInfo struct {
	Model      string
	TotalBytes uint64
	FreeBytes  uint64
	BlockSize  uint64
}

func getError(errorCode uint64) error {
	switch errorCode {
	case errUnknown:
		return errors.New("UnknownError")
	case errOperationHeaderInvalid:
		return errors.New("OperationHeaderInvalid")
	case errNoResources:
		return errors.New("NoResources")
	case errReadError:
		return errors.New("ReadError")
	case errWriteError:
		return errors.New("WriteError")
	case errUnknownPacketType:
		return errors.New("UnknownPacketType")
	case errInvalidArgument:
		return errors.New("InvalidArgument")
	case errObjectNotFound:
		return errors.New("ObjectNotFound")
	case errObjectIsDir:
		return errors.New("ObjectIsDir")
	case errPermDenied:
		return errors.New("PermDenied")
	case errServiceNotConnected:
		return errors.New("ServiceNotConnected")
	case errOperationTimeout:
		return errors.New("OperationTimeout")
	case errTooMuchData:
		return errors.New("TooMuchData")
	case errEndOfData:
		return errors.New("EndOfData")
	case errOperationNotSupported:
		return errors.New("OperationNotSupported")
	case errObjectExists:
		return errors.New("ObjectExists")
	case errObjectBusy:
		return errors.New("ObjectBusy")
	case errNoSpaceLeft:
		return errors.New("NoSpaceLeft")
	case errOperationWouldBlock:
		return errors.New("OperationWouldBlock")
	case errIoError:
		return errors.New("IoError")
	case errOperationInterrupted:
		return errors.New("OperationInterrupted")
	case errOperationInProgress:
		return errors.New("OperationInProgress")
	case errInternalError:
		return errors.New("InternalError")
	case errMuxError:
		return errors.New("MuxError")
	case errNoMemory:
		return errors.New("NoMemory")
	case errNotEnoughData:
		return errors.New("NotEnoughData")
	case errDirNotEmpty:
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
