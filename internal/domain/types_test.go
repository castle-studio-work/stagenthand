package domain_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/baochen10luo/stagenthand/internal/domain"
)

func TestJobStatus_String(t *testing.T) {
	tests := []struct {
		status domain.JobStatus
		want   string
	}{
		{domain.JobStatusPending, "pending"},
		{domain.JobStatusRunning, "running"},
		{domain.JobStatusWaitingHITL, "waiting_hitl"},
		{domain.JobStatusCompleted, "completed"},
		{domain.JobStatusFailed, "failed"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("JobStatus.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCheckpointStage_String(t *testing.T) {
	tests := []struct {
		stage domain.CheckpointStage
		want  string
	}{
		{domain.StageOutline, "outline"},
		{domain.StageStoryboard, "storyboard"},
		{domain.StageImages, "images"},
		{domain.StageFinal, "final"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.stage.String(); got != tt.want {
				t.Errorf("CheckpointStage.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPanel_HasCharacterRefs(t *testing.T) {
	p1 := domain.Panel{CharacterRefs: []string{"char_a.png"}}
	p2 := domain.Panel{CharacterRefs: []string{}}

	if !p1.HasCharacterRefs() {
		t.Error("expected p1.HasCharacterRefs() == true")
	}
	if p2.HasCharacterRefs() {
		t.Error("expected p2.HasCharacterRefs() == false")
	}
}

func TestJob_IsTerminal(t *testing.T) {
	cases := []struct {
		status   domain.JobStatus
		terminal bool
	}{
		{domain.JobStatusCompleted, true},
		{domain.JobStatusFailed, true},
		{domain.JobStatusPending, false},
		{domain.JobStatusRunning, false},
		{domain.JobStatusWaitingHITL, false},
	}
	for _, c := range cases {
		j := domain.Job{Status: c.status, CreatedAt: time.Now()}
		if got := j.IsTerminal(); got != c.terminal {
			t.Errorf("Job{%s}.IsTerminal() = %v, want %v", c.status, got, c.terminal)
		}
	}
}

func TestEditPlan_JSONRoundTrip(t *testing.T) {
	generatedAt := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	plan := domain.EditPlan{
		Version:     "v1",
		GeneratedAt: generatedAt,
		Operations: []domain.EditOperation{
			{
				Type: domain.EditOpPatchDialogue,
				TargetPanel: &domain.PanelRef{
					SceneNumber: 1,
					PanelNumber: 2,
				},
				Params:    map[string]interface{}{"dialogue": "Hello world"},
				Priority:  1,
				Rationale: "fix subtitle",
			},
			{
				Type:      domain.EditOpPatchGlobalDirective,
				Priority:  2,
				Rationale: "fix visual",
			},
		},
		EstimatedCost: 0.05,
		Rationale:     "improve quality",
	}

	data, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var got domain.EditPlan
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if got.Version != plan.Version {
		t.Errorf("Version: got %q, want %q", got.Version, plan.Version)
	}
	if len(got.Operations) != 2 {
		t.Fatalf("Operations len: got %d, want 2", len(got.Operations))
	}
	if got.Operations[0].Type != domain.EditOpPatchDialogue {
		t.Errorf("Op[0] Type: got %q, want %q", got.Operations[0].Type, domain.EditOpPatchDialogue)
	}
	if got.Operations[0].TargetPanel == nil {
		t.Fatal("Op[0] TargetPanel should not be nil")
	}
	if got.Operations[0].TargetPanel.SceneNumber != 1 {
		t.Errorf("SceneNumber: got %d, want 1", got.Operations[0].TargetPanel.SceneNumber)
	}
	if got.Operations[1].TargetPanel != nil {
		t.Error("Op[1] TargetPanel should be nil")
	}
	if got.EstimatedCost != plan.EstimatedCost {
		t.Errorf("EstimatedCost: got %f, want %f", got.EstimatedCost, plan.EstimatedCost)
	}
}

func TestEditResult_JSONRoundTrip(t *testing.T) {
	result := domain.EditResult{
		PlanVersion:       "v1",
		OperationsApplied: 3,
		OperationsFailed:  1,
		UpdatedProps: domain.RemotionProps{
			ProjectID: "proj-1",
			Title:     "test",
			FPS:       24,
		},
		Success: false,
		Errors:  []string{"panel not found: scene=5 panel=1"},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var got domain.EditResult
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if got.PlanVersion != result.PlanVersion {
		t.Errorf("PlanVersion: got %q, want %q", got.PlanVersion, result.PlanVersion)
	}
	if got.OperationsApplied != 3 {
		t.Errorf("OperationsApplied: got %d, want 3", got.OperationsApplied)
	}
	if got.OperationsFailed != 1 {
		t.Errorf("OperationsFailed: got %d, want 1", got.OperationsFailed)
	}
	if got.UpdatedProps.ProjectID != "proj-1" {
		t.Errorf("UpdatedProps.ProjectID: got %q, want proj-1", got.UpdatedProps.ProjectID)
	}
	if got.Success {
		t.Error("Success should be false")
	}
	if len(got.Errors) != 1 {
		t.Fatalf("Errors len: got %d, want 1", len(got.Errors))
	}
}

func TestEditOperationType_Constants(t *testing.T) {
	ops := []domain.EditOperationType{
		domain.EditOpRegenerateImage,
		domain.EditOpRegenerateAudio,
		domain.EditOpReplaceBGM,
		domain.EditOpPatchDialogue,
		domain.EditOpPatchDuration,
		domain.EditOpPatchPanelDirective,
		domain.EditOpPatchGlobalDirective,
		domain.EditOpRerender,
	}
	if len(ops) != 8 {
		t.Errorf("expected 8 EditOperationType constants, got %d", len(ops))
	}
}

func TestDialogueLine_JSONRoundTrip(t *testing.T) {
	line := domain.DialogueLine{
		Speaker: "Alice",
		Text:    "Hello world",
		Emotion: "happy",
	}

	data, err := json.Marshal(line)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var got domain.DialogueLine
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if got.Speaker != line.Speaker {
		t.Errorf("Speaker: got %q, want %q", got.Speaker, line.Speaker)
	}
	if got.Text != line.Text {
		t.Errorf("Text: got %q, want %q", got.Text, line.Text)
	}
	if got.Emotion != line.Emotion {
		t.Errorf("Emotion: got %q, want %q", got.Emotion, line.Emotion)
	}
}

func TestDialogueLine_EmotionOmittedWhenEmpty(t *testing.T) {
	line := domain.DialogueLine{
		Speaker: "",
		Text:    "Narrator voice",
	}

	data, err := json.Marshal(line)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	raw := string(data)
	if contains := func(s, sub string) bool {
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	}; contains(raw, `"emotion"`) {
		t.Errorf("emotion should be omitted when empty, got: %s", raw)
	}
}

func TestPanel_DialogueLinesField(t *testing.T) {
	panel := domain.Panel{
		SceneNumber: 1,
		PanelNumber: 1,
		Description: "A scene",
		Dialogue:    "Hello",
		DurationSec: 3.0,
		DialogueLines: []domain.DialogueLine{
			{Speaker: "Alice", Text: "Hello", Emotion: "neutral"},
			{Speaker: "Bob", Text: "World", Emotion: "happy"},
		},
	}

	data, err := json.Marshal(panel)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var got domain.Panel
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(got.DialogueLines) != 2 {
		t.Fatalf("DialogueLines len: got %d, want 2", len(got.DialogueLines))
	}
	if got.DialogueLines[0].Speaker != "Alice" {
		t.Errorf("DialogueLines[0].Speaker: got %q, want Alice", got.DialogueLines[0].Speaker)
	}
	if got.DialogueLines[1].Emotion != "happy" {
		t.Errorf("DialogueLines[1].Emotion: got %q, want happy", got.DialogueLines[1].Emotion)
	}
	// Ensure backward-compat Dialogue string still survives round-trip
	if got.Dialogue != "Hello" {
		t.Errorf("Dialogue string: got %q, want Hello", got.Dialogue)
	}
}

func TestPanel_BackwardCompat(t *testing.T) {
	// Old-format JSON with only "dialogue" (no dialogue_lines)
	raw := `{"scene_number":1,"panel_number":1,"description":"test","dialogue":"Old dialogue","character_refs":[],"duration_sec":4.0}`

	var panel domain.Panel
	if err := json.Unmarshal([]byte(raw), &panel); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if panel.Dialogue != "Old dialogue" {
		t.Errorf("Dialogue: got %q, want 'Old dialogue'", panel.Dialogue)
	}
	if panel.DialogueLines != nil {
		t.Errorf("DialogueLines should be nil for old-format JSON, got %v", panel.DialogueLines)
	}
}
