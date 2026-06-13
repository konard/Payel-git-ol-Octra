package core

import (
	"strings"
	"testing"
)

// TestSanitizeNixPackage verifies that tool aliases which are not standalone
// nixpkgs attributes are dropped or remapped, preventing the
// "error: undefined variable 'npm'" failure reported in issue #75.
func TestSanitizeNixPackage(t *testing.T) {
	cases := []struct {
		in      string
		wantPkg string
		wantOK  bool
	}{
		{"npm", "", false},     // bundled with nodejs
		{"npx", "", false},     // bundled with nodejs
		{"NPM", "", false},     // case-insensitive
		{"mix", "", false},     // bundled with elixir
		{"nodejs", "nodejs", true},
		{"jdk", "jdk", true},
		{"pip", "python3Packages.pip", true},
		{"cabal", "cabal-install", true},
		{"dotnet", "dotnet-sdk", true},
		{"composer", "php83Packages.composer", true},
		{"  nodejs  ", "nodejs", true}, // trims whitespace
		{"", "", false},
	}
	for _, c := range cases {
		gotPkg, gotOK := SanitizeNixPackage(c.in)
		if gotPkg != c.wantPkg || gotOK != c.wantOK {
			t.Errorf("SanitizeNixPackage(%q) = (%q, %v), want (%q, %v)",
				c.in, gotPkg, gotOK, c.wantPkg, c.wantOK)
		}
	}
}

// TestResolvePackagesNodeHasNoNpm is the regression test for issue #75 point 2:
// resolving the "node" install script must never emit a standalone "npm"
// package into flake.nix.
func TestResolvePackagesNodeHasNoNpm(t *testing.T) {
	Register(InstallScript{
		Name:     "test-node-npm",
		Requires: []string{"nodejs", "npm", "npx"},
	})
	pkgs := ResolvePackages([]string{"test-node-npm"})
	for _, p := range pkgs {
		if strings.EqualFold(p, "npm") || strings.EqualFold(p, "npx") {
			t.Fatalf("ResolvePackages leaked bundled tool %q into flake packages: %v", p, pkgs)
		}
	}
	found := false
	for _, p := range pkgs {
		if p == "nodejs" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected nodejs in resolved packages, got %v", pkgs)
	}
}
