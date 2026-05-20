package collectors

import (
	"encoding/csv"
	"fmt"
	"imiq/collector/internal/config"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dhconnelly/rtreego"
)

type Point struct {
	Lat float64
	Lon float64
}

// Segment implements rtreego.Spatial
type Segment struct {
	A, B   Point
	Street *ActiveStreet
}

// bounding box required by rtreego
func (s Segment) Bounds() rtreego.Rect {
	minX := math.Min(s.A.Lon, s.B.Lon)
	maxX := math.Max(s.A.Lon, s.B.Lon)
	minY := math.Min(s.A.Lat, s.B.Lat)
	maxY := math.Max(s.A.Lat, s.B.Lat)

	// 2D rectangle
	rect, _ := rtreego.NewRect(
		rtreego.Point{minX, minY},
		[]float64{maxX - minX, maxY - minY},
	)

	return rect
}

func pointSegmentDist(px, py, ax, ay, bx, by float64) float64 {
	abx := bx - ax
	aby := by - ay
	apx := px - ax
	apy := py - ay

	abLen2 := abx*abx + aby*aby
	if abLen2 == 0 {
		return math.Hypot(px-ax, py-ay)
	}

	t := (apx*abx + apy*aby) / abLen2
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}

	cx := ax + t*abx
	cy := ay + t*aby

	return math.Hypot(px-cx, py-cy)
}

type RunningMean struct {
	count int
	mean  float64
}

func (rm *RunningMean) Add(value float64) {
	rm.count++
	rm.mean += (value - rm.mean) / float64(rm.count)
}

func (rm *RunningMean) Get() float64 {
	return rm.mean
}

func (rm *RunningMean) Reset() {
	rm.count = 0
	rm.mean = 0
}

type ActiveStreet struct {
	ID       string
	Speed    RunningMean
	MaxSpeed int
}

type TrafficCollector struct {
	tree     *rtreego.Rtree
	streets  map[string]*ActiveStreet
	endpoint string
	apiKey   string
}

func NewTrafficCollector() (config.Collector, error) {
	apiKey, exists := os.LookupEnv("LASA2X_API_KEY")
	if !exists {
		return nil, fmt.Errorf("environment variable LASA2X_API_KEY not set")
	}
	endpoint, exists := os.LookupEnv("LASA2X_ENDPOINT")
	if !exists {
		return nil, fmt.Errorf("environment variable LASA2X_ENDPOINT not set")
	}
	return TrafficCollector{
		tree:     rtreego.NewTree(2, 25, 50),
		streets:  make(map[string]*ActiveStreet),
		endpoint: endpoint,
		apiKey:   apiKey,
	}, nil
}

func segmentStringToGeoJSON(segmentsStr string) any {
	var geojsonSegments []any

	segments := strings.Split(segmentsStr, ";")
	for _, coordsStr := range segments {
		coords := strings.Split(coordsStr, ",")
		var points []any
		for idx := 0; idx < len(coords); idx += 2 {
			lat, err1 := strconv.ParseFloat(coords[idx], 64)
			lon, err2 := strconv.ParseFloat(coords[idx+1], 64)
			if err1 != nil || err2 != nil {
				return nil
			}
			points = append(points, []float64{lon, lat})
		}
		geojsonSegments = append(geojsonSegments, points)
	}
	return map[string]any{
		"type":        "MultiLineString",
		"coordinates": geojsonSegments,
	}
}

func (collector TrafficCollector) Setup(locations []config.Location) {
	for _, loc := range locations {
		collector.streets[loc.ID] = &ActiveStreet{
			ID:       loc.ID,
			Speed:    RunningMean{},
			MaxSpeed: loc.Metadata["speed_limit"].(int),
		}
	}
	log.Printf("Initialized %d streets", len(collector.streets))

	for _, loc := range locations {
		segments_str := loc.Metadata["segments"].(string)
		segments := strings.Split(segments_str, ";")
		for _, coords_str := range segments {
			coords := strings.Split(coords_str, ",")
			for idx := 0; idx < len(coords)-2; idx += 2 {
				lat1, _ := strconv.ParseFloat(coords[idx], 64)
				lon1, _ := strconv.ParseFloat(coords[idx+1], 64)
				lat2, _ := strconv.ParseFloat(coords[idx+2], 64)
				lon2, _ := strconv.ParseFloat(coords[idx+3], 64)
				tree_segment := Segment{
					A:      Point{Lat: lat1, Lon: lon1},
					B:      Point{Lat: lat2, Lon: lon2},
					Street: collector.streets[loc.ID],
				}
				collector.tree.Insert(tree_segment)
			}
		}
	}
	log.Printf("R-tree initialized with %d segments", collector.tree.Size())
}

