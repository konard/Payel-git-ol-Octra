package csharp

import "orchestrator/pkg/instralutions/core"

func init() {
	core.Register(core.InstallScript{
		Name:     "csharp",
		Requires: []string{"dotnet"},
		Commands: []string{
			`dotnet new console -n app --force`,
		},
	})
}
