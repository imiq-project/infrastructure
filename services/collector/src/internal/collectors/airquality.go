package collectors

import (
	"encoding/csv"
	"fmt"
	"imiq/collector/internal/config"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type AirQualityCollector struct{}

func fetchCsv(stationId string, component string) (int, error) {
	// Build URL with query parameters
	params := url.Values{}
	params.Add("id", stationId)
	params.Add("component", component)
	params.Add("from", "1d-ago")
	url := "https://lupo-cloud.de/st-air-app/csv-export/csv/" + stationId + "_" + component + ".csv?" + params.Encode()

	// // Fetch the CSV file
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("bad status: %s", resp.Status)
	}

	// Parse csv
	reader := csv.NewReader(resp.Body)
	reader.Comma = ';'
	var lastRow []string
	numRows := 0
	for {
		row, err := reader.Read()
		if err != nil {
			// EOF reached
			break
		}
		numRows = numRows + 1
		if numRows > 50 {
			return 0, fmt.Errorf("csv seems to be too large")
		}
		lastRow = row
	}
	if lastRow == nil {
		return 0, fmt.Errorf("csv is empty")
	}
	if len(lastRow) != 2 {
		return 0, fmt.Errorf("last row must have 2 cells, but has %d", len(lastRow))
	}
	num, err := strconv.Atoi(lastRow[1])
	if err != nil {
		return 0, err
	}
	return num, nil
}

func (collector AirQualityCollector) Fetch(loc config.Location) (map[string]any, error) {
	result := map[string]any{
		"type": "AirQuality",
		"location": map[string]any{
			"type":  "geo:point",
			"value": fmt.Sprintf("%f, %f", loc.Coord.Lat, loc.Coord.Lon),
		},
	}
	stationId := loc.Metadata["stationId"].(string)
	components := strings.Split(loc.Metadata["components"].(string), ",")
	for _, component := range components {
		num, err := fetchCsv(stationId, component)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch %s: %s", component, err)
		}
		result[component] = map[string]any{
			"type":  "Integer",
			"value": num,
		}
	}

	return result, nil
}

func (collector AirQualityCollector) Name() string {
	return "AirQuality"
}

func NewAirQualityCollector() (config.Collector, error) {
	return AirQualityCollector{}, nil
}
