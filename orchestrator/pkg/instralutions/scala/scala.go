package scala

import "orchestrator/pkg/instralutions/core"

func init() {
	core.Register(core.InstallScript{
		Name:     "scala",
		Requires: []string{"sbt", "jdk"},
		Commands: []string{
			`sbt new scala/hello-world.g8 --name=app 2>/dev/null || mkdir -p app/src/main/scala`,
		},
	})
}
