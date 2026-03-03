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
	panelsOutputDir string
	workers         int
)

func runPanelsToImages(cmd *cobra.Command, args []string) error {
	inputData, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("reading stdin: %w", err)
	}

	var payload struct {
		ProjectID string         `json:"project_id"`
		Episode   int            `json:"episode"`
		Panels    []domain.Panel `json:"panels"`
	}

	if err := json.Unmarshal(inputData, &payload); err != nil {
		return fmt.Errorf("parsing panels payload: %w", err)
	}

	provider := "mock"
	if cfg != nil && cfg.Image.Provider != "" {
		provider = cfg.Image.Provider
	}

	client, err := image.NewClient(provider, dryRun, cfg)
	if err != nil {
		return fmt.Errorf("image client factory: %w", err)
	}

	outPanels, errs := image.GenerateBatch(context.Background(), client, payload.Panels, panelsOutputDir, workers)
	payload.Panels = outPanels

	if verbose {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "Worker error: %v\n", e)
		}
	}

	outBytes, _ := json.Marshal(payload)
	fmt.Fprintln(os.Stdout, string(outBytes))

	if len(errs) > 0 {
		return fmt.Errorf("completed with %d errors", len(errs))
	}
	return nil
}

var panelsToImagesCmd = &cobra.Command{
	Use:   "panels-to-images",
	Short: "Generate all images for an array of panels concurrently",
	RunE:  runPanelsToImages,
}

func init() {
	panelsToImagesCmd.Flags().StringVar(&panelsOutputDir, "out-dir", ".", "Directory to save generated images")
	panelsToImagesCmd.Flags().IntVar(&workers, "workers", 3, "Number of concurrent workers")
	rootCmd.AddCommand(panelsToImagesCmd)
}
