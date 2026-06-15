package collectors

import (
	"encoding/json"
	"fmt"
	"imiq/collector/internal/config"
	"io"
	"net/http"
)

type ChargingResponse struct {
	Status  int             `json:"status"`
	Payload ChargingPayload `json:"payload"`
}

type ChargingPayload struct {
	Connectors []Connector `json:"connectors"`
}

type Connector struct {
	OcppID         int    `json:"ocppId"`
	EvseID         string `json:"evseId"`
	MaxPowerInWatt int    `json:"maxPowerInWatt"`
	DisplayName    string `json:"displayName"`
	Type           string `json:"type"`
	CurrentType    string `json:"currentType"`
	Status         string `json:"status"`
}

type ChargingCollector struct{}

func (collector ChargingCollector) fetchChargingStation(csname string) (error, int, int) {
	url := "https://map.eround.de/api/chargemap/rest/public/marker/24d89345-0ac5-40e4-a982-a501f85758cc?csName=" + csname

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch charging station: %w", err), 0, 0
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err), 0, 0
	}

	var result ChargingResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err), 0, 0
	}

	availableConnectors := 0
	for _, connector := range result.Payload.Connectors {
		if connector.Status == "AVAILABLE" {
			availableConnectors++
		}
	}

	return nil, availableConnectors, len(result.Payload.Connectors)
}

func (collector ChargingCollector) Fetch(loc config.Location) (map[string]any, error) {

	err, available, total := collector.fetchChargingStation(loc.Metadata["csname"].(string))
	if err != nil {
		return nil, err
	}

	response := map[string]any{
		"type": "EVChargingStation",
		"location": map[string]any{
			"type":  "geo:point",
			"value": fmt.Sprintf("%f, %f", loc.Coord.Lat, loc.Coord.Lon),
		},
		"name": map[string]any{
			"type":  "String",
			"value": loc.Name,
		},
		"availableCapacity": map[string]any{
			"type":  "Integer",
			"value": available,
		},
		"capacity": map[string]any{
			"type":  "Integer",
			"value": total,
		},
	}

	return response, nil
}

func (collector ChargingCollector) Setup(locations []config.Location) {}

func (collector ChargingCollector) BeforeFetch() error {
	return nil
}

func (collector ChargingCollector) AfterFetch() error {
	return nil
}

func (collector ChargingCollector) Name() string {
	return "Charging"
}

func NewChargingCollector() (config.Collector, error) {
	return ChargingCollector{}, nil
}
