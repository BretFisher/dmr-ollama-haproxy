package main

import (
	"encoding/json"
	"os"
	"testing"

	"dmr-models-convert/pkg/converter"
)

func TestSaveOllamaResponse(t *testing.T) {
	response := converter.OllamaResponse{
		Models: []converter.OllamaModel{
			{
				Name:       "test-model",
				Model:      "test-model",
				ModifiedAt: "2025-01-01T00:00:00Z",
				Size:       1024,
				Digest:     "test-digest",
				Details: converter.OllamaDetails{
					ParentModel:       "",
					Format:            "gguf",
					Family:            "llama",
					Families:          []string{"llama"},
					ParameterSize:     "1B",
					QuantizationLevel: "F16",
				},
			},
		},
	}

	tempFile := "test-output.json"
	defer os.Remove(tempFile)

	err := saveOllamaResponse(response, tempFile)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify file was created and contains valid JSON
	data, err := os.ReadFile(tempFile)
	if err != nil {
		t.Errorf("Expected to read file, got error %v", err)
	}

	var parsedResponse converter.OllamaResponse
	err = json.Unmarshal(data, &parsedResponse)
	if err != nil {
		t.Errorf("Expected valid JSON, got error %v", err)
	}

	if len(parsedResponse.Models) != 1 {
		t.Errorf("Expected 1 model in parsed response, got %d", len(parsedResponse.Models))
	}

	if parsedResponse.Models[0].Name != "test-model" {
		t.Errorf("Expected model name 'test-model', got '%s'", parsedResponse.Models[0].Name)
	}
}

func TestPrintOllamaResponse(t *testing.T) {
	response := converter.OllamaResponse{
		Models: []converter.OllamaModel{
			{
				Name:       "test-model",
				Model:      "test-model",
				ModifiedAt: "2025-01-01T00:00:00Z",
				Size:       1024,
				Digest:     "test-digest",
				Details: converter.OllamaDetails{
					ParentModel:       "",
					Format:            "gguf",
					Family:            "llama",
					Families:          []string{"llama"},
					ParameterSize:     "1B",
					QuantizationLevel: "F16",
				},
			},
		},
	}

	err := printOllamaResponse(response)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestSaveOllamaResponseInvalidPath(t *testing.T) {
	response := converter.OllamaResponse{
		Models: []converter.OllamaModel{},
	}

	// Try to save to an invalid path
	err := saveOllamaResponse(response, "/invalid/path/that/does/not/exist/test.json")
	if err == nil {
		t.Error("Expected error for invalid path, got nil")
	}
}
