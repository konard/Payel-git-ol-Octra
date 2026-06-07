package java

import "orchestrator/pkg/guids/core"

func init() {
	core.Register(core.Guide{
		Name:  "Maven",
		Tool:  "mvn",
		Tools: []string{"maven", "jdk"},
		Desc:  "Java build automation tool (Maven)",
		Commands: []core.CommandExample{
			{Purpose: "New quickstart project", Command: "mvn archetype:generate -DgroupId=com.example -DartifactId=<name> -DarchetypeArtifactId=maven-archetype-quickstart"},
			{Purpose: "New webapp project", Command: "mvn archetype:generate -DgroupId=com.example -DartifactId=<name> -DarchetypeArtifactId=maven-archetype-webapp"},
			{Purpose: "Build (compile + test + package)", Command: "mvn package"},
			{Purpose: "Build without tests", Command: "mvn package -DskipTests"},
			{Purpose: "Compile only", Command: "mvn compile"},
			{Purpose: "Run tests", Command: "mvn test"},
			{Purpose: "Run specific test", Command: "mvn test -Dtest=<ClassName>"},
			{Purpose: "Clean build artifacts", Command: "mvn clean"},
			{Purpose: "Clean + package", Command: "mvn clean package"},
			{Purpose: "Install to local repo", Command: "mvn install"},
			{Purpose: "Run Spring Boot app", Command: "mvn spring-boot:run"},
			{Purpose: "Run exec:java", Command: "mvn exec:java -Dexec.mainClass=\"com.example.Main\""},
			{Purpose: "Generate site docs", Command: "mvn site"},
			{Purpose: "Download sources", Command: "mvn dependency:sources"},
			{Purpose: "Check for updates", Command: "mvn versions:display-dependency-updates"},
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
}
