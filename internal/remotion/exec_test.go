package remotion_test

import (
	"context"
	"testing"

	"github.com/baochen10luo/stagenthand/internal/remotion"
)

func TestCLIExecutor_DryRun(t *testing.T) {
	executor := remotion.NewCLIExecutor(true) // dryRun=true

	err := executor.Render(context.Background(), "/tmp/template", "ShortDrama", "/tmp/props.json", "out.mp4")
	if err != nil {
		t.Fatalf("Expected nil for dry run render, got %v", err)
	}

	err = executor.Preview(context.Background(), "/tmp/template", "ShortDrama", "/tmp/props.json")
	if err != nil {
		t.Fatalf("Expected nil for dry run preview, got %v", err)
	}
}

func TestCLIExecutor_Failure(t *testing.T) {
	executor := remotion.NewCLIExecutor(false) // dryRun=false

	// Using a nonexistent path should cause a failure quickly
	err := executor.Render(context.Background(), "/nonexistent/path/for/shands", "ShortDrama", "/tmp/props.json", "out.mp4")
	if err == nil {
		t.Errorf("Expected error when running npx in nonexistent dir, got nil")
	}

	err = executor.Preview(context.Background(), "/nonexistent/path/for/shands", "ShortDrama", "/tmp/props.json")
	if err == nil {
		t.Errorf("Expected error when running studio in nonexistent dir, got nil")
	}
}
