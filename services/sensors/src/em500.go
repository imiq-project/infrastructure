package main

import (
	"encoding/binary"
	"fmt"
)

func decodeEm500Payload(payload []byte) (map[string]any, error) {
	var result map[string]any = map[string]any{}
	for len(payload) > 0 {
		if len(payload) < 2 {
			return nil, fmt.Errorf("frame too short")
		}
		channelId := payload[0]
		typeId := payload[1]

		if channelId == 0xff {
			// device info, skip
			switch typeId {
			case 0x01, 0x09, 0x0a, 0x0b, 0x0f:
				if len(payload) < 3 {
					return nil, fmt.Errorf("frame too short")
				}
				payload = payload[3:]
			case 0x16:
				if len(payload) < 18 {
					return nil, fmt.Errorf("frame too short")
				}
				payload = payload[18:]
			default:
				return nil, fmt.Errorf("invalid type id %d", typeId)
			}
		} else {
			// measurement, parse
			switch typeId {
			case 0x67:
				if len(payload) < 4 {
					return nil, fmt.Errorf("frame too short")
				}
				result["temperature"] = float64(binary.LittleEndian.Uint16(payload[2:4])) / 10
				payload = payload[4:]
			case 0x68:
				if len(payload) < 3 {
					return nil, fmt.Errorf("frame too short")
				}
				result["humidity"] = float64(payload[2]) / 2
				payload = payload[3:]
			case 0x73:
				if len(payload) < 4 {
					return nil, fmt.Errorf("frame too short")
				}
				result["airPressure"] = float64(binary.LittleEndian.Uint16(payload[2:4])) / 10
				payload = payload[4:]
			case 0x75:
				if len(payload) < 3 {
					return nil, fmt.Errorf("frame too short")
				}
				result["battery"] = int(payload[2])
				payload = payload[3:]
			case 0x7d:
				if len(payload) < 4 {
					return nil, fmt.Errorf("frame too short")
				}
				result["co2"] = int(binary.LittleEndian.Uint16(payload[2:4]))
				payload = payload[4:]
			case 0x77, 0x7b, 0x7f, 0x82, 0xca:
				// 2 bytes, not used
				if len(payload) < 4 {
					return nil, fmt.Errorf("frame too short")
				}
				payload = payload[4:]
			case 0x94:
				// 4 bytes  not used
				if len(payload) < 6 {
					return nil, fmt.Errorf("frame too short")
				}
				payload = payload[6:]
			default:
				return nil, fmt.Errorf("invalid type id %d", typeId)
			}
		}
	}
	return result, nil
}
