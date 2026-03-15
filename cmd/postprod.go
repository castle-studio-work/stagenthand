package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/baochen10luo/stagenthand/internal/domain"
	"github.com/baochen10luo/stagenthand/internal/llm"
	"github.com/baochen10luo/stagenthand/internal/postprod"
	"github.com/baochen10luo/stagenthand/internal/remotion"
	"github.com/baochen10luo/stagenthand/internal/video"
	"github.com/spf13/cobra"
)

// ---- Adapter: video.Critic → postprod.VideoEvaluator ----

// criticEvaluatorAdapter wraps video.Critic to implement postprod.VideoEvaluator.
type criticEvaluatorAdapter struct {
	critic *video.Critic
}

func (a *criticEvaluatorAdapter) Evaluate(ctx context.Context, videoPath string, propsJSON []byte) (*postprod.EvaluationResult, error) {
	eval, err := a.critic.Evaluate(ctx, videoPath, propsJSON)
	if err != nil {
		return nil, err
	}
	return &postprod.EvaluationResult{
		VisualScore:    eval.VisualScore,
		AudioSyncScore: eval.AudioSyncScore,
		AdherenceScore: eval.AdherenceScore,
		ToneScore:      eval.ToneScore,
		Feedback:       eval.Feedback,
		Action:         eval.Action,
	}, nil
}

// ---- Adapter: remotion.Executor → postprod.VideoRenderer ----

// remotionRendererAdapter wraps remotion.Executor to implement postprod.VideoRenderer.
// It writes propsJSON to a temp file before calling Render.
type remotionRendererAdapter struct {
	executor    remotion.Executor
	templatePath string
	composition  string
}

func (a *remotionRendererAdapter) Render(ctx context.Context, propsJSON []byte, outputPath string) error {
	// Write propsJSON to a temp file
	f, err := os.CreateTemp("", "shand-postprod-props-*.json")
	if err != nil {
		return fmt.Errorf("remotionRendererAdapter: create temp file: %w", err)
	}
	defer os.Remove(f.Name())

	if _, err := f.Write(propsJSON); err != nil {
		f.Close()
		return fmt.Errorf("remotionRendererAdapter: write props: %w", err)
	}
	f.Close()

	absProps, _ := filepath.Abs(f.Name())
	absOutput, _ := filepath.Abs(outputPath)

	return a.executor.Render(ctx, a.templatePath, a.composition, absProps, absOutput)
}

// ---- Command flags ----

var (
	postprodVideoPath  string
	postprodPropsPath  string
	postprodPlanPath   string
	postprodOutputPath string
	postprodOutputDir  string
	postprodMaxIter    int
	postprodDryRun     bool
)

// ---- Top-level postprod command ----

var postprodCmd = &cobra.Command{
	Use:   "postprod",
	Short: "Agentic post-production: evaluate, plan, apply, and loop",
}

// ---- postprod evaluate ----

var postprodEvaluateCmd = &cobra.Command{
	Use:   "evaluate",
	Short: "Evaluate a rendered video against its RemotionProps using the AI Critic",
	RunE: func(cmd *cobra.Command, args []string) error {
		if postprodVideoPath == "" || postprodPropsPath == "" {
			return fmt.Errorf("--video and --props are required")
		}

		bedrockClient, err := llm.NewBedrockClient(
			cfg.LLM.AWSAccessKeyID,
			cfg.LLM.AWSSecretAccessKey,
			cfg.LLM.AWSRegion,
			"amazon.nova-pro-v1:0",
		)
		if err != nil {
			return fmt.Errorf("create bedrock client: %w", err)
		}

		propsBytes, err := os.ReadFile(postprodPropsPath)
		if err != nil {
			return fmt.Errorf("read props file: %w", err)
		}

		critic := video.NewCritic(bedrockClient)
		evaluator := &criticEvaluatorAdapter{critic: critic}

		eval, err := evaluator.Evaluate(cmd.Context(), postprodVideoPath, propsBytes)
		if err != nil {
			return fmt.Errorf("evaluation failed: %w", err)
		}

		out, _ := json.MarshalIndent(eval, "", "  ")
		fmt.Println(string(out))
		return nil
	},
}

