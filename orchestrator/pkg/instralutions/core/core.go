package core

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// InstallScript describes how to scaffold a project for a given tech stack.
// Registered at init() time by each tech-stack package.
type InstallScript struct {
	Name     string   // "java", "node", "spring"
	Requires []string // nix packages to add to flake.nix: ["jdk", "maven"]
	Commands []string // bash commands to scaffold the project
}

var scripts []*InstallScript

// Register adds an InstallScript to the global registry.
func Register(s InstallScript) {
	scripts = append(scripts, &s)
}

// Get returns the InstallScript by name (case-insensitive).
func Get(name string) *InstallScript {
	for _, s := range scripts {
		if strings.EqualFold(s.Name, name) {
			return s
		}
	}
	return nil
}

// All returns all registered scripts.
func All() []*InstallScript {
	return scripts
}

// Resolve resolves a list of install names to a deduplicated list of commands.
// For example: ["java", "spring"] → java's commands + spring's commands.
func Resolve(names []string) []string {
	seen := map[string]bool{}
	var cmds []string
	for _, name := range names {
		s := Get(name)
		if s == nil {
			continue
		}
		for _, cmd := range s.Commands {
			if seen[cmd] {
				continue
			}
			seen[cmd] = true
			cmds = append(cmds, cmd)
		}
	}
	return cmds
}

// ResolvePackages returns all unique nix packages required by the given install names.
// Used by FlakeBuilder to ensure the flake includes the right buildInputs.
// Package names are sanitized via SanitizeNixPackage so that tool aliases that
// are NOT standalone nixpkgs attributes (e.g. "npm", "pip") never leak into
// flake.nix and break the build with "undefined variable".
func ResolvePackages(names []string) []string {
	seen := map[string]bool{}
	var pkgs []string
	for _, name := range names {
		s := Get(name)
		if s == nil {
			continue
		}
		for _, p := range s.Requires {
			pkg, ok := SanitizeNixPackage(p)
			if !ok || seen[pkg] {
				continue
			}
			seen[pkg] = true
			pkgs = append(pkgs, pkg)
		}
	}
	return pkgs
}

// nixPackageAliases maps common tool names to their real nixpkgs attribute, or
// to "" when the tool ships bundled with its runtime (so it must NOT be added as
// a separate package). Nix has no standalone "npm"/"mix"/"npx" derivations — they
// come with nodejs / elixir — and "pip"/"cabal"/"dotnet"/"composer" live under a
// different attribute path than their bare name.
var nixPackageAliases = map[string]string{
	"npm":      "", // bundled with nodejs
	"npx":      "", // bundled with nodejs
	"mix":      "", // bundled with elixir
	"gradlew":  "", // project wrapper, not a package
	"pip":      "python3Packages.pip",
	"pip3":     "python3Packages.pip",
	"cabal":    "cabal-install",
	"dotnet":   "dotnet-sdk",
	"composer": "php83Packages.composer",
	"go-tools": "gotools",
}

// SanitizeNixPackage maps a requested package/tool name to a valid nixpkgs
// attribute. It returns (attr, true) for a usable package and ("", false) when
// the name should be skipped entirely (e.g. tools bundled with their runtime).
func SanitizeNixPackage(name string) (string, bool) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", false
	}
	if mapped, ok := nixPackageAliases[strings.ToLower(trimmed)]; ok {
		if mapped == "" {
			return "", false
		}
		return mapped, true
	}
	return trimmed, true
}

// EnsureNixPackages updates the project's flake.nix to include the given Nix packages
// in both devShell.buildInputs and derivation.nativeBuildInputs.
// Also regenerates flake.lock via "nix flake lock".
// Non-fatal: logs warnings on failure, the pipeline continues.
func EnsureNixPackages(projectPath string, packages []string) {
	if len(packages) == 0 {
		return
	}
	flakePath := filepath.Join(projectPath, "flake.nix")
	content, err := os.ReadFile(flakePath)
	if err != nil {
		log.Printf("[instralutions] flake.nix not found at %s, skipping: %v", flakePath, err)
		return
	}
	updated := addPackagesToFlake(string(content), packages)
	if updated == string(content) {
		return // nothing changed
	}
	if err := os.WriteFile(flakePath, []byte(updated), 0644); err != nil {
		log.Printf("[instralutions] failed to write flake.nix: %v", err)
		return
	}
	log.Printf("[instralutions] Added packages to flake.nix: %v", packages)

	// Regenerate flake.lock so nix develop picks up the new packages
	cmd := exec.Command("git", "-C", projectPath, "add", "flake.nix")
	cmd.Run()
	lock := exec.Command("nix", "flake", "lock",
		"--extra-experimental-features", "nix-command flakes")
	lock.Dir = projectPath
	if out, err := lock.CombinedOutput(); err != nil {
		log.Printf("[instralutions] nix flake lock failed (non-fatal): %v\n%s", err, string(out))
	} else {
		log.Printf("[instralutions] flake.lock regenerated")
	}
}

// flakeSectionRe matches sections like:
//
//	buildInputs = with pkgs; [
//	  jdk
//	];
var flakeSectionRe = regexp.MustCompile(`(?s)((?:buildInputs|nativeBuildInputs)\s*=\s*with\s+pkgs;\s*\[)(.*?)(\])`)

// addPackagesToFlake inserts missing packages into both buildInputs and nativeBuildInputs sections.
func addPackagesToFlake(flake string, packages []string) string {
	return flakeSectionRe.ReplaceAllStringFunc(flake, func(match string) string {
		parts := flakeSectionRe.FindStringSubmatch(match)
		if len(parts) < 4 {
			return match
		}
		header := parts[1] // "buildInputs = with pkgs; ["
		body := parts[2]   // existing entries
		closer := parts[3] // "]"

		// Collect existing package names
		existing := map[string]bool{}
		for _, line := range strings.Split(body, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			existing[line] = true
		}

		// Add missing packages
		var added []string
		for _, p := range packages {
			if existing[p] {
				continue
			}
			existing[p] = true
			added = append(added, p)
		}
		if len(added) == 0 {
			return match
		}

		// Append to body
		var sb strings.Builder
		sb.WriteString(header)
		sb.WriteString(body)
		if body != "" && !strings.HasSuffix(body, "\n") {
			sb.WriteString("\n")
		}
		for _, p := range added {
			sb.WriteString(fmt.Sprintf("  %s\n", p))
		}
		sb.WriteString(closer)
		return sb.String()
	})
}
