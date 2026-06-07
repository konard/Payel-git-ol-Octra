package java

import "orchestrator/pkg/guids/core"

func init() {
	core.Register(core.Guide{
		Name:  "Maven",
		Tool:  "mvn",
		Tools: []string{"maven", "jdk"},
		Desc:  "Java build automation tool (Maven)",
		Commands: []core.CommandExample{
			{"New quickstart project", "mvn archetype:generate -DgroupId=com.example -DartifactId=<name> -DarchetypeArtifactId=maven-archetype-quickstart"},
			{"New webapp project", "mvn archetype:generate -DgroupId=com.example -DartifactId=<name> -DarchetypeArtifactId=maven-archetype-webapp"},
			{"Build (compile + test + package)", "mvn package"},
			{"Build without tests", "mvn package -DskipTests"},
			{"Compile only", "mvn compile"},
			{"Run tests", "mvn test"},
			{"Run specific test", "mvn test -Dtest=<ClassName>"},
			{"Clean build artifacts", "mvn clean"},
			{"Clean + package", "mvn clean package"},
			{"Install to local repo", "mvn install"},
			{"Run Spring Boot app", "mvn spring-boot:run"},
			{"Run exec:java", "mvn exec:java -Dexec.mainClass=\"com.example.Main\""},
			{"Generate site docs", "mvn site"},
			{"Download sources", "mvn dependency:sources"},
			{"Check for updates", "mvn versions:display-dependency-updates"},
		},
		Structure: `pom.xml
src/
  main/
    java/
      com/example/
        App.java
    resources/
      application.properties
  test/
    java/
      com/example/
        AppTest.java
target/             (built artifacts)`,
	})

	core.Register(core.Guide{
		Name:  "Gradle",
		Tool:  "gradle",
		Tools: []string{"gradle", "jdk"},
		Desc:  "Build automation for Java, Kotlin, and more",
		Commands: []core.CommandExample{
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