// ---- postprod apply ----

var postprodApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply an EditPlan to RemotionProps",
	RunE: func(cmd *cobra.Command, args []string) error {
		if postprodPlanPath == "" || postprodPropsPath == "" {
			return fmt.Errorf("--plan and --props are required")
		}

		planBytes, err := os.ReadFile(postprodPlanPath)
		if err != nil {
			return fmt.Errorf("read plan file: %w", err)
		}

		propsBytes, err := os.ReadFile(postprodPropsPath)
		if err != nil {
			return fmt.Errorf("read props file: %w", err)
		}

		var plan domain.EditPlan
		if err := json.Unmarshal(planBytes, &plan); err != nil {
			return fmt.Errorf("parse plan JSON: %w", err)
		}

		if postprodDryRun {
			out, _ := json.MarshalIndent(plan, "", "  ")
			fmt.Println(string(out))
			return nil
		}

		var props domain.RemotionProps
		if err := json.Unmarshal(propsBytes, &props); err != nil {
			return fmt.Errorf("parse props JSON: %w", err)
		}

		applier := postprod.NewDefaultEditApplier()
		result, err := applier.Apply(cmd.Context(), &plan, props)
		if err != nil {
			return fmt.Errorf("apply failed: %w", err)
		}

		if postprodOutputPath != "" {
			updatedJSON, _ := json.MarshalIndent(result.UpdatedProps, "", "  ")
			if err := os.WriteFile(postprodOutputPath, updatedJSON, 0644); err != nil {
				return fmt.Errorf("write output props: %w", err)
			}
		}

		out, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(out))
		return nil
	},
}

// ---- postprod rerender ----

var postprodRerenderCmd = &cobra.Command{
	Use:   "rerender",
	Short: "Re-render a video from updated RemotionProps",
	RunE: func(cmd *cobra.Command, args []string) error {
		if postprodPropsPath == "" {
			return fmt.Errorf("--props is required")
		}

		propsBytes, err := os.ReadFile(postprodPropsPath)
		if err != nil {
			return fmt.Errorf("read props file: %w", err)
		}

		outputPath := postprodOutputPath
		if outputPath == "" {
			outputPath = "output.mp4"
		}

		templatePath := cfg.Remotion.TemplatePath
		if templatePath == "" {
			templatePath = "./remotion-template"
		}
		templatePath, _ = filepath.Abs(templatePath)

		composition := cfg.Remotion.Composition
		if composition == "" {
			composition = "ShortDrama"
		}

		executor := remotion.NewCLIExecutor(dryRun)
		renderer := &remotionRendererAdapter{
			executor:     executor,
			templatePath: templatePath,
			composition:  composition,
		}

		absOutput, _ := filepath.Abs(outputPath)
		if err := renderer.Render(cmd.Context(), propsBytes, absOutput); err != nil {
			return fmt.Errorf("rerender failed: %w", err)
		}

		result := map[string]string{"output": absOutput}
		out, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(out))
		return nil
	},
}

// ---- postprod loop ----

