package kotlin

import "orchestrator/pkg/instralutions/core"

func init() {
	core.Register(core.InstallScript{
		Name:     "kotlin",
		Requires: []string{"jdk", "kotlin"},
		Commands: []string{
			`mkdir -p app/src/main/kotlin/com/octra`,
		},
	})
	core.Register(core.InstallScript{
		Name:     "kotlin-spring",
		Requires: []string{"jdk", "kotlin", "maven"},
		Commands: []string{
			`mvn -B archetype:generate -DgroupId=com.octra -DartifactId=app -DarchetypeArtifactId=maven-archetype-quickstart -DarchetypeVersion=1.5 -DinteractiveMode=false`,
		},
	})
}
