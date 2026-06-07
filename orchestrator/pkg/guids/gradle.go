package guids

func init() {
	register(Guide{
		Name:  "Gradle",
		Tool:  "gradle",
		Tools: []string{"gradle", "jdk"},
		Desc:  "Build automation for Java, Kotlin, and more",
		Commands: []CommandExample{
			{"Init Java app", "gradle init --type java-application"},
			{"Init Kotlin app", "gradle init --type kotlin-application"},
			{"Build", "gradle build"},
			{"Run", "gradle run"},
			{"Test", "gradle test"},
			{"Add dependency", "add to build.gradle(.kts) dependencies block"},
			{"Clean", "gradle clean"},
			{"Check", "gradle check"},
			{"List tasks", "gradle tasks"},
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
