package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	// Use a temp directory for config
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg := &Config{
		APIKey:  "ag_test_key_123",
		BaseURL: "https://api.example.com",
		OrgID:   "org_123",
	}

	// Ensure directory exists
	dir := filepath.Join(tmpDir, ".config", "authgraph")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.APIKey != cfg.APIKey {
		t.Errorf("APIKey = %q, want %q", loaded.APIKey, cfg.APIKey)
	}
	if loaded.BaseURL != cfg.BaseURL {
		t.Errorf("BaseURL = %q, want %q", loaded.BaseURL, cfg.BaseURL)
	}
	if loaded.OrgID != cfg.OrgID {
		t.Errorf("OrgID = %q, want %q", loaded.OrgID, cfg.OrgID)
	}
}

func TestLoadNotLoggedIn(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when not logged in")
	}
}

func TestSaveDefaultBaseURL(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	dir := filepath.Join(tmpDir, ".config", "authgraph")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}

	cfg := &Config{APIKey: "ag_test_key_123"}
	if err := Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.BaseURL != "https://api.authgraph.dev" {
		t.Errorf("BaseURL = %q, want default", loaded.BaseURL)
	}
}

func TestDelete(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	dir := filepath.Join(tmpDir, ".config", "authgraph")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}

	cfg := &Config{APIKey: "ag_test_key_123"}
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}

	if err := Delete(); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := Load()
	if err == nil {
		t.Fatal("expected error after delete")
	}
}
