package converter

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewConverter(t *testing.T) {
	conv := NewConverter()
	if conv == nil {
		t.Error("Expected converter to be created, got nil")
	}
	if conv.client == nil {
		t.Error("Expected HTTP client to be initialized, got nil")
	}
}

func TestNewConverterWithClient(t *testing.T) {
	customClient := &http.Client{Timeout: 10 * time.Second}
	conv := NewConverterWithClient(customClient)
	if conv == nil {
		t.Error("Expected converter to be created, got nil")
	}
	if conv.client != customClient {
		t.Error("Expected custom HTTP client to be used")
	}
}

func TestConvertFromURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
			{
				"id": "sha256:test1",
				"tags": ["model1"],
				"created": 1745698622,
				"config": {
					"format": "gguf",
					"quantization": "F16",
					"parameters": "1B",
					"architecture": "llama",
					"size": "1 GiB"
				}
			}
		]`))
	}))
	defer server.Close()

	conv := NewConverter()
	response, err := conv.ConvertFromURL(server.URL + "/models")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(response.Models) != 1 {
		t.Errorf("Expected 1 model, got %d", len(response.Models))
	}

	if response.Models[0].Name != "model1" {
		t.Errorf("Expected model name 'model1', got '%s'", response.Models[0].Name)
	}
}

func TestConvertFromJSON(t *testing.T) {
	jsonData := []byte(`[
		{
			"id": "sha256:test1",
			"tags": ["model1"],
			"created": 1745698622,
			"config": {
				"format": "gguf",
				"quantization": "F16",
				"parameters": "1B",
				"architecture": "llama",
				"size": "1 GiB"
			}
		}
	]`)

	conv := NewConverter()
	response, err := conv.ConvertFromJSON(jsonData)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(response.Models) != 1 {
		t.Errorf("Expected 1 model, got %d", len(response.Models))
	}

	if response.Models[0].Name != "model1" {
		t.Errorf("Expected model name 'model1', got '%s'", response.Models[0].Name)
	}
}

func TestConvertFromJSONInvalid(t *testing.T) {
	jsonData := []byte(`invalid json`)

	conv := NewConverter()
	_, err := conv.ConvertFromJSON(jsonData)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestFetchDMRModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
			{
				"id": "sha256:test1",
				"tags": ["model1"],
				"created": 1745698622,
				"config": {
					"format": "gguf",
					"quantization": "F16",
					"parameters": "1B",
					"architecture": "llama",
					"size": "1 GiB"
				}
			}
		]`))
	}))
	defer server.Close()

	conv := NewConverter()
	models, err := conv.FetchDMRModels(server.URL + "/models")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(models) != 1 {
		t.Errorf("Expected 1 model, got %d", len(models))
	}

	if models[0].ID != "sha256:test1" {
		t.Errorf("Expected ID 'sha256:test1', got '%s'", models[0].ID)
	}
}

func TestFetchDMRModelsError(t *testing.T) {
	conv := NewConverter()
	_, err := conv.FetchDMRModels("http://invalid-url-that-does-not-exist.com/models")
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}
}

func TestFetchDMRModelsHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	conv := NewConverter()
	_, err := conv.FetchDMRModels(server.URL + "/models")
	if err == nil {
		t.Error("Expected error for HTTP 500, got nil")
	}
}

func TestConvertDMRToOllama(t *testing.T) {
	dmrModels := []DMRModel{
		{
			ID:      "sha256:test1",
			Tags:    []string{"model1"},
			Created: 1745698622,
			Config: DMRConfig{
				Format:       "gguf",
				Quantization: "F16",
				Parameters:   "1B",
				Architecture: "llama",
				Size:         "1 GiB",
			},
		},
		{
			ID:      "sha256:test2",
			Tags:    []string{"model2"},
			Created: 1745698623,
			Config: DMRConfig{
				Format:       "gguf",
				Quantization: "Q4",
				Parameters:   "2B",
				Architecture: "phi3",
				Size:         "2 GiB",
			},
		},
	}

	conv := NewConverter()
	result := conv.ConvertDMRToOllama(dmrModels)

	if len(result.Models) != 2 {
		t.Errorf("Expected 2 models, got %d", len(result.Models))
	}

	if result.Models[0].Name != "model1" {
		t.Errorf("Expected first model name 'model1', got '%s'", result.Models[0].Name)
	}

	if result.Models[1].Details.Family != "phi3" {
		t.Errorf("Expected second model family 'phi3', got '%s'", result.Models[1].Details.Family)
	}
}

func TestConvertSingleModelEmptyTags(t *testing.T) {
	dmrModel := DMRModel{
		ID:      "sha256:test123",
		Tags:    []string{}, // Empty tags
		Created: 1745698622,
		Config: DMRConfig{
			Format:       "gguf",
			Quantization: "F16",
			Parameters:   "1B",
			Architecture: "llama",
			Size:         "1 GiB",
		},
	}

	conv := NewConverter()
	result := conv.convertSingleModel(dmrModel)

	// Should use digest as model name when tags are empty
	expectedName := "test123" // digest without sha256: prefix
	if result.Name != expectedName {
		t.Errorf("Expected model name '%s', got '%s'", expectedName, result.Name)
	}

	if result.Model != expectedName {
		t.Errorf("Expected model '%s', got '%s'", expectedName, result.Model)
	}
}

// Test the exported types can be marshaled to JSON
func TestOllamaResponseJSON(t *testing.T) {
	response := OllamaResponse{
		Models: []OllamaModel{
			{
				Name:       "test-model",
				Model:      "test-model",
				ModifiedAt: "2025-01-01T00:00:00Z",
				Size:       1024,
				Digest:     "test-digest",
				Details: OllamaDetails{
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

	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Errorf("Expected no error marshaling JSON, got %v", err)
	}

	var parsedResponse OllamaResponse
	err = json.Unmarshal(jsonData, &parsedResponse)
	if err != nil {
		t.Errorf("Expected no error unmarshaling JSON, got %v", err)
	}

	if len(parsedResponse.Models) != 1 {
		t.Errorf("Expected 1 model in parsed response, got %d", len(parsedResponse.Models))
	}
}
