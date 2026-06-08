package simlocation

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocationBytes(t *testing.T) {
	l := &locationData{lat: 37.7749, lon: -122.4194}

	got, err := l.LocationBytes()
	require.NoError(t, err)

	// Expected encoding:
	//   uint32(0) marker
	//   uint32(len(lat))  + lat ascii ("%f")
	//   uint32(len(lon))  + lon ascii ("%f")
	latStr := fmt.Sprintf("%f", 37.7749)   // "37.774900"
	lonStr := fmt.Sprintf("%f", -122.4194) // "-122.419400"

	want := new(bytes.Buffer)
	require.NoError(t, binary.Write(want, binary.BigEndian, uint32(0)))
	require.NoError(t, binary.Write(want, binary.BigEndian, uint32(len(latStr))))
	want.WriteString(latStr)
	require.NoError(t, binary.Write(want, binary.BigEndian, uint32(len(lonStr))))
	want.WriteString(lonStr)

	assert.Equal(t, want.Bytes(), got)

	// Spot-check the known-good byte layout explicitly.
	assert.Equal(t, []byte{0, 0, 0, 0}, got[0:4])
	assert.Equal(t, uint32(len("37.774900")), binary.BigEndian.Uint32(got[4:8]))
	assert.Equal(t, "37.774900", string(got[8:8+len("37.774900")]))
}
