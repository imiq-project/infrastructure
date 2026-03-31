package collectors

import (
	"crypto/tls"
	"fmt"
	"imiq/collector/internal/config"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
)

type ParkingCollector struct{}

const (
	parkURL = "https://www.movi.de/parkinfo/uebersicht.shtml"
)

type ParkingInfo struct {
	AvailableSpaces int
}

func (collector ParkingCollector) Fetch(loc config.Location) (map[string]any, error) {

	data, err := scrapeFreeSpaces()
	if err != nil {
		return nil, fmt.Errorf("failed to scrape parking info: %w", err)
	}

	targetName := loc.Name
	if metadataName, ok := loc.Metadata["parking-name"].(string); ok {
		targetName = metadataName
	}

	targetName = strings.TrimSpace(targetName)

	normalize := func(s string) string {
		s = strings.ToLower(s)
		return strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
				return r
			}
			return -1
		}, s)
	}

	normTarget := normalize(targetName)
	var foundInfo ParkingInfo
	var exists bool

	for scrapedName, info := range data {
		normScraped := normalize(scrapedName)

		// Match if one contains the other (e.g., "tiefgaragedomviertel" contains "domviertel")
		if normScraped != "" && (strings.Contains(normTarget, normScraped) || strings.Contains(normScraped, normTarget)) {
			foundInfo = info
			exists = true
			break
		}
	}

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

	if totalSpace, ok := loc.Metadata["total_spots"].(int); ok {
		response["total_spots"] = map[string]any{
			"type":  "Number",
			"value": totalSpace,
		}
	}

	if exists {
		response["available_spots"] = map[string]any{
			"type":  "Number",
			"value": foundInfo.AvailableSpaces,
		}
		response["status"] = map[string]any{
			"type":  "String",
			"value": "Online",
		}
	} else {
		log.Printf("Parking info for '%s' (Normalized: '%s') not found in data\n", targetName, normTarget)
		response["available_spots"] = map[string]any{
			"type":  "Number",
			"value": 0,
		}
		response["status"] = map[string]any{
			"type":  "String",
			"value": "Offline",
		}
	}
	return response, nil
}

func scrapeFreeSpaces() (map[string]ParkingInfo, error) {
	var results map[string]ParkingInfo
	c := colly.NewCollector()

	//the website does not have a valid certificate, so we need to skip verification
	c.WithTransport(&http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	})

	c.OnHTML("table", func(e *colly.HTMLElement) {
		if strings.Contains(e.Text, "Parkanlage") && strings.Contains(e.Text, "Freie Plätze") {
			if results == nil {
				results = extractParkingSpotsFromTable(e.DOM)
			}
		}
	})
	err := c.Visit(parkURL)
	if err != nil {
		return nil, fmt.Errorf("failed to visit parking info page: %w", err)
	}
	if results == nil {
		results = make(map[string]ParkingInfo)
	}
	return results, nil
}

func extractParkingSpotsFromTable(table *goquery.Selection) map[string]ParkingInfo {
	results := make(map[string]ParkingInfo)
	table.Find("tr").Each(func(i int, row *goquery.Selection) {
		nameNode := row.Find("a")

		if nameNode.Length() == 0 {
			return
		}
		name := strings.TrimSpace(nameNode.Text())
		tdContainingATag := nameNode.Closest("td")
		emptySpaceColumn := tdContainingATag.Next()
		freestr := strings.TrimSpace(emptySpaceColumn.Text())
		if name == "" {
			return
		}
		freestr = strings.ReplaceAll(freestr, "\u00a0", "")
		freestr = strings.ReplaceAll(freestr, " ", "")
		freestr = strings.TrimSpace(freestr)
		if strings.Contains(strings.ToLower(freestr), "offline") {
			return
		}
		freeSpace, err := strconv.Atoi(freestr)
		if err == nil {
			results[name] = ParkingInfo{AvailableSpaces: freeSpace}
		}
	})
	return results
}

func (collector ParkingCollector) Name() string {
	return "Parking"
}

func NewParkingCollector() (config.Collector, error) {
	return ParkingCollector{}, nil
}
