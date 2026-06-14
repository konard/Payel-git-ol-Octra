package node

import "orchestrator/pkg/instralutions/core"

func init() {
	// NOTE: "npm"/"npx" are bundled with the "nodejs" nixpkgs derivation — there is
	// no standalone npm package, so listing it would break flake.nix with
	// "error: undefined variable 'npm'". Only declare "nodejs".
	core.Register(core.InstallScript{
		Name:     "node",
		Requires: []string{"nodejs"},
		Commands: []string{
			`npm init -y`,
		},
	})
	core.Register(core.InstallScript{
		Name:     "node-react",
		Requires: []string{"nodejs"},
		Commands: []string{
			`npx --yes create-vite@latest app --template react`,
			`cd app && npm install`,
		},
	})
	core.Register(core.InstallScript{
		Name:     "node-next",
		Requires: []string{"nodejs"},
		Commands: []string{
			`npx --yes create-next-app@latest app --js --no-tailwind --eslint`,
			`cd app && npm install`,
		},
	})
	core.Register(core.InstallScript{
		Name:     "node-nest",
		Requires: []string{"nodejs"},
		Commands: []string{
			`npx --yes @nestjs/cli@latest new app --package-manager npm --skip-git`,
		},
	})
}
