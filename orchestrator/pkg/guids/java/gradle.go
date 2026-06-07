package java

import "orchestrator/pkg/guids/core"

func init() {
	core.Register(core.Guide{
		Name:  "Gradle",
		Tool:  "gradle",
		Tools: []string{"gradle", "jdk"},
		Desc:  "Build automation for Java, Kotlin, and more",
		Commands: []core.CommandExample{
			{Purpose: "Init Java app", Command: "gradle init --type java-application"},
			{Purpose: "Init Kotlin app", Command: "gradle init --type kotlin-application"},
			{Purpose: "Build", Command: "gradle build"},
			{Purpose: "Run", Command: "gradle run"},
			{Purpose: "Test", Command: "gradle test"},
			{Purpose: "Add dependency", Command: "add to build.gradle(.kts) dependencies block"},
			{Purpose: "Clean", Command: "gradle clean"},
			{Purpose: "Check", Command: "gradle check"},
			{Purpose: "List tasks", Command: "gradle tasks"},
		},
		Structure: `build.gradle or build.gradle.kts
settings.gradle.kts
gradle/
  wrapper/
    gradle-wrapper.jar
    gradle-wrapper.properties
src/
  main/
    java/ or kotlin/
    resources/
  test/
    java/ or kotlin/`,
	})
}
