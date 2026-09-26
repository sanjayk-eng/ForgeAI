package provider

import "testing"

func TestHasOAuthScope(t *testing.T) {
	tests := []struct {
		name        string
		header      string
		required    string
		wantPresent bool
	}{
		{name: "scope present", header: "repo, read:org, user:email", required: "read:org", wantPresent: true},
		{name: "scope with whitespace", header: "repo,  read:org ", required: "read:org", wantPresent: true},
		{name: "scope missing", header: "repo, user:email", required: "read:org"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := hasOAuthScope(test.header, test.required); got != test.wantPresent {
				t.Errorf("hasOAuthScope(%q, %q) = %v, want %v", test.header, test.required, got, test.wantPresent)
			}
		})
	}
}
