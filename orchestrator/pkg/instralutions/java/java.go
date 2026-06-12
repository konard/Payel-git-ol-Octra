package java

import "orchestrator/pkg/instralutions/core"

func init() {
	core.Register(core.InstallScript{
		Name:     "java",
		Requires: []string{"jdk", "maven"},
		Commands: []string{
			`mvn -B archetype:generate -DgroupId=com.octra -DartifactId=app -DarchetypeArtifactId=maven-archetype-quickstart -DarchetypeVersion=1.5 -DinteractiveMode=false`,
		},
	})
	core.Register(core.InstallScript{
		Name:     "spring",
		Requires: []string{"jdk", "maven"},
		Commands: []string{
			`mvn -B archetype:generate -DgroupId=com.octra -DartifactId=app -DarchetypeArtifactId=maven-archetype-webapp -DarchetypeVersion=1.5 -DinteractiveMode=false`,
		},
	})
}
