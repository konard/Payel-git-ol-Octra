package dotnet

import "orchestrator/pkg/instralutions/core"

func init() {
	core.Register(core.InstallScript{
		Name:     "dotnet",
		Requires: []string{"dotnet"},
		Commands: []string{
			`dotnet new console -n app --force`,
		},
	})
	core.Register(core.InstallScript{
		Name:     "dotnet-webapi",
		Requires: []string{"dotnet"},
		Commands: []string{
			`dotnet new webapi -n app --force`,
		},
	})
	core.Register(core.InstallScript{
		Name:     "dotnet-mvc",
		Requires: []string{"dotnet"},
		Commands: []string{
			`dotnet new mvc -n app --force`,
		},
	})
}
