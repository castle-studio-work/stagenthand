package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/baochen10luo/stagenthand/internal/character"
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

func init() {
	characterCmd.AddCommand(characterListCmd)
	characterCmd.AddCommand(characterShowCmd)
	rootCmd.AddCommand(characterCmd)
}
