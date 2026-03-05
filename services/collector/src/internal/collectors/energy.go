package collectors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"imiq/collector/internal/config"
	"io"
	"net/http"
	"strings"
	"time"
)

type EnergyCollector struct{}

const (
	energyAPIURL = "https://kbr.ovgu.de:446/api/CustomPages/CustomPage.ashx"
	energyPID    = "5e76a437-8924-49b4-914c-aac955c68fbb" //same for all OVGU buildings at universitatplatz
)

func (collector EnergyCollector) Fetch(loc config.Location) (map[string]any, error) {

	response := map[string]any{
		"type": "Building",
		"location": map[string]any{
			"type":  "geo:point",
			"value": fmt.Sprintf("%f, %f", loc.Coord.Lat, loc.Coord.Lon),
		},
		"name": map[string]any{
			"type":  "String",
			"value": sanitize(loc.Name),
		},
	}

	for key, value := range loc.Metadata {
		if key == "metadata" || key == "resources" || key == "userKey" || key == "meter-id" || key == "obis" {
			continue
		}

		if key == "type" {
			if strVal, ok := value.(string); ok {
				response["type"] = sanitize(strVal)
			}
			continue
		}
		safeValue := value
		if strVal, ok := value.(string); ok {
			safeValue = sanitize(strVal)
		}
		response[key] = map[string]any{
			"type":  "String",
			"value": safeValue,
		}
	}

	// The energy collector expects the following metadata fields to be present in the location:
	// - userKey: the userKey to access the API
	// - meter-id: the id of the meter to fetch data for
	// - obis: the obis code of the data point to fetch (e.g. 1-0:1.8.0 for total energy consumption)

	var rawResources []any

	rawMetadata, ok := loc.Metadata["metadata"].(map[string]any)
	if ok {
		if resources, exists := rawMetadata["resources"].([]any); exists {
			rawResources = resources
		}
	}

	// if there are no resources, we can return early with the basic information we have
	if len(rawResources) == 0 {
		return response, nil
	}

	for _, rawResource := range rawResources {

		resourceMap, ok := rawResource.(map[string]any)
		if !ok {
			continue
		}

		resourceType := resourceMap["type"].(string)
		unit, _ := resourceMap["unit"].(string)
		if unit == "" {
			lowerResourceType := strings.ToLower(resourceType)
			if strings.Contains(lowerResourceType, "water") {
				unit = "m³"
			} else {
				unit = "kWh"
			}
		}
		var finalValue float64
		timestamp := time.Now().Format(time.RFC3339)

		if userKey, exists := resourceMap["userKey"].(string); exists {
			meterID, _ := resourceMap["meter-id"].(string)
			obis, _ := resourceMap["obis"].(string)
			value, t, err := fetchEnergyData(meterID, obis, userKey)
			if err != nil {
				fmt.Printf("Error fetching energy data for %s: %v\n", loc.Name, err)
				continue
			}
			finalValue = value
			if t != "" {
				timestamp = t
			}
		} else {
			continue
		}

		response[resourceType] = map[string]any{
			"type":  "Number",
			"value": finalValue,
			"metadata": map[string]any{
				"unit": map[string]any{
					"type":  "Text",
					"value": unit,
				},
				"timestamp": map[string]any{
					"type":  "DateTime",
					"value": timestamp,
				},
			},
		}
	}

	return response, nil
}

func fetchEnergyData(meterID, obis, userKey string) (float64, string, error) {
	now := time.Now()
	endTicks := now.UnixMilli()                         //current time in milliseconds since epoch, which is the format expected by the API
	startTicks := now.Add(-720 * time.Hour).UnixMilli() // take data from the last 24 hours to ensure we get a valid data point even if there are delays in data availability

	innerQuery := map[string]any{
		"id":         meterID,
		"obis":       obis,
		"fn":         0,
		"interval":   1,
		"startticks": startTicks,
		"endticks":   endTicks,
		"culture":    "de",
	}

	innerQueryBytes, err := json.Marshal(innerQuery)
	if err != nil {
		return 0, "", err
	}

	payload := map[string]any{
		"action":  "getVEDataSet",
		"culture": "de",
		"pid":     energyPID,
		"userKey": userKey,
		"query":   string(innerQueryBytes),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return 0, "", err
	}

	req, err := http.NewRequest("POST", energyAPIURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return 0, "", err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Origin", "https://kbr.ovgu.de:446")
	req.Header.Set("Referer", "https://kbr.ovgu.de:446/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	time.Sleep(1 * time.Second)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return 0, "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, bodyBytes)
	}

	var respData struct {
		Values []struct {
			Value   float64 `json:"value"`
			EndTime int64   `json:"endTime"`
		} `json:"values"`
	}

	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &respData); err != nil {
		return 0, "", fmt.Errorf("failed to parse response: %w", err)
	}
	//CASE 1: There are no values, return an entire JSON response for better debugging
	if len(respData.Values) == 0 {
		return 0, "", fmt.Errorf("API returned empty values for meter %s. Raw response from KBR: %s", meterID, string(body))
	}

	// CASE 2: There are non-zero values,  Iterate over the values in reverse order to find the most recent non-zero value
	for i := len(respData.Values) - 1; i >= 0; i-- {
		point := respData.Values[i]
		if point.Value != 0 {
			timestamp := time.Now().Format(time.RFC3339) // SAFEGUARD: for now because we are getting 1970 in response Timestamp
			if point.EndTime > 0 {
				timestamp = time.UnixMilli(point.EndTime).Format(time.RFC3339)
			}
			return point.Value, timestamp, nil
		}
	}
	// CASE 3: All values are zero, return 0 with the timestamp of the most recent data point (even if it's zero) to ensure we have a valid timestamp in the metadata and to avoid returning an error just because all values are zero (which can be a valid case, e.g. for a building that is currently not consuming any energy)
	timestamp := time.Now().Format(time.RFC3339)
	if len(respData.Values) > 0 {
		lastPoint := respData.Values[len(respData.Values)-1]
		if lastPoint.EndTime > 0 {
			timestamp = time.UnixMilli(lastPoint.EndTime).Format(time.RFC3339)
		}
	}
	return 0, timestamp, nil
}

func (collector EnergyCollector) Name() string {
	return "Building"
}

func NewEnergyCollector() (config.Collector, error) {
	return EnergyCollector{}, nil
}
