package character_test

import (
	"context"
	"testing"

	"github.com/baochen10luo/stagenthand/internal/character"
)

func TestFileRegistry_RegisterAndLookup(t *testing.T) {
	rootDir := t.TempDir()
	reg := character.NewFileRegistry(rootDir)
	ctx := context.Background()

	imgBytes := []byte("fakepng")
	path, err := reg.Register(ctx, "Alice", imgBytes)
	if err != nil {
		t.Fatalf("Register error: %v", err)
	}
	if path == "" {
		t.Fatal("Register returned empty path")
	}

	looked, err := reg.Lookup(ctx, "Alice")
	if err != nil {
		t.Fatalf("Lookup error: %v", err)
	}
	if looked == "" {
		t.Fatal("Lookup returned empty path for registered character")
	}
}

func TestFileRegistry_LookupNotFound(t *testing.T) {
	rootDir := t.TempDir()
	reg := character.NewFileRegistry(rootDir)
	ctx := context.Background()

	path, err := reg.Lookup(ctx, "Ghost")
	if err != nil {
		t.Fatalf("Lookup unknown character should not error: %v", err)
	}
	if path != "" {
		t.Errorf("Lookup unknown character: expected empty string, got %q", path)
	}
}

func TestFileRegistry_List(t *testing.T) {
	rootDir := t.TempDir()
	reg := character.NewFileRegistry(rootDir)
	ctx := context.Background()

	names := []string{"Alice", "Bob", "Charlie"}
	for _, name := range names {
		if _, err := reg.Register(ctx, name, []byte("img")); err != nil {
			t.Fatalf("Register(%q) error: %v", name, err)
		}
	}

	list, err := reg.List(ctx)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(list) != len(names) {
		t.Errorf("List returned %d items, want %d", len(list), len(names))
	}
}

func TestFileRegistry_ListEmpty(t *testing.T) {
	rootDir := t.TempDir()
	reg := character.NewFileRegistry(rootDir)
	ctx := context.Background()

	list, err := reg.List(ctx)
	if err != nil {
		t.Fatalf("List on empty dir error: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected empty list, got %v", list)
	}
}

func TestMockRegistry(t *testing.T) {
	reg := character.NewMockRegistry()
	ctx := context.Background()

	// Register
	path, err := reg.Register(ctx, "Hero", []byte("heroimg"))
	if err != nil {
		t.Fatalf("Register error: %v", err)
	}
	if path == "" {
		t.Fatal("Register returned empty path")
	}

	// Lookup
	looked, err := reg.Lookup(ctx, "Hero")
	if err != nil {
		t.Fatalf("Lookup error: %v", err)
	}
	if looked == "" {
		t.Fatal("Lookup returned empty path for registered character")
	}

	// Lookup unknown
	unknown, err := reg.Lookup(ctx, "Ghost")
	if err != nil {
		t.Fatalf("Lookup unknown error: %v", err)
	}
	if unknown != "" {
		t.Errorf("Lookup unknown: expected empty, got %q", unknown)
	}

	// List
	names, err := reg.List(ctx)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(names) != 1 || names[0] != "Hero" {
		t.Errorf("List = %v, want [Hero]", names)
	}
}

func TestFileRegistry_OverwriteExisting(t *testing.T) {
	rootDir := t.TempDir()
	reg := character.NewFileRegistry(rootDir)
	ctx := context.Background()

	if _, err := reg.Register(ctx, "Alice", []byte("original")); err != nil {
		t.Fatalf("First Register error: %v", err)
	}
	if _, err := reg.Register(ctx, "Alice", []byte("updated")); err != nil {
		t.Fatalf("Second Register (overwrite) error: %v", err)
	}

	path, err := reg.Lookup(ctx, "Alice")
	if err != nil {
		t.Fatalf("Lookup after overwrite error: %v", err)
	}
	if path == "" {
		t.Fatal("path should not be empty after overwrite")
	}
}

func TestFileRegistry_GetMeta_Found(t *testing.T) {
	rootDir := t.TempDir()
	reg := character.NewFileRegistry(rootDir)
	ctx := context.Background()

	_, err := reg.RegisterWithMeta(ctx, "Alice", []byte("img"), character.CharacterMeta{
		VoiceID: "Joanna",
		EmotionPresets: map[string]string{
			"angry": "fast",
		},
	})
	if err != nil {
		t.Fatalf("RegisterWithMeta error: %v", err)
	}

	meta, err := reg.GetMeta(ctx, "Alice")
	if err != nil {
		t.Fatalf("GetMeta error: %v", err)
	}
	if meta == nil {
		t.Fatal("GetMeta returned nil for registered character")
	}
	if meta.Name != "Alice" {
		t.Errorf("meta.Name = %q, want Alice", meta.Name)
	}
	if meta.VoiceID != "Joanna" {
		t.Errorf("meta.VoiceID = %q, want Joanna", meta.VoiceID)
	}
	if meta.EmotionPresets["angry"] != "fast" {
		t.Errorf("meta.EmotionPresets[angry] = %q, want fast", meta.EmotionPresets["angry"])
	}
}

func TestFileRegistry_GetMeta_NotFound(t *testing.T) {
	rootDir := t.TempDir()
	reg := character.NewFileRegistry(rootDir)
	ctx := context.Background()

	meta, err := reg.GetMeta(ctx, "Ghost")
	if err != nil {
		t.Fatalf("GetMeta for unknown character should not error: %v", err)
	}
	if meta != nil {
		t.Errorf("GetMeta for unknown character should return nil, got %+v", meta)
	}
}

func TestFileRegistry_Register_WithVoiceID(t *testing.T) {
	rootDir := t.TempDir()
	reg := character.NewFileRegistry(rootDir)
	ctx := context.Background()

	_, err := reg.RegisterWithMeta(ctx, "Bob", []byte("img"), character.CharacterMeta{
		VoiceID: "Matthew",
	})
	if err != nil {
		t.Fatalf("RegisterWithMeta error: %v", err)
	}

	meta, err := reg.GetMeta(ctx, "Bob")
	if err != nil {
		t.Fatalf("GetMeta error: %v", err)
	}
	if meta == nil {
		t.Fatal("expected meta, got nil")
	}
	if meta.VoiceID != "Matthew" {
		t.Errorf("VoiceID persisted = %q, want Matthew", meta.VoiceID)
	}
}
