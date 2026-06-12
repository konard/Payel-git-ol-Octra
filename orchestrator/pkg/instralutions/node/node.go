package node

import "orchestrator/pkg/instralutions/core"

func init() {
	core.Register(core.InstallScript{
		Name:     "node",
		Requires: []string{"nodejs", "npm"},
		Commands: []string{
			`npm init -y`,
		},
	})
	core.Register(core.InstallScript{
		Name:     "node-react",
		Requires: []string{"nodejs", "npm"},
		Commands: []string{
			`npx --yes create-vite@latest app --template react`,
			`cd app && npm install`,
		},
	})
	core.Register(core.InstallScript{
		Name:     "node-next",
		Requires: []string{"nodejs", "npm"},
		Commands: []string{
			`npx --yes create-next-app@latest app --js --no-tailwind --eslint`,
			`cd app && npm install`,
		},
	})
	core.Register(core.InstallScript{
		Name:     "node-nest",
		Requires: []string{"nodejs", "npm"},
		Commands: []string{
			`npx --yes @nestjs/cli@latest new app --package-manager npm --skip-git`,
		},
	})
}
