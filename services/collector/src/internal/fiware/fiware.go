package fiware

import (
	"bytes"
	"encoding/json"
	"fmt"
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
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return nil
}

func UpdateEntity(fiwareUrl string, id string, data map[string]any) error {
	jsonData, err := json.Marshal(data)
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
	}
	return nil
}
