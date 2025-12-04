package afc

import (
	"errors"
	"fmt"
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
		return fmt.Errorf("unknown AFC error code: %d", errorCode)
	}
}

type afcError struct {
	code int
}

func (a afcError) Error() string {
	return fmt.Sprintf("afc error code: %d", a.code)
}

func isPermissionDeniedError(err error) bool {
	var aError afcError
	return errors.As(err, &aError) && aError.code == errPermDenied
}
