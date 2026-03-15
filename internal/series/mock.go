package series

import "context"

// MockRepository is an in-memory Repository for testing.
type MockRepository struct {
	Memory *SeriesMemory
	Err    error
}

func (m *MockRepository) Load(_ context.Context) (*SeriesMemory, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	if m.Memory == nil {
		return &SeriesMemory{}, nil
	}
	return m.Memory, nil
}

func (m *MockRepository) Save(_ context.Context, mem *SeriesMemory) error {
	if m.Err != nil {
		return m.Err
	}
	m.Memory = mem
	return nil
}

func (m *MockRepository) Append(ctx context.Context, ep EpisodeMemory) error {
	if m.Err != nil {
		return m.Err
	}
	loaded, err := m.Load(ctx)
	if err != nil {
		return err
	}
	loaded.Episodes = append(loaded.Episodes, ep)
	return m.Save(ctx, loaded)
}

// MockSummarizer is an in-memory Summarizer for testing.
type MockSummarizer struct {
	Result EpisodeMemory
	Global string
	Err    error

	// Optionally track calls
	SummarizeCalls    int
	CompressGlobalCalls int
}

func (m *MockSummarizer) Summarize(_ context.Context, episodeNum int, _ []byte) (EpisodeMemory, error) {
	m.SummarizeCalls++
	if m.Err != nil {
		return EpisodeMemory{}, m.Err
	}
	result := m.Result
	result.Episode = episodeNum
	return result, nil
}

func (m *MockSummarizer) CompressGlobal(_ context.Context, _ *SeriesMemory) (string, error) {
	m.CompressGlobalCalls++
	if m.Err != nil {
		return "", m.Err
	}
	return m.Global, nil
}
