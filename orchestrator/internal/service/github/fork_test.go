package github

import (
	"errors"
	"testing"
)

func TestIsPermissionError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"403 status", errors.New("git push branch failed: exit status 128 - The requested URL returned error: 403"), true},
		{"permission denied phrase", errors.New("remote: Permission to Payel-git-ol/Octra.git denied to Octra-git."), true},
		{"write access not granted", errors.New("remote: Write access to repository not granted."), true},
		{"unrelated error", errors.New("fatal: could not resolve host"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsPermissionError(tc.err); got != tc.want {
				t.Fatalf("IsPermissionError(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}