func (collector TrafficCollector) Name() string {
	return "Traffic"
}

func ParseCams(reader *csv.Reader, tree *rtreego.Rtree) error {
	// Skip header: date;id;lat;lon;heading;speed_kmh
	reader.Comma = ';'
	_, err := reader.Read()
	if err != nil {
		return err
	}

	numCams := 0
	tick := time.Now()
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		lat, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			return err
		}
		lon, err := strconv.ParseFloat(record[3], 64)
		if err != nil {
			return err
		}
		speed, err := strconv.ParseFloat(record[5], 64)
		if err != nil {
			return err
		}
		searchPoint := rtreego.Point{lon, lat}
		rect, _ := rtreego.NewRect(searchPoint, []float64{.001, .001})

		// search nearest segments
		results := tree.SearchIntersect(rect)
		if len(results) == 0 {
			continue
		}

		bestDist := math.MaxFloat64
		var best Segment

		px, py := searchPoint[0], searchPoint[1]

		for i, item := range results {
			if i > 20 {
				break
			}

			s := item.(Segment)

			d := pointSegmentDist(
				px, py,
				s.A.Lon, s.A.Lat,
				s.B.Lon, s.B.Lat,
			)

			if d < bestDist {
				bestDist = d
				best = s
			}
		}
		best.Street.Speed.Add(speed)
		numCams++
	}
	tock := time.Since(tick)
	log.Printf("Finished parsing %d CAMs in %s", numCams, tock)
	return nil
}

func FetchCsv(endpoint string, apiKey string) (*csv.Reader, error) {
	// returns cams (car awareness messages) from the last 15 minutes
	url := endpoint + "/v1/cam/messages/recent"

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("x-api-key", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bad status: %s, body: %s", resp.Status, string(body))
	}

	// TODO: defer body close? But then how to read it after returning?

	return csv.NewReader(resp.Body), nil
}

func (collector TrafficCollector) BeforeFetch() error {
	for _, street := range collector.streets {
		street.Speed.Reset()
	}
	// file, err := os.Open("/app/src/internal/collectors/testdata/cams_evening.csv")
	// if err != nil {
	// 	return err
	// }
	// defer file.Close()
	// reader := csv.NewReader(file)
	reader, err := FetchCsv(collector.endpoint, collector.apiKey)
	if err != nil {
		return err
	}
	err = ParseCams(reader, collector.tree)
	if err != nil {
		return err
	}
	return nil
}

func (collector TrafficCollector) AfterFetch() error {
	return nil
}

func (collector TrafficCollector) Fetch(loc config.Location) (map[string]any, error) {
	var value any

	speed := collector.streets[loc.ID].Speed
	if speed.count == 0 {
		value = nil
	} else {
		value = speed.Get()
	}

	return map[string]any{
		"type": "Traffic",
		"avgSpeed": map[string]any{
			"type":  "Float",
			"value": value,
		},
		"location": map[string]any{
			"type":  "geo:point",
			"value": fmt.Sprintf("%f, %f", loc.Coord.Lat, loc.Coord.Lon),
		},
		"outline": map[string]any{
			"type": "StructuredValue",
			// geo:json is not allowed as only one geocoded value permitted, ie location
			"value": segmentStringToGeoJSON(loc.Metadata["segments"].(string)),
		},
		"speedLimit": map[string]any{
			"type":  "Integer",
			"value": loc.Metadata["speed_limit"].(int),
		},
	}, nil
}
