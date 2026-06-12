package core

import "strings"

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
func ResolvePackages(names []string) []string {
	seen := map[string]bool{}
	var pkgs []string
	for _, name := range names {
		s := Get(name)
		if s == nil {
			continue
		}
		for _, p := range s.Requires {
			if seen[p] {
				continue
			}
			seen[p] = true
			pkgs = append(pkgs, p)
		}
	}
	return pkgs
}
