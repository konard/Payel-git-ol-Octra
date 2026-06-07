package boss

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// TechStackPackages maps a tech stack name to the Nix packages needed
// for building and for the development shell.
type TechStackPackages struct {
	BuildInputs []string
	ShellInputs []string
}

// FlakeConfig controls how the flake.nix is generated.
type FlakeConfig struct {
	TaskID    string
	Title     string
	NixSystem string
	Packages  TechStackPackages
}

// FlakeBuilder generates Nix flake.nix files with proper build dependencies
// based on the project's tech stack.
type FlakeBuilder struct {
	packageMaps map[string]TechStackPackages
}

// NewFlakeBuilder creates a builder pre-populated with common tech stack mappings.
func NewFlakeBuilder() *FlakeBuilder {
	return &FlakeBuilder{
		packageMaps: map[string]TechStackPackages{
			"go":         {BuildInputs: []string{"go"}, ShellInputs: []string{"go", "gopls", "gotools"}},
			"golang":     {BuildInputs: []string{"go"}, ShellInputs: []string{"go", "gopls", "gotools"}},
			"python":     {BuildInputs: []string{"python3", "pip"}, ShellInputs: []string{"python3", "pip", "black", "ruff"}},
			"node":       {BuildInputs: []string{"nodejs"}, ShellInputs: []string{"nodejs", "typescript"}},
			"nodejs":     {BuildInputs: []string{"nodejs"}, ShellInputs: []string{"nodejs", "typescript"}},
			"typescript": {BuildInputs: []string{"nodejs", "typescript"}, ShellInputs: []string{"nodejs", "typescript"}},
			"javascript": {BuildInputs: []string{"nodejs"}, ShellInputs: []string{"nodejs"}},
			"rust":       {BuildInputs: []string{"rustc", "cargo"}, ShellInputs: []string{"rustc", "cargo", "rust-analyzer"}},
			"php":        {BuildInputs: []string{"php83", "php83Packages.composer"}, ShellInputs: []string{"php83", "php83Packages.composer"}},
			"flutter":    {BuildInputs: []string{"flutter"}, ShellInputs: []string{"flutter", "dart"}},
			"dart":       {BuildInputs: []string{"dart"}, ShellInputs: []string{"dart"}},
			"c++":        {BuildInputs: []string{"gcc", "cmake"}, ShellInputs: []string{"gcc", "cmake", "ninja", "gdb"}},
			"cpp":        {BuildInputs: []string{"gcc", "cmake"}, ShellInputs: []string{"gcc", "cmake", "ninja", "gdb"}},
			"c":          {BuildInputs: []string{"gcc", "cmake"}, ShellInputs: []string{"gcc", "cmake", "gdb"}},
			"java":       {BuildInputs: []string{"jdk", "maven"}, ShellInputs: []string{"jdk", "maven"}},
			"kotlin":     {BuildInputs: []string{"kotlin"}, ShellInputs: []string{"kotlin"}},
			"ruby":       {BuildInputs: []string{"ruby", "bundler"}, ShellInputs: []string{"ruby", "bundler"}},
			"dotnet":     {BuildInputs: []string{"dotnet-sdk"}, ShellInputs: []string{"dotnet-sdk"}},
			"csharp":     {BuildInputs: []string{"dotnet-sdk"}, ShellInputs: []string{"dotnet-sdk"}},
			"zig":        {BuildInputs: []string{"zig"}, ShellInputs: []string{"zig"}},
			"elixir":     {BuildInputs: []string{"elixir"}, ShellInputs: []string{"elixir"}},
			"haskell":    {BuildInputs: []string{"ghc", "cabal-install"}, ShellInputs: []string{"ghc", "cabal-install", "haskell-language-server"}},
			"scala":      {BuildInputs: []string{"scala", "sbt"}, ShellInputs: []string{"scala", "sbt"}},
			"r":          {BuildInputs: []string{"R"}, ShellInputs: []string{"R"}},
			"swift":      {BuildInputs: []string{"swift"}, ShellInputs: []string{"swift"}},
		},
	}
}

// ResolveFromTechStack looks up a single tech stack name and returns
// the corresponding Nix packages. Returns false if not found.
func (b *FlakeBuilder) ResolveFromTechStack(stack string) (TechStackPackages, bool) {
	pkg, ok := b.packageMaps[strings.ToLower(strings.TrimSpace(stack))]
	return pkg, ok
}

