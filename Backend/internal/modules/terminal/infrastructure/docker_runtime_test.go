package infrastructure

import (
	"strings"
	"testing"
)

func TestParseContainerID(t *testing.T) {
	validID := strings.Repeat("a", 64)
	tests := []struct {
		name    string
		output  string
		wantID  string
		wantErr bool
	}{
		{name: "trims Docker output", output: validID + "\n", wantID: validID},
		{name: "rejects empty output", wantErr: true},
		{name: "rejects multiple lines", output: validID + "\nprogress output", wantErr: true},
		{name: "rejects values exceeding the database column", output: strings.Repeat("a", 256), wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseContainerID(test.output)
			if (err != nil) != test.wantErr {
				t.Fatalf("parseContainerID() error = %v, wantErr %v", err, test.wantErr)
			}
			if got != test.wantID {
				t.Fatalf("parseContainerID() = %q, want %q", got, test.wantID)
			}
		})
	}
}
