package collectors

import (
	"fmt"
	"imiq/collector/internal/config"
	"math/rand"
)

type ParkingCollector struct{}

func (collector ParkingCollector) Fetch(loc config.Location) (map[string]any, error) {
	const total = 20

	return map[string]any{
		"type": "Parking",
		"freeSpaces": map[string]any{
			"type":  "Integer",
			"value": rand.Intn(total),
		},
		"totalSpaces": map[string]any{
			"type":  "Integer",
			"value": total,
		},
		"location": map[string]any{
			"type":  "geo:point",
			"value": fmt.Sprintf("%f, %f", loc.Coord.Lat, loc.Coord.Lon),
		},
	}, nil
}

func (collector ParkingCollector) Name() string {
	return "Parking"
}

func NewParkingCollector() (config.Collector, error) {
	return ParkingCollector{}, nil
}
