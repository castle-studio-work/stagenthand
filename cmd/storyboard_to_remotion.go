package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/baochen10luo/stagenthand/internal/domain"
	"github.com/spf13/cobra"
)

var storyboardToRemotionCmd = &cobra.Command{
	Use:   "storyboard-to-remotion-props",
	Short: "Convert a storyboard with image URLs into remotion props JSON",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Read storyboard or panels array from stdin
		var input interface{} // can be domain.Storyboard or []domain.Panel depending on pipeline
		if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
			return fmt.Errorf("failed to decode input: %w", err)
		}

		// Convert to RemotionProps
		var panels []domain.Panel

		// Very basic type discovery since the pipeline stages send different payload roots occasionally
		if m, ok := input.(map[string]interface{}); ok {
			// Storyboard struct layout
			if scenesVal, has := m["scenes"]; has {
				scenesBytes, _ := json.Marshal(scenesVal)
				var scenes []domain.Scene
				json.Unmarshal(scenesBytes, &scenes)
				for _, s := range scenes {
					panels = append(panels, s.Panels...)
				}
			} else {
				// Single Scene or Project? Just fallback
			}
		} else if arr, ok := input.([]interface{}); ok {
			// Array of Panels layout
			arrBytes, _ := json.Marshal(arr)
			json.Unmarshal(arrBytes, &panels)
		} else {
			return fmt.Errorf("unrecognized pipeline output structure")
		}

		props := domain.RemotionProps{
			ProjectID: "default",
			Title:     "Generated Drama",
			Panels:    panels,
			FPS:       24,
			Width:     cfg.Image.Width,
			Height:    cfg.Image.Height,
		}

		if dryRun {
			// Dry-run mode sends valid output but no side effects
		}

		if err := json.NewEncoder(os.Stdout).Encode(props); err != nil {
			return fmt.Errorf("failed to encode props: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(storyboardToRemotionCmd)
}
