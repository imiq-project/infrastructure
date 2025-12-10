package main

import (
	"log"
	"net/http"
	"os"
)

const (
	ID_KEY       = "wsid"
	PASSWORD_KEY = "wspw"
)

func main() {
	expectedPassword, exists := os.LookupEnv("WEATHER_STATION_PASSWORD")
	if !exists {
		panic("Cannot lookup password")
	}

	http.HandleFunc("/data/upload.php", func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		_, exists := params[ID_KEY]
		if !exists {
			http.Error(w, ID_KEY+" is missing", http.StatusBadRequest)
			return
		}
		delete(params, ID_KEY)
		password, exists := params[PASSWORD_KEY]
		if !exists {
			http.Error(w, PASSWORD_KEY+" is missing", http.StatusBadRequest)
			return
		}
		delete(params, PASSWORD_KEY)

		if password[0] != expectedPassword {
			http.Error(w, "invalid password", http.StatusBadRequest)
		}

		entity := map[string]any{
			"type": "WeatherObserved",
			"humidity": map[string]any{
				"type":  "float",
				"value": params.Get("t1hum"),
			},
			"temperature": map[string]any{
				"type":  "float",
				"value": params.Get("t1tem"),
			},
			"location": map[string]any{
				"type":  "geo:point",
				"value": "52.141234471041685, 11.654583803189286",
			},
			"additionalData": params,
		}
		err := UpdateEntity("http://orion:1026", "Sensor:Weather:ImiqOffice", entity)
		if err != nil {
			log.Println("Failed to update weather data: ", err)
		}
	})

	log.Println("Listening on :80")
	log.Fatal(http.ListenAndServe(":80", nil))
}
