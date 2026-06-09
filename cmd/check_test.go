package cmd

import "testing"

func TestParseEntity(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *entity
		wantErr bool
	}{
		{
			name:  "valid user",
			input: "user:alice",
			want:  &entity{Type: "user", ID: "alice"},
		},
		{
			name:  "valid document",
			input: "document:readme",
			want:  &entity{Type: "document", ID: "readme"},
		},
		{
			name:  "id with colon",
			input: "resource:org/project:main",
			want:  &entity{Type: "resource", ID: "org/project:main"},
		},
		{
			name:    "missing colon",
			input:   "useralice",
			wantErr: true,
		},
		{
			name:    "empty type",
			input:   ":alice",
			wantErr: true,
		},
		{
			name:    "empty id",
			input:   "user:",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseEntity(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for input %q, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Type != tt.want.Type || got.ID != tt.want.ID {
				t.Errorf("parseEntity(%q) = {%s, %s}, want {%s, %s}",
					tt.input, got.Type, got.ID, tt.want.Type, tt.want.ID)
			}
		})
	}
}
