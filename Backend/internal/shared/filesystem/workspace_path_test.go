package filesystem

import "testing"

func TestResolveWorkspacePath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "relative", input: "src/main.go", want: "/workspace/src/main.go"},
		{name: "absolute in workspace", input: "/workspace/src/main.go", want: "/workspace/src/main.go"},
		{name: "traversal", input: "src/../../outside", wantErr: true},
		{name: "outside absolute", input: "/etc/passwd", wantErr: true},
		{name: "git metadata", input: "src/.git/config", wantErr: true},
		{name: "windows path", input: `C:\outside.txt`, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ResolveWorkspacePath("/workspace", test.input)
			if (err != nil) != test.wantErr {
				t.Fatalf("ResolveWorkspacePath() error = %v, wantErr %v", err, test.wantErr)
			}
			if got != test.want {
				t.Fatalf("ResolveWorkspacePath() = %q, want %q", got, test.want)
			}
		})
	}
}
