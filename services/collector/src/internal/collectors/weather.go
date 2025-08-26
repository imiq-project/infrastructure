package collectors

import (
	"encoding/json"
	"fmt"
	"imiq/collector/internal/config"
	"io"
	"net/http"
	"net/url"
	"os"
)

// OpenWeatherMap data types

type WeatherCoord struct {
	Lon float64 `json:"lon"`
	Lat float64 `json:"lat"`
}

type WeatherResponse struct {
	Coord      WeatherCoord `json:"coord"`
	Weather    []Weather    `json:"weather"`
	Base       string       `json:"base"`
	Main       Main         `json:"main"`
	Visibility int          `json:"visibility"`
	Wind       Wind         `json:"wind"`
	Clouds     Clouds       `json:"clouds"`
	Dt         int64        `json:"dt"`
	Sys        Sys          `json:"sys"`
	Timezone   int          `json:"timezone"`
	ID         int          `json:"id"`
	Name       string       `json:"name"`
	Cod        int          `json:"cod"`
}

type Weather struct {
	ID          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type Main struct {
	Temp      float64 `json:"temp"`
	FeelsLike float64 `json:"feels_like"`
	TempMin   float64 `json:"temp_min"`
	TempMax   float64 `json:"temp_max"`
	Pressure  int     `json:"pressure"`
	Humidity  int     `json:"humidity"`
	SeaLevel  int     `json:"sea_level"`
	GrndLevel int     `json:"grnd_level"`
}

type Wind struct {
	Speed float64 `json:"speed"`
	Deg   int     `json:"deg"`
	Gust  float64 `json:"gust"`
}

type Clouds struct {
	All int `json:"all"`
}

type Sys struct {
	Sunrise int64 `json:"sunrise"`
	Sunset  int64 `json:"sunset"`
}

func fetch(baseUrl string, params map[string]string) ([]byte, error) {
	p := url.Values{}
	for key, value := range params {
		p.Add(key, value)
	}

	resp, err := http.Get(baseUrl + "?" + p.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

type OpenWeatherMapCollector struct {
	apiKey string
}

func NewOpenWeatherMapCollector() (config.Collector, error) {
	apiKey, exists := os.LookupEnv("OWM_API_KEY")
	if !exists {
		return nil, fmt.Errorf("environment variable OWM_API_KEY not set")
	}
	return &OpenWeatherMapCollector{apiKey}, nil
}

func (collector OpenWeatherMapCollector) Name() string {
	return "OpenWeatherMap"
}

func (collector OpenWeatherMapCollector) Fetch(coord config.Coord) (map[string]any, error) {
	rsp, err := fetch("https://api.openweathermap.org/data/2.5/weather", map[string]string{
		"lat":   fmt.Sprintf("%.5f", coord.Lat),
		"lon":   fmt.Sprintf("%.5f", coord.Lon),
		"appid": collector.apiKey,
		"units": "metric",
	})

	if err != nil {
		return nil, err
	}

	var weather WeatherResponse
	if err := json.Unmarshal(rsp, &weather); err != nil {
		return nil, err
	}

	return map[string]any{
		"temperature": map[string]any{
			"value": weather.Main.Temp,
			"type":  "Float",
		},
		"humidity": map[string]any{
			"value": weather.Main.Humidity,
			"type":  "Float",
		},
		"location": map[string]any{
			"type":  "geo:point",
			"value": fmt.Sprintf("%f, %f", coord.Lat, coord.Lon),
		},
	}, nil
}
