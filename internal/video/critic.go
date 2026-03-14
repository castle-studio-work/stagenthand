package video

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/baochen10luo/stagenthand/internal/llm"
)

// Evaluation represents the multi-dimensional scoring and feedback from the AI Critic.
type Evaluation struct {
	VisualScore    int    `json:"visual_score"`    // 0-10: Visual Coherence
	AudioSyncScore int    `json:"audio_sync_score"` // 0-10: Audio-Visual Sync
	AdherenceScore int    `json:"adherence_score"`  // 0-10: Directive Adherence
	ToneScore      int    `json:"tone_score"`       // 0-10: Narrative Tone
	Feedback       string `json:"feedback"`         // Specific advice for JSON JSON configuration modifications
	Action         string `json:"action"`           // "APPROVE" or "REJECT"
}

// CheckApproval returns true if the video meets the hard-stop convergence criteria.
func (e *Evaluation) CheckApproval() bool {
	// Linus rules: objective failures are fatal.
	if e.VisualScore < 8 || e.AudioSyncScore < 8 {
		return false
	}
	// Subjective/Directive failures are tolerable if the total score is high enough.
	total := e.VisualScore + e.AudioSyncScore + e.AdherenceScore + e.ToneScore
	return total >= 32
}

// Critic evaluates a generated video MP4 against the original JSON configuration
// using an advanced multi-modal foundation model.
type Critic struct {
	client llm.VideoCriticClient
}

// NewCritic creates a new AI Critic with the given multi-modal LLM client.
func NewCritic(client llm.VideoCriticClient) *Critic {
	return &Critic{client: client}
}

// Evaluate uses the configured multi-modal model to watch the video, read the directives,
// and return a structured Evaluation.
func (c *Critic) Evaluate(ctx context.Context, videoPath string, propsJSONData []byte) (*Evaluation, error) {
	videoBytes, err := os.ReadFile(videoPath)
	if err != nil {
		return nil, fmt.Errorf("could not read video file for critique: %w", err)
	}

	systemPrompt := `You are an elite AI Film Critic and Technical Director evaluating an AI-generated video.
You will be provided with:
1. The rendered MP4 video.
2. The original 'RemotionProps' JSON that generated the video. This JSON contains 'directives' (global rendering settings) and 'panels' (which contain 'duration_sec' and per-panel 'directive' settings).

Your job is to strictly evaluate the video across 4 dimensions, scoring each out of 10:
1. 'visual_score': Are there glitched frames, unintended flickering, or physical anomalies? (10 = flawless, <8 = fatal visual bugs)
2. 'audio_sync_score': Do the narrative voiceovers match the visuals? Is the background music ducking (lowering in volume) during the voiceovers correctly? (10 = flawless sync, <8 = audio is out of sync or cut off prematurely)
3. 'adherence_score': Did the video obey the 'directives' inside the JSON? (e.g., if color_filter is 'cyberpunk', does it look cyberpunk? If motion_effect is 'pan_left', does the camera pan left?)
4. 'tone_score': Does the final mood align with the script's intention?

If 'visual_score' or 'audio_sync_score' are below 8, or the total score is below 32, you MUST provide 'action': 'REJECT'. Otherwise, 'APPROVE'.
In 'feedback', concisely explain why points were deducted and specific JSON field adjustments needed. Example: 'Panel 2 duration (3.0s) is too short for the dialogue, extend it to 5.0s.'

Respond ONLY with valid JSON matching EXACTLY this structure, with no markdown formatting around it:
{
  "visual_score": 10,
  "audio_sync_score": 10,
  "adherence_score": 10,
  "tone_score": 10,
  "action": "APPROVE",
  "feedback": "..."
}`

	respBytes, err := c.client.ReviewVideo(ctx, systemPrompt, propsJSONData, "mp4", videoBytes)
	if err != nil {
		return nil, fmt.Errorf("video analysis by LLM failed: %w", err)
	}

	var eval Evaluation
	if err := json.Unmarshal(respBytes, &eval); err != nil {
		return nil, fmt.Errorf("failed to parse critic evaluation json: %w (raw response: %s)", err, string(respBytes))
	}

	return &eval, nil
}
