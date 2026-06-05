package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixturePath(name string) string {
	return filepath.Join("..", "..", "..", "tests", "fixtures", name)
}

func TestFixtures_DeviceCSV_Parse(t *testing.T) {
	cases := []struct {
		file       string
		wantRows   int
		wantErrors int
		errContain string
	}{
		{"devices_valid.csv", 1, 0, ""},
		{"devices_invalid_header.csv", 0, 1, "header must be"},
		{"devices_duplicate.csv", 1, 1, "duplicate device_id"},
		{"devices_missing_required.csv", 0, 1, "required"},
	}

	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			data, err := os.ReadFile(fixturePath(tc.file))
			if err != nil {
				t.Fatalf("ReadFile() error = %v", err)
			}
			rows, rowErrors, err := parseDeviceCSV(strings.NewReader(string(data)))
			if err != nil {
				t.Fatalf("parseDeviceCSV() error = %v", err)
			}
			if len(rows) != tc.wantRows {
				t.Fatalf("len(rows) = %d, want %d", len(rows), tc.wantRows)
			}
			if len(rowErrors) != tc.wantErrors {
				t.Fatalf("len(rowErrors) = %d, want %d; errors=%+v", len(rowErrors), tc.wantErrors, rowErrors)
			}
			if tc.errContain != "" && len(rowErrors) > 0 && !strings.Contains(rowErrors[0].Message, tc.errContain) {
				t.Fatalf("message = %q, want contain %q", rowErrors[0].Message, tc.errContain)
			}
		})
	}
}
