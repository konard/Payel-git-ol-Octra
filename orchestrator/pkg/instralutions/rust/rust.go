package rust

import "orchestrator/pkg/instralutions/core"

func init() {
	core.Register(core.InstallScript{
		Name:     "rust",
		Requires: []string{"cargo", "rustc"},
		Commands: []string{
			`cargo init --name app`,
		},
	})
	core.Register(core.InstallScript{
		Name:     "rust-rocket",
		Requires: []string{"cargo", "rustc"},
		Commands: []string{
			`cargo init --name app`,
			`cargo add rocket`,
		},
	})
	core.Register(core.InstallScript{
		Name:     "rust-axum",
		Requires: []string{"cargo", "rustc"},
		Commands: []string{
			`cargo init --name app`,
			`cargo add axum tokio serde serde_json`,
		},
	})
}
