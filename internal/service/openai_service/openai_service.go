package openai_service

import (
	"aidashboard/internal/config"
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

const openAIURL = "https://api.openai.com/v1/chat/completions"

func CallOpenAI(prompt []byte) ([]byte, error) {
	apiKey := config.GetOpenAiApiKey()

	req, err := http.NewRequest("POST", openAIURL, bytes.NewBuffer(prompt))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Use a map to decode the JSON response
	var openAIResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&openAIResponse); err != nil {
		return nil, err
	}

	// Extract the choices from the response
	if choices, ok := openAIResponse["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				// if content, ok := message["tool_calls"].(map[string]interface{}); ok {
				// }
				respJSON, err := json.MarshalIndent(message, "", "  ")
				if err != nil {
					log.Fatalf("Error marshaling person to JSON: %v", err)
				}
				return respJSON, nil // Return the content as a string
			}
		}
	}

	return nil, nil
}
