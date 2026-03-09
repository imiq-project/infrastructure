package collectors

import (
	"encoding/json"
	"fmt"
	"imiq/collector/internal/config"
	"net/http"
)

type Station struct {
	UUID       string       `json:"uuid"`
	Number     string       `json:"number"`
	ShortName  string       `json:"shortname"`
	LongName   string       `json:"longname"`
	Km         float64      `json:"km"`
	Agency     string       `json:"agency"`
	Longitude  float64      `json:"longitude"`
	Latitude   float64      `json:"latitude"`
	Water      Water        `json:"water"`
	TimeSeries []TimeSeries `json:"timeseries"`
}

type Water struct {
	ShortName string `json:"shortname"`
	LongName  string `json:"longname"`
}

type TimeSeries struct {
	ShortName          string             `json:"shortname"`
	LongName           string             `json:"longname"`
	Unit               string             `json:"unit"`
	Equidistance       int                `json:"equidistance"`
	CurrentMeasurement CurrentMeasurement `json:"currentMeasurement"`
	GaugeZero          GaugeZero          `json:"gaugeZero"`
}

type CurrentMeasurement struct {
	Timestamp   string  `json:"timestamp"`
	Value       float64 `json:"value"`
	StateMnwMhw string  `json:"stateMnwMhw"`
	StateNswHsw string  `json:"stateNswHsw"`
}

type GaugeZero struct {
	Unit      string  `json:"unit"`
	Value     float64 `json:"value"`
	ValidFrom string  `json:"validFrom"`
}

type WaterLevelCollector struct{}

func (collector WaterLevelCollector) Fetch(loc config.Location) (map[string]any, error) {
	uuid := loc.Metadata["uuid"].(string)
	url := "https://www.pegelonline.wsv.de/webservices/rest-api/v2/stations/" + uuid + ".json?includeTimeseries=true&includeCurrentMeasurement=true"

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var station Station
	err = json.NewDecoder(resp.Body).Decode(&station)
	if err != nil {
		return nil, err
	}

	entity := map[string]any{
		"type": "WaterLevel",
		"location": map[string]any{
			"type":  "geo:point",
			"value": fmt.Sprintf("%f, %f", station.Latitude, station.Longitude),
		},
	}

	mapping := map[string]string{
		"W": "level",
		"Q": "volume",
	}

	for _, entry := range station.TimeSeries {
		fieldName, known := mapping[entry.ShortName]
		if known {
			entity[fieldName] = map[string]any{
				"type":  "Number",
				"value": entry.CurrentMeasurement.Value,
			}
		}
	}

	return entity, nil
}

func (collector WaterLevelCollector) Name() string {
	return "WaterLevel"
}

func NewWaterLevelCollector() (config.Collector, error) {
	return WaterLevelCollector{}, nil
}
