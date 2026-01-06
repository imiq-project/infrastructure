package main

import (
	"encoding/base64"
	"encoding/json"
	"io"
)

type ProcessWebhookResponse struct {
	DeviceID string
	Payload  []byte
}

type TtnWebhookPayload struct {
	EndDeviceIDs struct {
		DeviceID       string `json:"device_id"`
		ApplicationIDs struct {
			ApplicationID string `json:"application_id"`
		} `json:"application_ids"`
		DevEUI  string `json:"dev_eui"`
		JoinEUI string `json:"join_eui"`
		DevAddr string `json:"dev_addr"`
	} `json:"end_device_ids"`

	CorrelationIDs []string `json:"correlation_ids"`
	ReceivedAt     string   `json:"received_at"`

	UplinkMessage struct {
		SessionKeyID string `json:"session_key_id"`
		FPort        int    `json:"f_port"`
		FCnt         int    `json:"f_cnt"`
		FrmPayload   string `json:"frm_payload"`

		DecodedPayload struct {
			Err      int      `json:"err"`
			Messages []string `json:"messages"`
			Payload  string   `json:"payload"`
			Valid    bool     `json:"valid"`
		} `json:"decoded_payload"`

		RxMetadata []struct {
			GatewayIDs struct {
				GatewayID string `json:"gateway_id"`
				EUI       string `json:"eui"`
			} `json:"gateway_ids"`

			Time        string  `json:"time"`
			Timestamp   uint32  `json:"timestamp"`
			RSSI        int     `json:"rssi"`
			ChannelRSSI int     `json:"channel_rssi"`
			SNR         float64 `json:"snr"`

			Location struct {
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
				Altitude  int     `json:"altitude"`
				Source    string  `json:"source"`
			} `json:"location"`

			UplinkToken  string `json:"uplink_token"`
			ChannelIndex int    `json:"channel_index"`
			GPSTime      string `json:"gps_time,omitempty"`
			ReceivedAt   string `json:"received_at"`
		} `json:"rx_metadata"`

		Settings struct {
			DataRate struct {
				Lora struct {
					Bandwidth       int    `json:"bandwidth"`
					SpreadingFactor int    `json:"spreading_factor"`
					CodingRate      string `json:"coding_rate"`
				} `json:"lora"`
			} `json:"data_rate"`

			Frequency string `json:"frequency"`
			Timestamp uint32 `json:"timestamp"`
			Time      string `json:"time"`
		} `json:"settings"`

		ReceivedAt      string `json:"received_at"`
		Confirmed       bool   `json:"confirmed"`
		ConsumedAirtime string `json:"consumed_airtime"`

		NetworkIDs struct {
			NetID          string `json:"net_id"`
			NSID           string `json:"ns_id"`
			TenantID       string `json:"tenant_id"`
			ClusterID      string `json:"cluster_id"`
			ClusterAddress string `json:"cluster_address"`
		} `json:"network_ids"`

		LastBatteryPercentage struct {
			FCnt       int    `json:"f_cnt"`
			Value      int    `json:"value"`
			ReceivedAt string `json:"received_at"`
		} `json:"last_battery_percentage"`
	} `json:"uplink_message"`
}

func ProcessWebhook(r io.Reader) (ProcessWebhookResponse, error) {
	var msg TtnWebhookPayload
	err := json.NewDecoder(r).Decode(&msg)
	if err != nil {
		return ProcessWebhookResponse{}, err
	}
	payload, err := base64.StdEncoding.DecodeString(msg.UplinkMessage.FrmPayload)
	if err != nil {
		return ProcessWebhookResponse{}, err
	}
	return ProcessWebhookResponse{
		DeviceID: msg.EndDeviceIDs.DeviceID,
		Payload:  payload,
	}, nil
}
