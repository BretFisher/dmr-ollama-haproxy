package converter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// DMR API response structures
type DMRModel struct {
	ID      string    `json:"id"`
	Tags    []string  `json:"tags"`
	Created int64     `json:"created"`
	Config  DMRConfig `json:"config"`
}

type DMRConfig struct {
	Format       string `json:"format"`
	Quantization string `json:"quantization"`
	Parameters   string `json:"parameters"`
	Architecture string `json:"architecture"`
	Size         string `json:"size"`
}

// Ollama API response structures
type OllamaResponse struct {
	Models []OllamaModel `json:"models"`
}

type OllamaModel struct {
	Name       string        `json:"name"`
	Model      string        `json:"model"`
	ModifiedAt string        `json:"modified_at"`
	Size       int64         `json:"size"`
	Digest     string        `json:"digest"`
	Details    OllamaDetails `json:"details"`
}

type OllamaDetails struct {
	ParentModel       string   `json:"parent_model"`
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
}

// Converter provides methods to convert DMR models to Ollama format
type Converter struct {
	client *http.Client
}

// NewConverter creates a new Converter instance
func NewConverter() *Converter {
	return &Converter{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewConverterWithClient creates a new Converter with a custom HTTP client
func NewConverterWithClient(client *http.Client) *Converter {
	return &Converter{
		client: client,
	}
}

// FetchDMRModels fetches models from the DMR API
func (c *Converter) FetchDMRModels(url string) ([]DMRModel, error) {
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from DMR API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("DMR API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var dmrModels []DMRModel
	err = json.Unmarshal(body, &dmrModels)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DMR JSON: %w", err)
	}

	return dmrModels, nil
}

// ConvertDMRToOllama converts DMR models to Ollama format
func (c *Converter) ConvertDMRToOllama(dmrModels []DMRModel) OllamaResponse {
	var ollamaModels []OllamaModel

	for _, dmrModel := range dmrModels {
		ollamaModel := c.convertSingleModel(dmrModel)
		ollamaModels = append(ollamaModels, ollamaModel)
	}

	return OllamaResponse{Models: ollamaModels}
}

// ConvertFromURL fetches DMR models from a URL and converts them to Ollama format
func (c *Converter) ConvertFromURL(url string) (OllamaResponse, error) {
	dmrModels, err := c.FetchDMRModels(url)
	if err != nil {
		return OllamaResponse{}, err
	}

	return c.ConvertDMRToOllama(dmrModels), nil
}

// ConvertFromJSON converts DMR models from JSON string to Ollama format
func (c *Converter) ConvertFromJSON(jsonData []byte) (OllamaResponse, error) {
	var dmrModels []DMRModel
	err := json.Unmarshal(jsonData, &dmrModels)
	if err != nil {
		return OllamaResponse{}, fmt.Errorf("failed to parse DMR JSON: %w", err)
	}

	return c.ConvertDMRToOllama(dmrModels), nil
}

// convertSingleModel converts a single DMR model to Ollama format
func (c *Converter) convertSingleModel(dmrModel DMRModel) OllamaModel {
	// Convert timestamp from Unix timestamp to RFC3339 format
	modifiedAt := time.Unix(dmrModel.Created, 0).Format(time.RFC3339)

	// Convert size string to bytes (approximate)
	sizeBytes := parseSizeString(dmrModel.Config.Size)

	// Extract digest from ID (remove "sha256:" prefix)
	digest := strings.TrimPrefix(dmrModel.ID, "sha256:")

	// Determine family from architecture
	family := determineFamily(dmrModel.Config.Architecture)

	// Get model name from first tag, or use digest as fallback
	modelName := digest
	if len(dmrModel.Tags) > 0 {
		modelName = dmrModel.Tags[0]
	}

	return OllamaModel{
		Name:       modelName,
		Model:      modelName,
		ModifiedAt: modifiedAt,
		Size:       sizeBytes,
		Digest:     digest,
		Details: OllamaDetails{
			ParentModel:       "",
			Format:            dmrModel.Config.Format,
			Family:            family,
			Families:          []string{family},
			ParameterSize:     dmrModel.Config.Parameters,
			QuantizationLevel: dmrModel.Config.Quantization,
		},
	}
}

// parseSizeString converts size strings like "690.24 MiB" to bytes
func parseSizeString(sizeStr string) int64 {
	// Remove spaces and convert to lowercase
	sizeStr = strings.ToLower(strings.ReplaceAll(sizeStr, " ", ""))

	// Handle different size units
	var multiplier int64 = 1
	if strings.HasSuffix(sizeStr, "gib") {
		multiplier = 1024 * 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "gib")
	} else if strings.HasSuffix(sizeStr, "mib") {
		multiplier = 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "mib")
	} else if strings.HasSuffix(sizeStr, "kib") {
		multiplier = 1024
		sizeStr = strings.TrimSuffix(sizeStr, "kib")
	} else if strings.HasSuffix(sizeStr, "b") {
		multiplier = 1
		sizeStr = strings.TrimSuffix(sizeStr, "b")
	}

	// Parse the numeric value
	size, err := strconv.ParseFloat(sizeStr, 64)
	if err != nil {
		return 0
	}

	return int64(size * float64(multiplier))
}

// determineFamily maps architecture to family
func determineFamily(architecture string) string {
	switch strings.ToLower(architecture) {
	case "llama", "llama2", "llama3":
		return "llama"
	case "phi3", "phi4":
		return "phi3"
	case "qwen", "qwen3":
		return "qwen"
	default:
		return architecture
	}
}
