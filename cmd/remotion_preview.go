package cmd

import (
	"io"
	"os"
	"path/filepath"

	"github.com/baochen10luo/stagenthand/internal/remotion"
	"github.com/spf13/cobra"
)

var previewPropsPath string

var remotionPreviewCmd = &cobra.Command{
	Use:   "remotion-preview",
	Short: "Open remotion studio to preview the short drama",
	RunE: func(cmd *cobra.Command, args []string) error {
		executor := remotion.NewCLIExecutor(dryRun)

		propsFile := previewPropsPath
		if propsFile == "" {
			// Read from stdin
			propsData, err := io.ReadAll(os.Stdin)
			if err != nil {
				return err
			}
			f, err := os.CreateTemp("", "shand-preview-props-*.json")
			if err != nil {
				return err
			}
			defer os.Remove(f.Name())
			f.Write(propsData)
			f.Close()
			propsFile = f.Name()
		}

		absProps, _ := filepath.Abs(propsFile)
		templatePath, _ := filepath.Abs(cfg.Remotion.TemplatePath)
		if templatePath == "" {
			templatePath = "./remotion-template"
		}

		composition := cfg.Remotion.Composition
		if composition == "" {
			composition = "ShortDrama"
		}

		if err := executor.Preview(cmd.Context(), templatePath, composition, absProps); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	remotionPreviewCmd.Flags().StringVarP(&previewPropsPath, "props", "p", "", "Path to props json file (leave empty to use stdin)")
	rootCmd.AddCommand(remotionPreviewCmd)
}
