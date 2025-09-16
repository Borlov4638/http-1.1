package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaders(t *testing.T) {
	// TEST: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// TEST: Valid single header with extra spase
	headers = NewHeaders()
	data = []byte("   Host:    localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 29, n)
	assert.False(t, done)

	// TEST: Valid 2 headers with existing headers
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\n     Port:    1488\r\n\r\n")
	n, done, err = headers.Parse(data)
	assert.False(t, done)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	n, done, err = headers.Parse(data[n:])
	require.NoError(t, err)
	assert.Equal(t, "1488", headers["port"])
	assert.Equal(t, 20, n)
	assert.False(t, done)

	// TEST: Valid single header with extra spase
	headers = NewHeaders()
	data = []byte("\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, 0, n)
	assert.True(t, done)

	// TEST: Header concatination
	headers = Headers{
		"host": "192.168.0.1",
	}
	data = []byte("Host: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	assert.Equal(t, "192.168.0.1, localhost:42069", headers["host"])
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// TEST: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// TEST: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       H©st: localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}