var postprodLoopCmd = &cobra.Command{
	Use:   "loop",
	Short: "Run the full agentic post-production loop",
	RunE: func(cmd *cobra.Command, args []string) error {
		if postprodVideoPath == "" || postprodPropsPath == "" {
			return fmt.Errorf("--video and --props are required")
		}

		propsBytes, err := os.ReadFile(postprodPropsPath)
		if err != nil {
			return fmt.Errorf("read props file: %w", err)
		}

		var props domain.RemotionProps
		if err := json.Unmarshal(propsBytes, &props); err != nil {
			return fmt.Errorf("parse props JSON: %w", err)
		}

		// Build evaluator
		bedrockClient, err := llm.NewBedrockClient(
			cfg.LLM.AWSAccessKeyID,
			cfg.LLM.AWSSecretAccessKey,
			cfg.LLM.AWSRegion,
			"amazon.nova-pro-v1:0",
		)
		if err != nil {
			return fmt.Errorf("create bedrock client: %w", err)
		}
		critic := video.NewCritic(bedrockClient)
		evaluator := &criticEvaluatorAdapter{critic: critic}

		// Build planner (uses a separate Bedrock client for text generation)
		plannerLLMClient, err := llm.NewBedrockClient(
			cfg.LLM.AWSAccessKeyID,
			cfg.LLM.AWSSecretAccessKey,
			cfg.LLM.AWSRegion,
			"amazon.nova-pro-v1:0",
		)
		if err != nil {
			return fmt.Errorf("create planner llm client: %w", err)
		}
		planner := postprod.NewLLMEditPlanner(plannerLLMClient)

		// Build applier
		applier := postprod.NewDefaultEditApplier()

		// Build renderer
		templatePath := cfg.Remotion.TemplatePath
		if templatePath == "" {
			templatePath = "./remotion-template"
		}
		templatePath, _ = filepath.Abs(templatePath)
		composition := cfg.Remotion.Composition
		if composition == "" {
			composition = "ShortDrama"
		}
		executor := remotion.NewCLIExecutor(dryRun)
		renderer := &remotionRendererAdapter{
			executor:     executor,
			templatePath: templatePath,
			composition:  composition,
		}

		outputDir := postprodOutputDir
		if outputDir == "" {
			outputDir = "."
		}
		absOutputDir, _ := filepath.Abs(outputDir)

		loopCfg := postprod.LoopConfig{
			MaxIterations: postprodMaxIter,
			OutputDir:     absOutputDir,
		}
		loop := postprod.NewPostProdLoop(evaluator, planner, applier, renderer, loopCfg)

		result, err := loop.Run(cmd.Context(), postprodVideoPath, props)
		if err != nil {
			return fmt.Errorf("loop failed: %w", err)
		}

		out, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(out))
		return nil
	},
}

func init() {
	// evaluate flags
	postprodEvaluateCmd.Flags().StringVarP(&postprodVideoPath, "video", "v", "", "Path to the rendered MP4 file")
	postprodEvaluateCmd.Flags().StringVarP(&postprodPropsPath, "props", "p", "", "Path to remotion_props.json")

	// apply flags
	postprodApplyCmd.Flags().StringVar(&postprodPlanPath, "plan", "", "Path to edit_plan.json")
	postprodApplyCmd.Flags().StringVarP(&postprodPropsPath, "props", "p", "", "Path to remotion_props.json")
	postprodApplyCmd.Flags().StringVarP(&postprodOutputPath, "output", "o", "", "Write updated props JSON to this file")
	postprodApplyCmd.Flags().BoolVar(&postprodDryRun, "dry-run", false, "Print plan without executing")

	// rerender flags
	postprodRerenderCmd.Flags().StringVarP(&postprodPropsPath, "props", "p", "", "Path to remotion_props.json")
	postprodRerenderCmd.Flags().StringVarP(&postprodOutputPath, "output", "o", "output.mp4", "Output video path")

	// loop flags
	postprodLoopCmd.Flags().StringVarP(&postprodVideoPath, "video", "v", "", "Path to the initial MP4 file")
	postprodLoopCmd.Flags().StringVarP(&postprodPropsPath, "props", "p", "", "Path to remotion_props.json")
	postprodLoopCmd.Flags().IntVar(&postprodMaxIter, "max-iterations", 3, "Maximum number of retry iterations")
	postprodLoopCmd.Flags().StringVar(&postprodOutputDir, "output-dir", ".", "Directory for versioned outputs")

	// Register subcommands
	postprodCmd.AddCommand(postprodEvaluateCmd)
	postprodCmd.AddCommand(postprodApplyCmd)
	postprodCmd.AddCommand(postprodRerenderCmd)
	postprodCmd.AddCommand(postprodLoopCmd)

	// Register postprod with root
	rootCmd.AddCommand(postprodCmd)
}
