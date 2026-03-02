package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/baochen10luo/stagenthand/internal/store"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status <job-id>",
	Short: "Query the status of a pipeline job",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := store.New(cfg.Store.DBPath)
		if err != nil {
			return fmt.Errorf("opening db: %w", err)
		}
		repo := store.NewGormJobRepository(db)
		job, err := repo.GetByID(args[0])
		if err == store.ErrNotFound {
			fmt.Fprintf(os.Stderr, "job %q not found\n", args[0])
			os.Exit(1)
		}
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(job)
	},
}
