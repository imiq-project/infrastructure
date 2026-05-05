package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEm500DecodeSuccess(t *testing.T) {
	payload := []byte{
		0x01, 0x75, 0x64, // battery
		0x03, 0x67, 0x10, 0x01, // temperature
		0x04, 0x68, 0x71, // humidity
		0x05, 0x7d, 0x67, 0x04, // co2
		0x06, 0x73, 0x68, 0x27, // pressure
	}
	result, err := decodeEm500Payload(payload)
	require.Nil(t, err)
	assert.Equal(t, 100, result["battery"])
	assert.Equal(t, 27.2, result["temperature"])
	assert.Equal(t, 56.5, result["humidity"])
	assert.Equal(t, 1127, result["co2"])
	assert.Equal(t, 1008.8, result["pressure"])
}

func TestDecodeRealPayload(t *testing.T) {
	payload1 := []byte{
		0xFF, 0x0B, 0xFF,
		0xFF, 0x01, 0x01,
		0xFF, 0x16, 0x61, 0x26, 0xF3, 0x30, 0x55, 0x82, 0x60, 0x09, 0xFF, 0x09, 0x02, 0x10, 0xFF, 0x0A, 0x01, 0x11,
	}
	result, err := decodeEm500Payload(payload1)
	require.Nil(t, err)

	payload2 := []byte{
		0x01, 0x75, 0x64, 0x05, 0x7D, 0x94, 0x05, 0x06, 0x73, 0x49, 0x27, 0x03, 0x67, 0xF3, 0x00, 0x04, 0x68, 0x76,
	}
	result, err = decodeEm500Payload(payload2)
	require.Nil(t, err)

	payload3 := []byte{
		0x05, 0x7D, 0x10, 0x03, 0x06, 0x73, 0x43, 0x27, 0x03, 0x67, 0xFB, 0x00, 0x04, 0x68, 0x69,
	}

	result, err = decodeEm500Payload(payload3)
	require.Nil(t, err)
	assert.Equal(t, 25.1, result["temperature"])
	assert.Equal(t, 52.5, result["humidity"])
	assert.Equal(t, 784, result["co2"])
	assert.Equal(t, 1005.1, result["pressure"])
}

func TestEm500DecodeError(t *testing.T) {
	for len := 1; len < 255; len++ {
		payload := make([]byte, len)
		// will payload with 0x01 (valid type id)
		for i := range payload {
			payload[i] = 0x01
		}
		_, err := decodeEm500Payload(payload)
		assert.NotNil(t, err)
	}
	emptyPayload := []byte{}
	result, err := decodeEm500Payload(emptyPayload)
	require.Nil(t, err)
	assert.Empty(t, result)
}
