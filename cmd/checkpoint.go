package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/baochen10luo/stagenthand/internal/domain"
	"github.com/baochen10luo/stagenthand/internal/store"
	"github.com/spf13/cobra"
)

var checkpointCmd = &cobra.Command{
	Use:   "checkpoint",
	Short: "Manage HITL checkpoints",
}

var checkpointListCmd = &cobra.Command{
	Use:   "list <job-id>",
	Short: "List all checkpoints for a job",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := store.New(cfg.Store.DBPath)
		if err != nil {
			return fmt.Errorf("opening db: %w", err)
		}
		repo := store.NewGormCheckpointRepository(db)
		cps, err := repo.ListByJobID(args[0])
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(cps)
	},
}

var checkpointApproveCmd = &cobra.Command{
	Use:   "approve <checkpoint-id>",
	Short: "Approve a HITL checkpoint",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		notes, _ := cmd.Flags().GetString("notes")
		db, err := store.New(cfg.Store.DBPath)
		if err != nil {
			return fmt.Errorf("opening db: %w", err)
		}
		repo := store.NewGormCheckpointRepository(db)
		if err := repo.UpdateStatus(args[0], domain.CheckpointStatusApproved, notes); err != nil {
			if err == store.ErrNotFound {
				fmt.Fprintf(os.Stderr, "checkpoint %q not found\n", args[0])
				os.Exit(1)
			}
			return err
		}
		fmt.Fprintf(os.Stdout, `{"checkpoint_id":%q,"status":"approved"}`+"\n", args[0])
		return nil
	},
}

var checkpointRejectCmd = &cobra.Command{
	Use:   "reject <checkpoint-id>",
	Short: "Reject a HITL checkpoint",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		notes, _ := cmd.Flags().GetString("notes")
		db, err := store.New(cfg.Store.DBPath)
		if err != nil {
			return fmt.Errorf("opening db: %w", err)
		}
		repo := store.NewGormCheckpointRepository(db)
		if err := repo.UpdateStatus(args[0], domain.CheckpointStatusRejected, notes); err != nil {
			if err == store.ErrNotFound {
				fmt.Fprintf(os.Stderr, "checkpoint %q not found\n", args[0])
				os.Exit(1)
			}
			return err
		}
		fmt.Fprintf(os.Stdout, `{"checkpoint_id":%q,"status":"rejected"}`+"\n", args[0])
		return nil
	},
}

func init() {
	checkpointApproveCmd.Flags().String("notes", "", "approval notes")
	checkpointRejectCmd.Flags().String("notes", "", "rejection reason")
	checkpointCmd.AddCommand(checkpointListCmd, checkpointApproveCmd, checkpointRejectCmd)
}
