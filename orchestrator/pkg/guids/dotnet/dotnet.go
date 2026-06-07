package dotnet

import "orchestrator/pkg/guids/core"

func init() {
	core.Register(core.Guide{
		Name:  "dotnet",
		Tool:  "dotnet",
		Tools: []string{"dotnet-sdk"},
		Desc:  ".NET SDK — build apps for web, mobile, desktop, cloud",
		Commands: []core.CommandExample{
			{"New console app", "dotnet new console -n <name>"},
			{"New web API", "dotnet new webapi -n <name>"},
			{"New MVC app", "dotnet new mvc -n <name>"},
			{"New class library", "dotnet new classlib -n <name>"},
			{"New Blazor app", "dotnet new blazor -n <name>"},
			{"Build", "dotnet build"},
			{"Run", "dotnet run"},
			{"Test", "dotnet test"},
			{"Add NuGet package", "dotnet add package <pkg>"},
			{"Remove NuGet package", "dotnet remove package <pkg>"},
			{"Publish release", "dotnet publish -c Release -o ./publish"},
			{"Restore packages", "dotnet restore"},
			{"Clean build artifacts", "dotnet clean"},
			{"List templates", "dotnet new list"},
		},
		Structure: `<name>.csproj
Program.cs
appsettings.json
appsettings.Development.json
Properties/launchSettings.json
Controllers/        (MVC/API)
Models/
Views/              (MVC)
wwwroot/            (static files)`,
	})
}
