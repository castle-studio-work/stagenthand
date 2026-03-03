package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/baochen10luo/stagenthand/internal/domain"
	"github.com/baochen10luo/stagenthand/internal/image"
	"github.com/spf13/cobra"
)

var (
	outputDir string
)

func runPanelToImage(cmd *cobra.Command, args []string) error {
	inputData, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("reading stdin: %w", err)
	}

	var panel domain.Panel
	if err := json.Unmarshal(inputData, &panel); err != nil {
		return fmt.Errorf("parsing panel json: %w", err)
	}

	provider := "mock"
	if cfg != nil && cfg.Image.Provider != "" {
		provider = cfg.Image.Provider
	}

	// This assumes image client factory will exist similar to llm.NewClient
	client, err := image.NewClient(provider, dryRun, cfg)
	if err != nil {
		return fmt.Errorf("image client factory: %w", err)
	}

	imgBytes, err := client.GenerateImage(context.Background(), panel.Description, panel.CharacterRefs)
	if err != nil {
		return fmt.Errorf("image generation failed: %w", err)
	}

	// Output logic goes here -- in dry-run, we might still spit out the modified JSON block.
	// For now, let's pretend we saved it and update the URL.
	panel.ImageURL = "mocked_path.png"
	if len(imgBytes) == 0 {
		panel.ImageURL = "error.png"
	}

	outBytes, _ := json.Marshal(panel)
	fmt.Fprintln(os.Stdout, string(outBytes))
	return nil
}

var panelToImageCmd = &cobra.Command{
	Use:   "panel-to-image",
	Short: "Generate one image for a selected panel in dry-run mode",
	RunE:  runPanelToImage,
}

func init() {
	panelToImageCmd.Flags().StringVar(&outputDir, "out-dir", ".", "Directory to output simulated images")
	rootCmd.AddCommand(panelToImageCmd)
}
