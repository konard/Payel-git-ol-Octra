package guids

func init() {
	register(Guide{
		Name:  "npm",
		Tool:  "npm",
		Tools: []string{"nodejs"},
		Desc:  "Node.js package manager and project tool",
		Commands: []CommandExample{
			{"Init package.json", "npm init -y"},
			{"Create Next.js app", "npx create-next-app@latest ."},
			{"Create React app", "npx create-react-app ."},
			{"Create Vite app", "npm create vite@latest ."},
			{"Create Nuxt app", "npx nuxi@latest init ."},
			{"Run dev server", "npm run dev"},
			{"Start production", "npm start"},
			{"Build", "npm run build"},
			{"Test", "npm test"},
			{"Install all deps", "npm install"},
			{"Add dependency", "npm install <pkg>"},
			{"Add dev dependency", "npm install -D <pkg>"},
			{"Remove dependency", "npm uninstall <pkg>"},
			{"Update all deps", "npm update"},
			{"pnpm add dep", "pnpm add <pkg>"},
			{"pnpm add dev dep", "pnpm add -D <pkg>"},
			{"pnpm run", "pnpm run dev"},
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
