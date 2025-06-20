package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	// Used for flags
	output string
	dmrURL string
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

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dmr-models-convert",
	Short: "Convert DMR model output to Ollama format",
	Long: `A CLI tool that converts Docker Model Runner (DMR) API responses 
to Ollama API format. This allows tools configured for Ollama to work 
with DMR servers.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Execute the convert command by default
		convertCmd.Run(cmd, args)
	},
}

// convertCmd represents the convert command
var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert DMR models to Ollama format",
	Long: `Convert the models from DMR API format to Ollama API format 
and save the result to the specified output file or print to stdout.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Fetching models from DMR server: %s\n", dmrURL)

		// Fetch models from DMR API
		dmrModels, err := fetchDMRModels(dmrURL)
		if err != nil {
			fmt.Printf("Error fetching DMR models: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Found %d models in DMR response\n", len(dmrModels))

		// Convert DMR models to Ollama format
		ollamaResponse := convertDMRToOllama(dmrModels)

		// Save converted JSON to output file or print to stdout
		if output != "" {
			err = saveOllamaResponse(ollamaResponse, output)
			if err != nil {
				fmt.Printf("Error saving output file: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Successfully converted and saved to: %s\n", output)
		} else {
			err = printOllamaResponse(ollamaResponse)
			if err != nil {
				fmt.Printf("Error printing JSON: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Root command flags (available for all commands)
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "Output file path for converted JSON (optional, prints to stdout if not specified)")
	rootCmd.PersistentFlags().StringVarP(&dmrURL, "dmr", "d", "http://localhost:12434/models", "DMR server URL (optional, defaults to http://localhost:12434/models)")

	// Add the convert command to root
	rootCmd.AddCommand(convertCmd)
}

// fetchDMRModels fetches models from the DMR API
func fetchDMRModels(url string) ([]DMRModel, error) {
	resp, err := http.Get(url)
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

// convertDMRToOllama converts DMR models to Ollama format
func convertDMRToOllama(dmrModels []DMRModel) OllamaResponse {
	var ollamaModels []OllamaModel

	for _, dmrModel := range dmrModels {
		ollamaModel := convertSingleModel(dmrModel)
		ollamaModels = append(ollamaModels, ollamaModel)
	}

	return OllamaResponse{Models: ollamaModels}
}

// convertSingleModel converts a single DMR model to Ollama format
func convertSingleModel(dmrModel DMRModel) OllamaModel {
	// Convert timestamp from Unix timestamp to RFC3339 format
	modifiedAt := time.Unix(dmrModel.Created, 0).Format(time.RFC3339)

	// Convert size string to bytes (approximate)
	sizeBytes := parseSizeString(dmrModel.Config.Size)

	// Extract digest from ID (remove "sha256:" prefix)
	digest := strings.TrimPrefix(dmrModel.ID, "sha256:")

	// Determine family from architecture
	family := determineFamily(dmrModel.Config.Architecture)

	return OllamaModel{
		Name:       dmrModel.Tags[0],
		Model:      dmrModel.Tags[0],
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

// saveOllamaResponse saves the Ollama response to a JSON file
func saveOllamaResponse(response OllamaResponse, filename string) error {
	// Create pretty-printed JSON
	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to file
	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// printOllamaResponse prints the Ollama response to stdout
func printOllamaResponse(response OllamaResponse) error {
	// Create pretty-printed JSON
	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Print to stdout
	fmt.Println(string(jsonData))
	return nil
}

func main() {
	Execute()
}
