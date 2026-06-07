package java

import "orchestrator/pkg/guids/core"

func init() {
	core.Register(core.Guide{
		Name:  "Kotlin",
		Tool:  "kotlin",
		Tools: []string{"kotlin", "gradle", "jdk"},
		Desc:  "Kotlin — JVM language, builds with Gradle or kotlinc",
		Commands: []core.CommandExample{
			{"New JVM app (Gradle)", "gradle init --type kotlin-application"},
			{"New library (Gradle)", "gradle init --type kotlin-library"},
			{"New multiplatform project", "gradle init --type kotlin-multiplatform-library"},
			{"Build with Gradle", "gradle build"},
			{"Run with Gradle", "gradle run"},
			{"Test with Gradle", "gradle test"},
			{"Compile single file (kotlinc)", "kotlinc main.kt -include-runtime -d app.jar"},
			{"Run compiled jar", "java -jar app.jar"},
			{"New Spring Boot project", "curl start.spring.io -d type=gradle-project -d language=kotlin -d name=<name> | tar xz"},
			{"Add dependency", "add to build.gradle.kts dependencies block"},
		},
		Structure: `build.gradle.kts
settings.gradle.kts
gradle/
  wrapper/
src/
  main/
    kotlin/
      com/example/
        Main.kt
  test/
    kotlin/
      com/example/
        MainTest.kt`,
	})
}
