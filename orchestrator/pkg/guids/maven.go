package guids

func init() {
	register(Guide{
		Name:  "Maven",
		Tool:  "mvn",
		Tools: []string{"maven", "jdk"},
		Desc:  "Java build automation tool (Maven)",
		Commands: []CommandExample{
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
}
