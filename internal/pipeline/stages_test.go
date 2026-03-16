package pipeline_test

import (
	"context"
	"errors"
	"testing"

	"github.com/baochen10luo/stagenthand/internal/domain"
	"github.com/baochen10luo/stagenthand/internal/llm"
	"github.com/baochen10luo/stagenthand/internal/pipeline"
	"github.com/stretchr/testify/assert"
)

func TestRunTransformationStage(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		client := &llm.MockClient{
			GenerateFunc: func(ctx context.Context, systemPrompt string, inputData []byte) ([]byte, error) {
				return []byte(`{"ok": true}`), nil
			},
		}

		res, err := pipeline.RunTransformationStage(context.Background(), client, pipeline.PromptStoryToOutline, []byte("input story"))
		assert.NoError(t, err)
		assert.Equal(t, []byte(`{"ok": true}`), res)
		assert.Equal(t, 1, client.CallCount)
	})

	t.Run("empty input", func(t *testing.T) {
		client := &llm.MockClient{}
		_, err := pipeline.RunTransformationStage(context.Background(), client, pipeline.PromptStoryToOutline, nil)
		assert.ErrorContains(t, err, "input data cannot be empty")
		assert.Equal(t, 0, client.CallCount)
	})

	t.Run("empty sysprompt", func(t *testing.T) {
		client := &llm.MockClient{}
		_, err := pipeline.RunTransformationStage(context.Background(), client, "", []byte("123"))
		assert.ErrorContains(t, err, "system prompt cannot be empty")
		assert.Equal(t, 0, client.CallCount)
	})

	t.Run("llm failure", func(t *testing.T) {
		myErr := errors.New("API limit")
		client := &llm.MockClient{
			GenerateFunc: func(ctx context.Context, systemPrompt string, inputData []byte) ([]byte, error) {
				return nil, myErr
			},
		}
		_, err := pipeline.RunTransformationStage(context.Background(), client, pipeline.PromptStoryToOutline, []byte("req"))
		assert.ErrorIs(t, err, myErr)
		assert.ErrorContains(t, err, "transformer failed")
	})

	t.Run("llm returns empty", func(t *testing.T) {
		client := &llm.MockClient{
			GenerateFunc: func(ctx context.Context, systemPrompt string, inputData []byte) ([]byte, error) {
				return nil, nil
			},
		}
		_, err := pipeline.RunTransformationStage(context.Background(), client, pipeline.PromptStoryToOutline, []byte("req"))
		assert.ErrorContains(t, err, "transformer returned empty output")
	})
}

func TestBuildPrompt_ContainsDirectiveSchema(t *testing.T) {
	t.Parallel()

	prompt := pipeline.BuildStoryboardToPanelsPrompt("zh-TW", domain.Storyboard{})

	assert.Contains(t, prompt, "directive", "prompt must contain directive field in schema")
	assert.Contains(t, prompt, "motion_effect", "prompt must contain motion_effect field")
	assert.Contains(t, prompt, "ken_burns_in", "prompt must contain ken_burns_in as a motion_effect value")
}

func TestBuildPrompt_LanguageInjection_StillWorks(t *testing.T) {
	t.Parallel()

	prompt := pipeline.BuildStoryboardToPanelsPrompt("en-US", domain.Storyboard{})

	assert.Contains(t, prompt, "motion_effect", "prompt must still contain motion_effect")
	assert.Contains(t, prompt, "English", "prompt must contain language injection for en-US")
}

func TestBuildStoryboardToPanelsPrompt_IncludesDialogueLines(t *testing.T) {
	t.Parallel()

	prompt := pipeline.BuildStoryboardToPanelsPrompt("zh-TW", domain.Storyboard{})

	assert.Contains(t, prompt, "dialogue_lines", "prompt schema must include dialogue_lines field")
	assert.Contains(t, prompt, "speaker", "prompt schema must include speaker field in dialogue_lines")
}

func TestBuildStoryboardToPanelsPrompt_ForbidsSpeakerPrefix(t *testing.T) {
	t.Parallel()

	prompt := pipeline.BuildStoryboardToPanelsPrompt("zh-TW", domain.Storyboard{})

	assert.Contains(t, prompt, "CRITICAL DIALOGUE FORMAT RULES", "prompt must contain dialogue format rules")
	assert.Contains(t, prompt, "Do NOT include speaker names", "prompt must explicitly forbid speaker prefixes")
}

func TestCleanDialogue(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "plain text passes through",
			input: "在一個珍惜傳統的世界裡...",
			want:  "在一個珍惜傳統的世界裡...",
		},
		{
			name:  "empty string unchanged",
			input: "",
			want:  "",
		},
		{
			name:  "VO narrator single-quoted",
			input: "VO (Narrator): '在一個珍惜傳統的世界裡...'",
			want:  "在一個珍惜傳統的世界裡...",
		},
		{
			name:  "VO narrator double-quoted",
			input: `VO (Narrator): "在一個珍惜傳統的世界裡..."`,
			want:  "在一個珍惜傳統的世界裡...",
		},
		{
			name:  "Chinese speaker colon single-quoted",
			input: "奶奶: '啊，你來了！今天，我們要做些特別的東西。'",
			want:  "啊，你來了！今天，我們要做些特別的東西。",
		},
		{
			name:  "English speaker colon",
			input: "Alex: 'Let's go!'",
			want:  "Let's go!",
		},
		{
			name:  "Panel N prefix then plain text",
			input: "Panel 1: 在一個珍惜傳統的世界裡...",
			want:  "在一個珍惜傳統的世界裡...",
		},
		{
			name:  "Panel N prefix then VO quoted",
			input: "Panel 1: \"VO (Alex): '在路上'\"",
			want:  "在路上",
		},
		{
			name:  "bare wrapping single quotes",
			input: "'啊，你來了！'",
			want:  "啊，你來了！",
		},
		{
			name:  "bare wrapping double quotes",
			input: `"Let's go!"`,
			want:  "Let's go!",
		},
		{
			name:  "whitespace trimmed",
			input: "  Hello world  ",
			want:  "Hello world",
		},
		{
			name:  "Chinese speaker colon unquoted",
			input: "奶奶: 啊，你來了！",
			want:  "啊，你來了！",
		},
		{
			name:  "VO unquoted",
			input: "VO: 旁白繼續說明故事",
			want:  "旁白繼續說明故事",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := pipeline.CleanDialogue(tc.input)
			assert.Equal(t, tc.want, got)
		})
	}
}
