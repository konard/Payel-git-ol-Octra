package golang

import "orchestrator/pkg/guids/core"

func init() {
	core.Register(core.Guide{
		Name:  "Go",
		Tool:  "go",
		Tools: []string{"go"},
		Desc:  "Go toolchain — compile, run, test, and manage modules",
		Commands: []core.CommandExample{
			{"New module", "go mod init <module>"},
			{"Build package", "go build ./..."},
			{"Build to file", "go build -o <output> ."},
			{"Run current package", "go run ."},
			{"Run specific file", "go run main.go"},
			{"Test all packages", "go test ./..."},
			{"Test with coverage", "go test -cover ./..."},
			{"Test verbose", "go test -v ./..."},
			{"Add dependency", "go get <module>@latest"},
			{"Update go.mod", "go mod tidy"},
			{"Vendor dependencies", "go mod vendor"},
			{"Format code", "go fmt ./..."},
			{"Vet code", "go vet ./..."},
			{"Install binary", "go install <path>"},
			{"Download modules", "go mod download"},
			{"Show module graph", "go mod graph"},
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
