package util

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func TestReadImageAsBase64(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		setup     func() string
		want      string
		wantError bool
	}{
		{
			name: "valid file returns base64",
			setup: func() string {
				p := filepath.Join(tmpDir, "img.bin")
				data := []byte{0x01, 0x02, 0x03, 0xFF}
				if err := os.WriteFile(p, data, 0644); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
				return p
			},
			want:      base64.StdEncoding.EncodeToString([]byte{0x01, 0x02, 0x03, 0xFF}),
			wantError: false,
		},
		{
			name: "empty file returns empty base64",
			setup: func() string {
				p := filepath.Join(tmpDir, "empty.bin")
				if err := os.WriteFile(p, []byte{}, 0644); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
				return p
			},
			want:      "",
			wantError: false,
		},
		{
			name: "file does not exist",
			setup: func() string {
				return filepath.Join(tmpDir, "missing.bin")
			},
			want:      "",
			wantError: true,
		},
		{
			name: "non-image content still encodes",
			setup: func() string {
				p := filepath.Join(tmpDir, "text.txt")
				data := []byte("hello world")
				if err := os.WriteFile(p, data, 0644); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
				return p
			},
			want:      base64.StdEncoding.EncodeToString([]byte("hello world")),
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()

			got, err := ReadImageAsBase64(path)
			if (err != nil) != tt.wantError {
				t.Fatalf("error = %v, wantError = %v", err, tt.wantError)
			}
			if !tt.wantError && got != tt.want {
				t.Fatalf("got = %q, want = %q", got, tt.want)
			}
		})
	}
}
