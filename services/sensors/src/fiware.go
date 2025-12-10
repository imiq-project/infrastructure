package main

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
