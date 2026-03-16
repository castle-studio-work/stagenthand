package postprod_test

import (
	"context"
	"errors"
	"testing"

	"github.com/baochen10luo/stagenthand/internal/postprod"
)

func TestPropsCritic_Evaluate(t *testing.T) {
	tests := []struct {
		name        string
		llmResponse []byte
		llmErr      error
		wantOK      bool
		wantIssues  int
		wantErr     bool
	}{
		{
			name:        "no issues found",
			llmResponse: []byte(`{"issues":[],"ok":true}`),
			wantOK:      true,
			wantIssues:  0,
			wantErr:     false,
		},
		{
			name:        "issues found",
			llmResponse: []byte(`{"issues":["Panel 0 DialogueLine contains metadata prefix 'VO:'","Panel 3 DurationSec=1.5 is below minimum 2.0"],"ok":false}`),
			wantOK:      false,
			wantIssues:  2,
			wantErr:     false,
		},
		{
			name:    "llm error propagates",
			llmErr:  errors.New("api timeout"),
			wantErr: true,
		},
		{
			name:        "invalid json from llm",
			llmResponse: []byte(`not json at all`),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &postprod.MockLLMClient{
				Response: tt.llmResponse,
				Err:      tt.llmErr,
			}
			critic := postprod.NewPropsCritic(mock)

			result, err := critic.Evaluate(context.Background(), []byte(`{"panels":[]}`))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.OK != tt.wantOK {
				t.Errorf("OK: got %v, want %v", result.OK, tt.wantOK)
			}
			if len(result.Issues) != tt.wantIssues {
				t.Errorf("Issues count: got %d, want %d", len(result.Issues), tt.wantIssues)
			}
		})
	}
}
