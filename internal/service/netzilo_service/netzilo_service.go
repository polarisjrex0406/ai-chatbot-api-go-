package netzilo_service

import (
	"aidashboard/internal/config"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const netziloBaseUrl = "https://appstaging.netzilo.com/api/"

func GetGroups() ([]interface{}, error) {
	return MakeGetCall("groups", nil)
}

func CreatePostureCheck(data map[string]interface{}) (map[string]interface{}, error) {
	return MakePostCall("posture-checks", data)
}

func CreatePolicy(data map[string]interface{}) (map[string]interface{}, error) {
	return MakePostCall("policies", data)
}

func CreateProfile(data map[string]interface{}) (map[string]interface{}, error) {
	return MakePostCall("profiles", data)
}

func MakePostCall(endpoint string, data map[string]interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	bytes, err := MakeNetziloApiCall("POST", endpoint, data)

	if err != nil {
		return nil, err
	} else {
		// Unmarshal the JSON string into the map
		err := json.Unmarshal(bytes, &result)
		if err != nil {
			fmt.Println("Error unmarshaling JSON:", err)
			return nil, err
		}
	}

	return result, err
}

func MakeGetCall(endpoint string, data map[string]interface{}) ([]interface{}, error) {
	var result []interface{}
	bytes, err := MakeNetziloApiCall("GET", endpoint, data)

	if err != nil {
		return nil, err
	} else {
		// Unmarshal the JSON string into the map
		err := json.Unmarshal(bytes, &result)
		if err != nil {
			fmt.Println("Error unmarshaling JSON:", err)
			return nil, err
		}
	}

	return result, err
}

func MakeNetziloApiCall(method string, endpoint string, data map[string]interface{}) ([]byte, error) {
	apiKey := config.GetNetziloApiKey()
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, netziloBaseUrl+endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	fmt.Println(method + " " + netziloBaseUrl + endpoint)
	fmt.Println(string(payload))

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		// Read the body to log the error message
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API call failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Read the response body as a string
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return bodyBytes, nil
}
