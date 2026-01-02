package main

import (
	"encoding/binary"
	"fmt"
)

func decodeSenseCapPayload(payload []byte) (map[string]any, error) {
	var result map[string]any = map[string]any{}
	for len(payload) > 0 {
		frameId := payload[0]
		switch frameId {
		case 0x01, 0x4A:
			if len(payload) < 11 {
				return nil, fmt.Errorf("frame too short")
			}
			result["temperature"] = float64(binary.BigEndian.Uint16(payload[1:3])) / 10
			result["humidity"] = int(payload[3])
			result["lightIntensity"] = int(binary.BigEndian.Uint32(payload[4:8]))
			result["lightUV"] = float64(payload[8]) / 10
			result["windSpeed"] = float64(binary.BigEndian.Uint16(payload[9:11])) / 10
			payload = payload[11:]
		case 0x02, 0x4B:
			if len(payload) < 9 {
				return nil, fmt.Errorf("frame too short")
			}
			result["windDirection"] = int(binary.BigEndian.Uint16(payload[1:3]))
			result["rainIntensity"] = float64(binary.BigEndian.Uint32(payload[3:7])) / 1000
			result["barometricPressure"] = int(binary.BigEndian.Uint16(payload[7:9])) * 10
			payload = payload[9:]
		case 0x4C:
			if len(payload) < 7 {
				return nil, fmt.Errorf("frame too short")
			}
			result["peakWindGust"] = float64(binary.BigEndian.Uint16(payload[1:3])) / 10
			result["cumulativeRainfall"] = float64(binary.BigEndian.Uint32(payload[3:7])) / 1000
			payload = payload[7:]
		case 0x03:
			if len(payload) < 2 {
				return nil, fmt.Errorf("frame too short")
			}
			result["batteryLevel"] = int(payload[1])
			payload = payload[2:]
		case 0x04:
			if len(payload) < 10 {
				return nil, fmt.Errorf("frame too short")
			}
			result["batteryLevel"] = int(payload[1])
			result["hardwareVersion"] = fmt.Sprintf("%d.%d", payload[2], payload[3])
			result["softwareVersion"] = fmt.Sprintf("%d.%d", payload[4], payload[5])
			result["uplinkInterval"] = int(binary.BigEndian.Uint16(payload[6:8]))
			// payload[8:10] denotes the gps uplink interval (customized version only)
			payload = payload[10:]
		case 0x05:
			if len(payload) < 5 {
				return nil, fmt.Errorf("frame too short")
			}
			result["uplinkInterval"] = int(binary.BigEndian.Uint16(payload[1:3]))
			// payload[3:5] denotes the gps uplink interval (customized version only)
			payload = payload[5:]
		case 0x06:
			if len(payload) < 2 {
				return nil, fmt.Errorf("frame too short")
			}
			// only contains one byte error code, discard
			payload = payload[2:]
		default:
			return nil, fmt.Errorf("invalid frame id %d", frameId)
		}
	}
	return result, nil
}
