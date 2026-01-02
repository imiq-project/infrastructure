package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSenseCapDecodeSuccess(t *testing.T) {
	payload1 := []byte{0x01, 0x01, 0x15, 0x41, 0x00, 0x00, 0x00, 0x5E, 0x05, 0x00, 0x11}
	for _, frameId := range []byte{0x01, 0x4A} {
		payload1[0] = frameId
		result, err := decodeSenseCapPayload(payload1)
		require.Nil(t, err)
		assert.Equal(t, 27.7, result["temperature"])
		assert.Equal(t, 65, result["humidity"])
		assert.Equal(t, 94, result["lightIntensity"])
		assert.Equal(t, 0.5, result["lightUV"])
		assert.Equal(t, 1.7, result["windSpeed"])
	}

	payload2 := []byte{0x02, 0x01, 0x56, 0x00, 0x00, 0x00, 0xFE, 0x27, 0x03}
	for _, frameId := range []byte{0x02, 0x4B} {
		payload2[0] = frameId
		result, err := decodeSenseCapPayload(payload2)
		require.Nil(t, err)
		assert.Equal(t, 342, result["windDirection"])
		assert.Equal(t, 0.254, result["rainIntensity"])
		assert.Equal(t, 99870, result["barometricPressure"])
	}

	payload3 := []byte{0x4C, 0x00, 0x0B, 0x00, 0x00, 0x06, 0xF2}
	result, err := decodeSenseCapPayload(payload3)
	require.Nil(t, err)
	assert.Equal(t, 1.1, result["peakWindGust"])
	assert.Equal(t, 1.778, result["cumulativeRainfall"])

	payload4 := []byte{0x03, 0x64}
	result, err = decodeSenseCapPayload(payload4)
	require.Nil(t, err)
	assert.Equal(t, 100, result["batteryLevel"])

	payload5 := []byte{0x04, 0x64, 0x01, 0x01, 0x01, 0x03, 0x00, 0x0A, 0x05, 0xA0}
	result, err = decodeSenseCapPayload(payload5)
	require.Nil(t, err)
	assert.Equal(t, 100, result["batteryLevel"])
	assert.Equal(t, "1.1", result["hardwareVersion"])
	assert.Equal(t, "1.3", result["softwareVersion"])
	assert.Equal(t, 10, result["uplinkInterval"])

	payload6 := []byte{0x05, 0x00, 0x0A, 0x05, 0xA0}
	result, err = decodeSenseCapPayload(payload6)
	require.Nil(t, err)
	assert.Equal(t, 10, result["uplinkInterval"])

	payload7 := []byte{0x06, 0xFF}
	_, err = decodeSenseCapPayload(payload7)
	require.Nil(t, err)

	combined := append(append(append(append(append(append(payload1, payload2...), payload3...), payload4...), payload5...), payload6...), payload7...)
	result, err = decodeSenseCapPayload(combined)
	require.Nil(t, err)
	assert.Equal(t, 27.7, result["temperature"])
	assert.Equal(t, 65, result["humidity"])
	assert.Equal(t, 94, result["lightIntensity"])
	assert.Equal(t, 0.5, result["lightUV"])
	assert.Equal(t, 1.7, result["windSpeed"])
	assert.Equal(t, 342, result["windDirection"])
	assert.Equal(t, 0.254, result["rainIntensity"])
	assert.Equal(t, 99870, result["barometricPressure"])
	assert.Equal(t, 1.1, result["peakWindGust"])
	assert.Equal(t, 1.778, result["cumulativeRainfall"])
	assert.Equal(t, 100, result["batteryLevel"])
	assert.Equal(t, "1.1", result["hardwareVersion"])
	assert.Equal(t, "1.3", result["softwareVersion"])
	assert.Equal(t, 10, result["uplinkInterval"])
}

func TestSenseCapDecodeError(t *testing.T) {
	for frameId := byte(0); frameId < 255; frameId++ {
		payload := []byte{frameId}
		_, err := decodeSenseCapPayload(payload)
		assert.NotNil(t, err)
	}
	emptyPayload := []byte{}
	result, err := decodeSenseCapPayload(emptyPayload)
	require.Nil(t, err)
	assert.Empty(t, result)
}
