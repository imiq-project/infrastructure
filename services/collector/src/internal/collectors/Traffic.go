package collectors

import (
	"fmt"
	"imiq/collector/internal/config"
	"math/rand"
	"time"
)

type sensors struct{}

func TrafficCollector() (config.Collector, error) {
	return &sensors{}, nil
}

func (collector sensors) Name() string {
	return "Traffic"
}

func (collector sensors) Fetch(coord config.Coord) (map[string]any, error) {
	vehiclesIn := rand.Intn(70) + 5
	vehiclesOut := rand.Intn(70) + 5
	avgSpeed := rand.Intn(50) + 20
	cyclists := rand.Intn(20) + 1
	pedestrians := rand.Intn(30)

	return map[string]any{
		"type": "Traffic",

		"vehiclesIn": map[string]any{
			"type":  "Integer",
			"value": vehiclesIn,
		},
		"vehiclesOut": map[string]any{
			"type":  "Integer",
			"value": vehiclesOut,
		},
		"avgSpeed": map[string]any{
			"type":  "Integer",
			"value": avgSpeed,
		},
		"cyclists": map[string]any{
			"type":  "Integer",
			"value": cyclists,
		},
		"pedestrians": map[string]any{
			"type":  "Integer",
			"value": pedestrians,
		},
		"location": map[string]any{
			"type":  "geo:point",
			"value": fmt.Sprintf("%f, %f", coord.Lat, coord.Lon),
		},
		"timestamp": map[string]any{
			"type":  "DateTime",
			"value": time.Now().Format(time.RFC3339),
		},
	}, nil
}
