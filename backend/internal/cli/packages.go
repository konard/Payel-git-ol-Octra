package cli

// CLIPackage describes how to install and run a CLI.
type CLIPackage struct {
	// Name is the display name used in the API.
	Name string
	// NixAttr is the nixpkgs attribute name. Empty means the package
	// is installed via an alternative method (npm, pip, go install, …).
	NixAttr string
	// InstallCmd is the fallback shell command used when NixAttr is
	// empty or when nixpkgs does not provide the package.
	InstallCmd string
	// Binary is the executable name on $PATH.
	Binary string
}

// BuiltinCLIs returns every CLI Octra can manage.
//
// CLIs without a NixpkgsAttr are provisioned via a shell command that
// temporarily brings the needed language runtime (Node.js, Go, Python, …)
// through `nix-shell -p` so the package manager works correctly inside the
// Nix-based runtime.
func BuiltinCLIs() []CLIPackage {
	return []CLIPackage{
		{Name: "codex", NixAttr: "codex", Binary: "codex"},
		{Name: "opencode", NixAttr: "", Binary: "opencode", InstallCmd: `curl -fsSL https://opencode.ai/install | bash`},
		{Name: "claude-code", NixAttr: "claude-code", Binary: "claude-code"},
		{Name: "cursor", NixAttr: "cursor", Binary: "cursor"},
		{Name: "antigravity", NixAttr: "", Binary: "antigravity", InstallCmd: `nix-shell -p curl bash --run "curl -fsSL https://antigravity.google/cli/install.sh | bash"`},
		{Name: "cline", NixAttr: "", Binary: "cline", InstallCmd: `nix-shell -p nodejs_22 --run "npm install -g cline"`},
		{Name: "openhands", NixAttr: "", Binary: "openhands", InstallCmd: `nix-shell -p python3 --run "python3 -m pip install openhands"`},
		{Name: "hermes", NixAttr: "hermes-agent", Binary: "hermes"},
		{Name: "ocawecore", NixAttr: "", Binary: "ocawecore", InstallCmd: `nix-shell -p crystal --run "cd /opt/ocawe && shards build --production && cp bin/ocawecore /usr/local/bin/ocawecore"`},
	}
}

// lookupCLIPackage finds the CLIPackage for a given CLI identifier.
func lookupCLIPackage(name string) (CLIPackage, bool) {
	for _, p := range BuiltinCLIs() {
		if p.Name == name {
			return p, true
		}
	}
	return CLIPackage{}, false
}

// NixpkgsAttr returns the nixpkgs attribute for the CLI, or empty if it's not
// available through nixpkgs.
func NixpkgsAttr(name string) string {
	p, ok := lookupCLIPackage(name)
	if !ok {
		return ""
	}
	return p.NixAttr
}
