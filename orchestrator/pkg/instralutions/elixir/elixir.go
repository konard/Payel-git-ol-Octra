package elixir

import "orchestrator/pkg/instralutions/core"

func init() {
	core.Register(core.InstallScript{
		Name:     "elixir",
		Requires: []string{"elixir", "mix"},
		Commands: []string{
			`mix new app`,
		},
	})
	core.Register(core.InstallScript{
		Name:     "elixir-phoenix",
		Requires: []string{"elixir", "mix", "nodejs"},
		Commands: []string{
			`mix archive.install --no-deps hex phx_new 2>/dev/null || true`,
			`mix phx.new app --no-ecto --no-live-reload --no-assets`,
		},
	})
}
