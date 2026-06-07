package guids

func init() {
	register(Guide{
		Name:  "Mix",
		Tool:  "mix",
		Tools: []string{"elixir"},
		Desc:  "Elixir build tool and project manager",
		Commands: []CommandExample{
			{"New project", "mix new <name>"},
			{"New Phoenix app", "mix phx.new <name>"},
			{"New Phoenix without Ecto", "mix phx.new <name> --no-ecto"},
			{"Compile", "mix compile"},
			{"Run project", "mix run"},
			{"Run Phoenix server", "mix phx.server"},
			{"Run with iex REPL", "iex -S mix"},
			{"Test", "mix test"},
			{"Get dependencies", "mix deps.get"},
			{"Update dependencies", "mix deps.update --all"},
			{"Format code", "mix format"},
			{"Lint (credo)", "mix credo"},
			{"Check types (dialyzer)", "mix dialyzer"},
			{"Generate docs", "mix docs"},
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
