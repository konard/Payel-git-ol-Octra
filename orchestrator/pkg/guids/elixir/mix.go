package elixir

import "orchestrator/pkg/guids/core"

func init() {
	core.Register(core.Guide{
		Name:  "Mix",
		Tool:  "mix",
		Tools: []string{"elixir"},
		Desc:  "Elixir build tool and project manager",
		Commands: []core.CommandExample{
			{Purpose: "New project", Command: "mix new <name>"},
			{Purpose: "New Phoenix app", Command: "mix phx.new <name>"},
			{Purpose: "New Phoenix without Ecto", Command: "mix phx.new <name> --no-ecto"},
			{Purpose: "Compile", Command: "mix compile"},
			{Purpose: "Run project", Command: "mix run"},
			{Purpose: "Run Phoenix server", Command: "mix phx.server"},
			{Purpose: "Run with iex REPL", Command: "iex -S mix"},
			{Purpose: "Test", Command: "mix test"},
			{Purpose: "Get dependencies", Command: "mix deps.get"},
			{Purpose: "Update dependencies", Command: "mix deps.update --all"},
			{Purpose: "Format code", Command: "mix format"},
			{Purpose: "Lint (credo)", Command: "mix credo"},
			{Purpose: "Check types (dialyzer)", Command: "mix dialyzer"},
			{Purpose: "Generate docs", Command: "mix docs"},
		},
		Structure: `mix.exs
mix.lock
lib/
  <app>/
    application.ex
    repo.ex
config/
  config.exs
  dev.exs
  prod.exs
  runtime.exs
test/
  test_helper.exs
priv/
  repo/
    migrations/`,
	})
}
