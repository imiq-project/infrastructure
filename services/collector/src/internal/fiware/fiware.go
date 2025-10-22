package fiware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func CreateEntity(fiwareUrl string, id string, data map[string]any) error {
	data["id"] = id
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	resp, err := http.Post(
		fiwareUrl+"/v2/entities/",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return err
	}
	if resp.StatusCode != 201 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, bodyBytes)
	}
	return nil
}

func UpdateEntity(fiwareUrl string, id string, data map[string]any) error {
	dataCopied := map[string]any{}
	for k, v := range data {
		if k != "type" {
			dataCopied[k] = v
		}
	}

	jsonData, err := json.Marshal(dataCopied)
	if err != nil {
		return err
	}
	resp, err := http.Post(
		fiwareUrl+"/v2/entities/"+id+"/attrs",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return err
	}
	if resp.StatusCode == 404 {
		err = CreateEntity(fiwareUrl, id, data)
		if err != nil {
			return err
		}
	} else if resp.StatusCode != 204 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, bodyBytes)
	}
	return nil
}

// -----------------------------------
// create subscriptions
// -----------------------------------

func EnsureSubscription(fiwareUrl string, qlNotifyURL string, sub map[string]any) error {
	resp, err := http.Get(fiwareUrl + "/v2/subscriptions")
	if err != nil {
		return fmt.Errorf("list subscriptions failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("list subscriptions unexpected status %d: %s", resp.StatusCode, bodyBytes)
	}

	var subs []struct {
		ID            string `json:"id"`
		Notifications struct {
			HTTP struct {
				URL string `json:"url"`
			} `json:"http"`
		} `json:"notification"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&subs); err != nil {
		return fmt.Errorf("list subscriptions decode failed: %w", err)
	}

	for _, existingSub := range subs {
		if existingSub.Notifications.HTTP.URL == qlNotifyURL {
			// subscription already exists
			return nil
		}
	}

	payload := map[string]any{
		"description": "Notify QL of entity changes",
		"subject": map[string]any{
			"entities": []map[string]any{
				{"idPattern": ".*"},
			},
		},
		"notification": map[string]any{
			"http": map[string]any{
				"url":         qlNotifyURL,
				"attrsFormat": "normalized",
			},
		},
		"throttling": 1,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", fiwareUrl+"/v2/subscriptions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("create subscription failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create subscription unexpected status %d: %s", resp.StatusCode, bodyBytes)
	}
	return nil
}
