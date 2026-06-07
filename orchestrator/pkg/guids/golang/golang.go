package golang

import "orchestrator/pkg/guids/core"

func init() {
	core.Register(core.Guide{
		Name:  "Go",
		Tool:  "go",
		Tools: []string{"go"},
		Desc:  "Go toolchain — compile, run, test, and manage modules",
		Commands: []core.CommandExample{
			{Purpose: "New module", Command: "go mod init <module>"},
			{Purpose: "Build package", Command: "go build ./..."},
			{Purpose: "Build to file", Command: "go build -o <output> ."},
			{Purpose: "Run current package", Command: "go run ."},
			{Purpose: "Run specific file", Command: "go run main.go"},
			{Purpose: "Test all packages", Command: "go test ./..."},
			{Purpose: "Test with coverage", Command: "go test -cover ./..."},
			{Purpose: "Test verbose", Command: "go test -v ./..."},
			{Purpose: "Add dependency", Command: "go get <module>@latest"},
			{Purpose: "Update go.mod", Command: "go mod tidy"},
			{Purpose: "Vendor dependencies", Command: "go mod vendor"},
			{Purpose: "Format code", Command: "go fmt ./..."},
			{Purpose: "Vet code", Command: "go vet ./..."},
			{Purpose: "Install binary", Command: "go install <path>"},
			{Purpose: "Download modules", Command: "go mod download"},
			{Purpose: "Show module graph", Command: "go mod graph"},
		},
		Structure: `go.mod
go.sum
main.go
cmd/
  <name>/
    main.go
internal/
pkg/
  <lib>/
    lib.go
tests/`,
	})
}
