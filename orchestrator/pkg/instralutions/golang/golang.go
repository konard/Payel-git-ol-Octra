package golang

import "orchestrator/pkg/instralutions/core"

func init() {
	core.Register(core.InstallScript{
		Name:     "go",
		Requires: []string{"go"},
		Commands: []string{
			`mkdir app`,
			`cd app && go mod init github.com/octra/app`,
		},
	})
	core.Register(core.InstallScript{
		Name:     "go-gin",
		Requires: []string{"go"},
		Commands: []string{
			`mkdir app`,
			`cd app && go mod init github.com/octra/app`,
			`cd app && go get github.com/gin-gonic/gin`,
		},
	})
	core.Register(core.InstallScript{
		Name:     "go-echo",
		Requires: []string{"go"},
		Commands: []string{
			`mkdir app`,
			`cd app && go mod init github.com/octra/app`,
			`cd app && go get github.com/labstack/echo/v4`,
		},
	})
}
