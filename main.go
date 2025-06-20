package main

import (
	"encoding/json"
	"fmt"
	"os"

	"dmr-models-convert/pkg/converter"

	"github.com/spf13/cobra"
)

var (
	// Used for flags
	output string
	dmrURL string
)

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

		// Create converter instance
		conv := converter.NewConverter()

		// Fetch and convert models
		ollamaResponse, err := conv.ConvertFromURL(dmrURL)
		if err != nil {
			fmt.Printf("Error converting DMR models: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Found %d models in DMR response\n", len(ollamaResponse.Models))

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

// saveOllamaResponse saves the Ollama response to a JSON file
func saveOllamaResponse(response converter.OllamaResponse, filename string) error {
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
func printOllamaResponse(response converter.OllamaResponse) error {
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
