package dotnet

import "orchestrator/pkg/guids/core"

func init() {
	core.Register(core.Guide{
		Name:  "dotnet",
		Tool:  "dotnet",
		Tools: []string{"dotnet-sdk"},
		Desc:  ".NET SDK — build apps for web, mobile, desktop, cloud",
		Commands: []core.CommandExample{
			{Purpose: "New console app", Command: "dotnet new console -n <name>"},
			{Purpose: "New web API", Command: "dotnet new webapi -n <name>"},
			{Purpose: "New MVC app", Command: "dotnet new mvc -n <name>"},
			{Purpose: "New class library", Command: "dotnet new classlib -n <name>"},
			{Purpose: "New Blazor app", Command: "dotnet new blazor -n <name>"},
			{Purpose: "Build", Command: "dotnet build"},
			{Purpose: "Run", Command: "dotnet run"},
			{Purpose: "Test", Command: "dotnet test"},
			{Purpose: "Add NuGet package", Command: "dotnet add package <pkg>"},
			{Purpose: "Remove NuGet package", Command: "dotnet remove package <pkg>"},
			{Purpose: "Publish release", Command: "dotnet publish -c Release -o ./publish"},
			{Purpose: "Restore packages", Command: "dotnet restore"},
			{Purpose: "Clean build artifacts", Command: "dotnet clean"},
			{Purpose: "List templates", Command: "dotnet new list"},
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