// ResolveFromTechStacks resolves multiple tech stacks and merges their
// packages into a single deduplicated set.
func (b *FlakeBuilder) ResolveFromTechStacks(stacks []string) TechStackPackages {
	var result TechStackPackages
	seenBuild := map[string]bool{}
	seenShell := map[string]bool{}

	for _, stack := range stacks {
		pkg, ok := b.ResolveFromTechStack(stack)
		if !ok {
			log.Printf("FlakeBuilder: unknown tech stack %q, skipping", stack)
			continue
		}
		for _, dep := range pkg.BuildInputs {
			if !seenBuild[dep] {
				result.BuildInputs = append(result.BuildInputs, dep)
				seenBuild[dep] = true
			}
		}
		for _, dep := range pkg.ShellInputs {
			if !seenShell[dep] {
				result.ShellInputs = append(result.ShellInputs, dep)
				seenShell[dep] = true
			}
		}
	}
	return result
}

// Generate produces the full flake.nix content based on the configuration.
func (b *FlakeBuilder) Generate(cfg FlakeConfig) string {
	nixSystem := cfg.NixSystem
	if nixSystem == "" {
		nixSystem = detectNixSystem()
	}

	if len(cfg.Packages.ShellInputs) == 0 && len(cfg.Packages.BuildInputs) == 0 {
		return b.generateMinimal(cfg.TaskID, cfg.Title, nixSystem)
	}

	buildInputsPkgs := b.formatPackages(cfg.Packages.BuildInputs)
	shellPkgs := b.formatPackages(cfg.Packages.ShellInputs)

	return fmt.Sprintf(`{
  description = "Octra project: %s - %s";
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  outputs = { self, nixpkgs }:
    let
      system = "%s";
      pkgs = nixpkgs.legacyPackages.${system};
    in {
      devShells.${system}.default = pkgs.mkShell {
        buildInputs = with pkgs; [
%s
        ];
        shellHook = ''
          echo "Octra development environment ready"
        '';
      };
      packages.${system}.default = pkgs.stdenv.mkDerivation {
        name = "octra-project-%s";
        src = ./.;
        nativeBuildInputs = with pkgs; [
%s
        ];
        installPhase = ''
          mkdir -p "$out"
          cp -r . "$out/"
        '';
      };
    };
}
`, sanitizeNixComment(cfg.TaskID), sanitizeNixComment(cfg.Title), nixSystem,
		shellPkgs, sanitizeNixComment(cfg.TaskID), buildInputsPkgs)
}

// generateMinimal creates the fallback flake.nix with no extra dependencies.
func (b *FlakeBuilder) generateMinimal(taskID, title, nixSystem string) string {
	return fmt.Sprintf(`{
  description = "Octra project: %s - %s";
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  outputs = { self, nixpkgs }:
    let
      system = "%s";
      pkgs = nixpkgs.legacyPackages.${system};
    in {
      packages.${system}.default = pkgs.stdenv.mkDerivation {
        name = "octra-project-%s";
        src = ./.;
        installPhase = ''
          mkdir -p "$out"
          cp -r . "$out/"
        '';
      };
    };
}
`, sanitizeNixComment(taskID), sanitizeNixComment(title), nixSystem, sanitizeNixComment(taskID))
}

// formatPackages formats a list of Nix package names for embedding in a flake.
func (b *FlakeBuilder) formatPackages(pkgs []string) string {
	if len(pkgs) == 0 {
		return ""
	}
	formatted := make([]string, len(pkgs))
	for i, p := range pkgs {
		formatted[i] = "          " + p
	}
	return strings.Join(formatted, "\n")
}

// WriteFlake generates and writes flake.nix to the project directory.
func (s *Service) WriteFlake(projectPath, taskID, title string, packages TechStackPackages) {
	cfg := FlakeConfig{
		TaskID:    taskID,
		Title:     title,
		NixSystem: detectNixSystem(),
		Packages:  packages,
	}
	content := NewFlakeBuilder().Generate(cfg)

	if err := os.WriteFile(filepath.Join(projectPath, "flake.nix"), []byte(content), 0644); err != nil {
		log.Printf("Failed to write flake.nix: %v", err)
		return
	}
	log.Printf("Generated flake.nix with build dependencies at: %s", filepath.Join(projectPath, "flake.nix"))
}
