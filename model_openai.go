package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// OpenAI
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIModelRequestPayload struct {
	Model    string          `json:"model"`
	Messages []OpenAIMessage `json:"messages"`
}

type OpenAIModelResponsePayload struct {
	ID      string `json:"id"`
	Choices []struct {
		Message      OpenAIMessage `json:"message"`
		FinishReason string        `json:"finish_reason"`
	} `json:"choices"`
}

type OpenAIModel struct {
	modelId string
	url     string
}

func NewOpenAIModel(modelId, url string) *OpenAIModel {
	return &OpenAIModel{
		modelId: modelId,
		url:     url,
	}
}

func (m *OpenAIModel) Invoke(input ModelInput) (*ModelResponse, error) {
	// Create the JSON payload
	payload := OpenAIModelRequestPayload{
		Model: m.modelId,
	}
	payload.Messages = append(payload.Messages, OpenAIMessage{Role: "user", Content: input.Prompt})

	// Convert the payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		//fmt.Println("Error encoding JSON:", err)
		return nil, err
	}

	apiURL := m.url + "/v1/chat/completions"
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		//fmt.Println("Error creating HTTP request:", err)
		return nil, err
	}

	// Set the "Content-Type" header to "application/json"
	req.Header.Set("Content-Type", "application/json")

	// Set the "Authorization" header with the bearer token
	req.Header.Set("Authorization", "Bearer "+input.APIKey)
	//req.Header.Set("Email", input.UserId)

	// Make the API call
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		//fmt.Println("Error making API request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %s", resp.Status)
	}

	// Parse the JSON response into the APIResponse struct
	var apiResp OpenAIModelResponsePayload
	err = json.NewDecoder(resp.Body).Decode(&apiResp)
	if err != nil {
		fmt.Println("Error decoding API response:", err)
		return nil, err
	}
	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("model returned no valid responses: %v", apiResp)
	}
	response := ModelResponse{}
	response.Input = input.Prompt
	response.Output = apiResp.Choices[0].Message.Content
	return &response, err
}