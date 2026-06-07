package kotlin

import "orchestrator/pkg/guids/core"

func init() {
	core.Register(core.Guide{
		Name:  "Kotlin",
		Tool:  "kotlin",
		Tools: []string{"kotlin", "gradle", "jdk"},
		Desc:  "Kotlin — JVM language, builds with Gradle or kotlinc",
		Commands: []core.CommandExample{
			{Purpose: "New JVM app (Gradle)", Command: "gradle init --type kotlin-application"},
			{Purpose: "New library (Gradle)", Command: "gradle init --type kotlin-library"},
			{Purpose: "New multiplatform project", Command: "gradle init --type kotlin-multiplatform-library"},
			{Purpose: "Build with Gradle", Command: "gradle build"},
			{Purpose: "Run with Gradle", Command: "gradle run"},
			{Purpose: "Test with Gradle", Command: "gradle test"},
			{Purpose: "Compile single file (kotlinc)", Command: "kotlinc main.kt -include-runtime -d app.jar"},
			{Purpose: "Run compiled jar", Command: "java -jar app.jar"},
			{Purpose: "New Spring Boot project", Command: "curl start.spring.io -d type=gradle-project -d language=kotlin -d name=<name> | tar xz"},
			{Purpose: "Add dependency", Command: "add to build.gradle.kts dependencies block"},
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
