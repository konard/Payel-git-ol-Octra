package node

import "orchestrator/pkg/guids/core"

func init() {
	core.Register(core.Guide{
		Name:  "npm",
		Tool:  "npm",
		Tools: []string{"nodejs"},
		Desc:  "Node.js package manager and project tool",
		Commands: []core.CommandExample{
			{Purpose: "Init package.json", Command: "npm init -y"},
			{Purpose: "Create Next.js app", Command: "npx create-next-app@latest ."},
			{Purpose: "Create React app", Command: "npx create-react-app ."},
			{Purpose: "Create Vite app", Command: "npm create vite@latest ."},
			{Purpose: "Create Nuxt app", Command: "npx nuxi@latest init ."},
			{Purpose: "Run dev server", Command: "npm run dev"},
			{Purpose: "Start production", Command: "npm start"},
			{Purpose: "Build", Command: "npm run build"},
			{Purpose: "Test", Command: "npm test"},
			{Purpose: "Install all deps", Command: "npm install"},
			{Purpose: "Add dependency", Command: "npm install <pkg>"},
			{Purpose: "Add dev dependency", Command: "npm install -D <pkg>"},
			{Purpose: "Remove dependency", Command: "npm uninstall <pkg>"},
			{Purpose: "Update all deps", Command: "npm update"},
			{Purpose: "pnpm add dep", Command: "pnpm add <pkg>"},
			{Purpose: "pnpm add dev dep", Command: "pnpm add -D <pkg>"},
			{Purpose: "pnpm run", Command: "pnpm run dev"},
		},
		Structure: `package.json
package-lock.json
node_modules/
src/
public/
index.js or index.ts
.next/              (Next.js build)
tsconfig.json       (TypeScript)`,
	})
}
