package fileservice

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/danielpaulus/go-ios/ios"
)

// sendRawData sends raw bytes through the raw connection
func sendRawData(conn ios.DeviceConnectionInterface, data []byte) error {
	// Write raw bytes directly to the connection
	_, err := conn.Write(data)
	return err
}

// receiveFileData receives file data from the data service
// Protocol: 36 bytes header + 4 bytes file size (BE) + file data
func receiveFileData(conn ios.DeviceConnectionInterface) ([]byte, error) {
	// Read the 36-byte header + 4-byte size field
	header := make([]byte, 40)
	_, err := io.ReadFull(conn, header)
	if err != nil {
		return nil, fmt.Errorf("receiveFileData: failed to read header: %w", err)
	}

	// Read file size (4 bytes, big-endian) at offset 36
	fileSize := binary.BigEndian.Uint32(header[36:40])

	// Validate file size to prevent OOM
	if fileSize > MaxFileSize {
		return nil, fmt.Errorf("receiveFileData: file size %d exceeds maximum allowed size %d", fileSize, MaxFileSize)
	}

	// Read file data
	fileData := make([]byte, fileSize)
	_, err = io.ReadFull(conn, fileData)
	if err != nil {
		return nil, fmt.Errorf("receiveFileData: failed to read file data: %w", err)
	}

	return fileData, nil
}

// receiveFileDataToWriter receives file data from the data service and streams it to a writer
// Protocol: 36 bytes header + 4 bytes file size (BE) + file data
func receiveFileDataToWriter(conn ios.DeviceConnectionInterface, writer io.Writer) error {
	// Read the 36-byte header + 4-byte size field
	header := make([]byte, 40)
	_, err := io.ReadFull(conn, header)
	if err != nil {
		return fmt.Errorf("receiveFileDataToWriter: failed to read header: %w", err)
	}

	// Read file size (4 bytes, big-endian) at offset 36
	fileSize := binary.BigEndian.Uint32(header[36:40])

	// Validate file size to prevent OOM
	if fileSize > MaxFileSize {
		return fmt.Errorf("receiveFileDataToWriter: file size %d exceeds maximum allowed size %d", fileSize, MaxFileSize)
	}

	// Stream file data in chunks
	const chunkSize = 256 * 1024 // 256KB chunks
	buffer := make([]byte, chunkSize)
	remaining := int64(fileSize)

	for remaining > 0 {
		toRead := chunkSize
		if remaining < int64(chunkSize) {
			toRead = int(remaining)
		}

		n, err := io.ReadFull(conn, buffer[:toRead])
		if err != nil {
			return fmt.Errorf("receiveFileDataToWriter: failed to read chunk: %w", err)
		}

		_, err = writer.Write(buffer[:n])
		if err != nil {
			return fmt.Errorf("receiveFileDataToWriter: failed to write chunk: %w", err)
		}

		remaining -= int64(n)
	}

	return nil
}
