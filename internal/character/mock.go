package character

import (
	"context"
	"fmt"
)

// MockRegistry is a test double for Registry.
// It stores registrations in memory.
type MockRegistry struct {
	registered  map[string][]byte
	lookupPaths map[string]string
}

// NewMockRegistry creates an empty MockRegistry.
func NewMockRegistry() *MockRegistry {
	return &MockRegistry{
		registered:  make(map[string][]byte),
		lookupPaths: make(map[string]string),
	}
}

// Register stores the image bytes in memory and returns a synthetic path.
func (m *MockRegistry) Register(_ context.Context, name string, imageBytes []byte) (string, error) {
	m.registered[name] = imageBytes
	path := fmt.Sprintf("/mock/characters/%s/ref.png", name)
	m.lookupPaths[name] = path
	return path, nil
}

// Lookup returns the path set by Register, or "" if not found.
func (m *MockRegistry) Lookup(_ context.Context, name string) (string, error) {
	return m.lookupPaths[name], nil
}

// List returns all registered character names.
func (m *MockRegistry) List(_ context.Context) ([]string, error) {
	names := make([]string, 0, len(m.registered))
	for name := range m.registered {
		names = append(names, name)
	}
	return names, nil
}
