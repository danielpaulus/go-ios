package wrtc

import (
	"io"

	"github.com/google/uuid"
)

func (r RTCConnection) TransferFile(file io.Reader, size int) (uuid.UUID, error) {
	cmd := map[string]interface{}{}
	cmd["cmd"] = "filetransfer"
	cmd["size"] = size

	conn, err := r.StreamingResponse("filetransfer")

}
