package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/baochen10luo/stagenthand/internal/character"
	"github.com/baochen10luo/stagenthand/internal/image"
	"github.com/spf13/cobra"
)

var characterCmd = &cobra.Command{
	Use:   "character",
	Short: "Manage character reference images for visual consistency",
}

var characterListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered characters",
	RunE:  runCharacterList,
}

var characterShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show metadata for a character",
	Args:  cobra.ExactArgs(1),
	RunE:  runCharacterShow,
}

var (
	characterRegisterImage       string
	characterGenerateDescription string
)

var characterRegisterCmd = &cobra.Command{
	Use:   "register <name>",
	Short: "Register a character reference image from a file",
	Args:  cobra.ExactArgs(1),
	RunE:  runCharacterRegister,
}

var characterGenerateCmd = &cobra.Command{
	Use:   "generate <name>",
	Short: "Generate and register a character reference sheet image",
	Args:  cobra.ExactArgs(1),
	RunE:  runCharacterGenerate,
}

func characterRegistryDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".shand")
}

func runCharacterList(cmd *cobra.Command, args []string) error {
	reg := character.NewFileRegistry(characterRegistryDir())
	names, err := reg.List(cmd.Context())
	if err != nil {
		return stageError("character list", "list_error", err.Error())
	}
	return json.NewEncoder(os.Stdout).Encode(names)
}

func runCharacterShow(cmd *cobra.Command, args []string) error {
	name := args[0]
	reg := character.NewFileRegistry(characterRegistryDir())
	path, err := reg.Lookup(cmd.Context(), name)
	if err != nil {
		return stageError("character show", "lookup_error", err.Error())
	}
	if path == "" {
		return stageError("character show", "not_found", fmt.Sprintf("character %q not found", name))
	}
	meta := character.CharacterMeta{
		Name:      name,
		ImagePath: path,
	}
	return json.NewEncoder(os.Stdout).Encode(meta)
}

func runCharacterRegister(cmd *cobra.Command, args []string) error {
	name := args[0]
	if characterRegisterImage == "" {
		return stageError("character register", "missing_flag", "--image flag is required")
	}
	imgBytes, err := os.ReadFile(characterRegisterImage)
	if err != nil {
		return stageError("character register", "read_error", fmt.Sprintf("reading image file: %v", err))
	}
	reg := character.NewFileRegistry(characterRegistryDir())
	imgPath, err := reg.Register(cmd.Context(), name, imgBytes)
	if err != nil {
		return stageError("character register", "register_error", err.Error())
	}
	result := map[string]string{
		"name":       name,
		"image_path": imgPath,
	}
	return json.NewEncoder(os.Stdout).Encode(result)
}

func runCharacterGenerate(cmd *cobra.Command, args []string) error {
	name := args[0]
	if characterGenerateDescription == "" {
		return stageError("character generate", "missing_flag", "--description flag is required")
	}

	// Determine image provider from config (same pattern as pipeline.go)
	imgProvider := "nanobanana"
	if cfg != nil && cfg.Image.Provider != "" {
		imgProvider = cfg.Image.Provider
	}
	imgClient, err := image.NewClient(imgProvider, dryRun, cfg)
	if err != nil {
		return stageError("character generate", "image_init_error", err.Error())
	}

	prompt := "Character reference sheet, full body portrait, plain white background, consistent lighting: " + characterGenerateDescription
	imgBytes, err := imgClient.GenerateImage(cmd.Context(), prompt, nil)
	if err != nil {
		return stageError("character generate", "generate_error", err.Error())
	}

	reg := character.NewFileRegistry(characterRegistryDir())
	imgPath, err := reg.Register(cmd.Context(), name, imgBytes)
	if err != nil {
		return stageError("character generate", "register_error", err.Error())
	}

	result := map[string]string{
		"name":       name,
		"image_path": imgPath,
	}
	return json.NewEncoder(os.Stdout).Encode(result)
}

func init() {
	characterRegisterCmd.Flags().StringVar(&characterRegisterImage, "image", "", "path to the character reference image file (required)")
	characterGenerateCmd.Flags().StringVar(&characterGenerateDescription, "description", "", "text description of the character for AI image generation (required)")

	characterCmd.AddCommand(characterListCmd)
	characterCmd.AddCommand(characterShowCmd)
	characterCmd.AddCommand(characterRegisterCmd)
	characterCmd.AddCommand(characterGenerateCmd)
	rootCmd.AddCommand(characterCmd)
}
