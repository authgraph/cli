package testing

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSuite(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
		checks  func(t *testing.T, suite *Suite)
	}{
		{
			name: "flat format",
			content: `tests:
  - name: "alice can read"
    subject: "user:alice"
    permission: "read"
    resource: "document:readme"
    expected: "allowed"
  - name: "bob cannot delete"
    subject: "user:bob"
    permission: "delete"
    resource: "project:main"
    expected: "denied"
`,
			checks: func(t *testing.T, suite *Suite) {
				if len(suite.Tests) != 2 {
					t.Fatalf("expected 2 tests, got %d", len(suite.Tests))
				}
				if suite.Tests[0].GetSubject() != "user:alice" {
					t.Errorf("expected subject user:alice, got %s", suite.Tests[0].GetSubject())
				}
				if suite.Tests[0].GetAction() != "read" {
					t.Errorf("expected action read, got %s", suite.Tests[0].GetAction())
				}
				if suite.Tests[0].GetResource() != "document:readme" {
					t.Errorf("expected resource document:readme, got %s", suite.Tests[0].GetResource())
				}
				if suite.Tests[0].GetExpected() != "allowed" {
					t.Errorf("expected allowed, got %s", suite.Tests[0].GetExpected())
				}
				if !suite.Tests[0].IsAllowed() {
					t.Error("expected IsAllowed() to be true")
				}
				if suite.Tests[1].IsAllowed() {
					t.Error("expected IsAllowed() to be false for denied test")
				}
			},
		},
		{
			name: "nested check format",
			content: `tests:
  - name: "editors can write"
    check:
      subject: "user:alice"
      action: "write"
      resource: "document:readme"
    expect: "allowed"
`,
			checks: func(t *testing.T, suite *Suite) {
				if len(suite.Tests) != 1 {
					t.Fatalf("expected 1 test, got %d", len(suite.Tests))
				}
				tc := suite.Tests[0]
				if tc.GetSubject() != "user:alice" {
					t.Errorf("expected subject user:alice, got %s", tc.GetSubject())
				}
				if tc.GetAction() != "write" {
					t.Errorf("expected action write, got %s", tc.GetAction())
				}
				if tc.GetResource() != "document:readme" {
					t.Errorf("expected resource document:readme, got %s", tc.GetResource())
				}
				if tc.GetExpected() != "allowed" {
					t.Errorf("expected allowed, got %s", tc.GetExpected())
				}
			},
		},
		{
			name: "full suite with setup and teardown",
			content: `setup:
  schema:
    file: ./permissions.yaml
  tuples:
    - subject: "user:alice"
      relation: "editor"
      resource: "document:readme"
    - subject: "user:bob"
      relation: "viewer"
      resource: "document:readme"

tests:
  - name: "editors can write"
    check:
      subject: "user:alice"
      action: "write"
      resource: "document:readme"
    expect: "allowed"

teardown:
  cleanup_setup: true
`,
			checks: func(t *testing.T, suite *Suite) {
				if suite.Setup == nil {
					t.Fatal("expected setup to be non-nil")
				}
				if suite.Setup.Schema == nil {
					t.Fatal("expected schema to be non-nil")
				}
				if len(suite.Setup.Tuples) != 2 {
					t.Errorf("expected 2 setup tuples, got %d", len(suite.Setup.Tuples))
				}
				if suite.Setup.Tuples[0].Subject != "user:alice" {
					t.Errorf("expected first tuple subject user:alice, got %s", suite.Setup.Tuples[0].Subject)
				}
				if suite.Teardown == nil {
					t.Fatal("expected teardown to be non-nil")
				}
				if !suite.Teardown.CleanupSetup {
					t.Error("expected cleanup_setup to be true")
				}
			},
		},
		{
			name:    "empty tests",
			content: `tests: []`,
			wantErr: true,
		},
		{
			name:    "invalid yaml",
			content: `{{{invalid`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "test.yaml")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			suite, err := LoadSuite(path)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.checks != nil {
				tt.checks(t, suite)
			}
		})
	}
}

func TestLoadSuiteResolvesSchemaPath(t *testing.T) {
	dir := t.TempDir()
	content := `setup:
  schema:
    file: ./schemas/permissions.yaml
tests:
  - name: "test"
    subject: "user:a"
    permission: "read"
    resource: "doc:1"
    expected: "allowed"
`
	path := filepath.Join(dir, "test.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	suite, err := LoadSuite(path)
	if err != nil {
		t.Fatal(err)
	}

	expected := filepath.Join(dir, "schemas/permissions.yaml")
	if suite.Setup.Schema.File != expected {
		t.Errorf("expected schema path %s, got %s", expected, suite.Setup.Schema.File)
	}
}

func TestLoadSuiteFileNotFound(t *testing.T) {
	_, err := LoadSuite("/nonexistent/path/test.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
