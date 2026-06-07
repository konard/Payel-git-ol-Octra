package prompts

import "strings"

// ToolCheatSheet returns a concise reference of common tool commands for the
// given tech stack. The AI reads this instead of guessing commands, saving
// tokens and avoiding mistakes. Returns empty string for unknown stacks.
func ToolCheatSheet(techStack string) string {
	s := strings.ToLower(strings.TrimSpace(techStack))

	switch s {
	case "go", "golang":
		return `TOOL REFERENCE — Go:
  New module: go mod init <module>
  Build:      go build ./...
  Run:        go run .
  Test:       go test ./...
  Add dep:    go get <module>
  Tidy:       go mod tidy
  Format:     go fmt ./...
  Lint:       go vet ./...
  Install:    go install <path>
  Structure:  go.mod, main.go, cmd/<name>/main.go, internal/`

	case "node", "nodejs", "javascript":
		return `TOOL REFERENCE — Node.js / JavaScript:
  Init:        npm init -y
  Create app:  npx create-next-app@latest .  or  npx create-react-app .
  Run:         npm run dev  or  npm start  or  node index.js
  Build:       npm run build
  Test:        npm test
  Add dep:     npm install <pkg>
  Dev dep:     npm install -D <pkg>
  pnpm:        pnpm add <pkg>  /  pnpm add -D <pkg>
  Structure:   package.json, src/, index.js, .env`

	case "typescript":
		return `TOOL REFERENCE — TypeScript / Node.js:
  Init:        npm init -y  then  npx tsc --init
  Create app:  npx create-next-app@latest . --typescript
  Run:         npm run dev  or  npx tsx index.ts
  Build:       npx tsc  or  npm run build
  Test:        npm test
  Add dep:     npm install <pkg>
  Dev dep:     npm install -D <pkg>  npm install -D @types/<pkg>  npm install -D tsx
  Structure:   package.json, tsconfig.json, src/, .env`

	case "rust":
		return `TOOL REFERENCE — Rust:
  Init:      cargo init --name <name>  or  cargo init --lib
  Build:     cargo build  or  cargo build --release
  Run:       cargo run
  Test:      cargo test
  Add dep:   cargo add <crate>
  Check:     cargo check
  Format:    cargo fmt
  Lint:      cargo clippy
  Structure: Cargo.toml, src/main.rs or src/lib.rs`

	case "php":
		return `TOOL REFERENCE — PHP:
  Init:        composer init
  Create proj: composer create-project laravel/laravel .
  Run:         php -S localhost:8000  or  php artisan serve
  Add dep:     composer require <pkg>
  Dev dep:     composer require --dev <pkg>
  Test:        ./vendor/bin/phpunit
  Migrate:     php artisan migrate
  Structure:   composer.json, src/, public/index.php, artisan (Laravel)`

	case "flutter", "dart":
		return `TOOL REFERENCE — Flutter / Dart:
  Create:  flutter create <name>  or  flutter create --org com.example .
  Build:   flutter build [apk|ios|web|linux|macos|windows]
  Run:     flutter run  or  dart run
  Test:    flutter test  or  dart test
  Add dep: flutter pub add <pkg>
  Get:     flutter pub get
  Format:  dart format .
  Analyze: dart analyze
  Structure: pubspec.yaml, lib/, test/, android/, ios/`

	case "python":
		return `TOOL REFERENCE — Python:
  Init:      python -m venv .venv  then  source .venv/bin/activate
  Run:       python main.py  or  flask run  or  uvicorn main:app  or  python manage.py runserver
  Add dep:   pip install <pkg>
  Freeze:    pip freeze > requirements.txt
  Install:   pip install -r requirements.txt
  Test:      pytest  or  python -m pytest  or  python -m unittest
  Django:    django-admin startproject <name> .
  FastAPI:   pip install fastapi uvicorn
  Structure: requirements.txt, main.py, .venv/, manage.py (Django)`

	case "c++", "cpp":
		return `TOOL REFERENCE — C++:
  Init:   mkdir build && cd build && cmake .. && make
  Quick:  cmake -B build && cmake --build build
  Run:    ./build/<binary>
  Test:   ctest --test-dir build
  Structure: CMakeLists.txt, src/, include/, build/`

	case "c":
		return `TOOL REFERENCE — C:
  Compile: gcc -o <output> <file.c>
  Build:   mkdir build && cd build && cmake .. && make
  Run:     ./<output>
  Structure: CMakeLists.txt, src/, include/, build/`

	case "java":
		return `TOOL REFERENCE — Java:
  Maven init: mvn archetype:generate -DgroupId=com.example -DartifactId=<name> -DarchetypeArtifactId=maven-archetype-quickstart
  Gradle:     gradle init --type java-application
  Build:      mvn package  or  gradle build
  Run:        mvn exec:java -Dexec.mainClass="com.example.Main"  or  gradle run
  Test:       mvn test  or  gradle test
  Add dep:    add to pom.xml  or  build.gradle
  Structure:  pom.xml (Maven) or build.gradle (Gradle), src/main/java/, src/test/java/`

	case "kotlin":
		return `TOOL REFERENCE — Kotlin:
  Init:    gradle init --type kotlin-application
  Build:   gradle build
  Run:     gradle run
  Test:    gradle test
  Add dep: add to build.gradle.kts
  Structure: build.gradle.kts, src/main/kotlin/, src/test/kotlin/`

	case "ruby":
		return `TOOL REFERENCE — Ruby:
  Init:  bundle init
  Run:   ruby main.rb  or  rails server  or  bundle exec ruby main.rb
  Add:   bundle add <gem>
  Install: bundle install
  Test:  bundle exec rspec  or  ruby -Ilib:test test/test_*.rb
  Rails: rails new .  then  rails generate scaffold <Model>
  Structure: Gemfile, main.rb, app/ (Rails)`

	case "dotnet", "csharp":
		return `TOOL REFERENCE — .NET / C#:
  Init:    dotnet new console  or  dotnet new webapi  or  dotnet new mvc
  Build:   dotnet build
  Run:     dotnet run
  Test:    dotnet test
  Add pkg: dotnet add package <pkg>
  Publish: dotnet publish -c Release
  Structure: <Name>.csproj, Program.cs, Controllers/ (MVC)`

	case "zig":
		return `TOOL REFERENCE — Zig:
  Init:    zig init-exe  or  zig init-lib
  Build:   zig build
  Run:     zig build run
  Test:    zig build test
  Add dep: add to build.zig.zon
  Structure: build.zig, src/main.zig, build.zig.zon`

	case "elixir":
		return `TOOL REFERENCE — Elixir:
  Init:    mix new <name>  or  mix phx.new <name>
  Run:     mix run  or  mix phx.server (Phoenix)
  Build:   mix compile
  Test:    mix test
  Add dep: add to mix.exs then mix deps.get
  Structure: mix.exs, lib/, config/, test/`

	case "haskell":
		return `TOOL REFERENCE — Haskell:
  Init:    cabal init  or  stack new <name>
  Build:   cabal build  or  stack build
  Run:     cabal run  or  stack run
  Test:    cabal test  or  stack test
  Add dep: add to .cabal or package.yaml
  Structure: <name>.cabal, app/Main.hs, src/, test/`

	case "scala":
		return `TOOL REFERENCE — Scala:
  Init:    sbt new scala/scala-seed.g8
  Build:   sbt compile
  Run:     sbt run
  Test:    sbt test
  Add dep: add to build.sbt
  Structure: build.sbt, src/main/scala/, src/test/scala/`

	case "r":
		return `TOOL REFERENCE — R:
  Init:    write a .R file
  Run:     Rscript script.R
  Install: install.packages("<pkg>")
  Structure: script.R, DESCRIPTION, NAMESPACE (packages)`

	case "swift":
		return `TOOL REFERENCE — Swift:
  Init:    swift package init --type executable  or  --type library
  Build:   swift build
  Run:     swift run
  Test:    swift test
  Add dep: add to Package.swift
  Structure: Package.swift, Sources/, Tests/`

	default:
		return ""
	}
}

// ToolCheatSheetBrief returns a shorter one-liner description for use in
// manager or boss prompts where full detail isn't needed.
func ToolCheatSheetBrief(techStack string) string {
	briefs := map[string]string{
		"go":         "go mod init + go build/run/test",
		"golang":     "go mod init + go build/run/test",
		"node":       "npm init/install + npm run/build/test",
		"nodejs":     "npm init/install + npm run/build/test",
		"typescript": "npm init + tsc --init + npm run/build",
		"javascript": "npm init + npm run/build/test",
		"rust":       "cargo init/build/run/test + cargo add",
		"php":        "composer init/require + php artisan",
		"flutter":    "flutter create/build/run + flutter pub add",
		"dart":       "dart create/run + dart pub add",
		"c++":        "cmake -B build + cmake --build build",
		"cpp":        "cmake -B build + cmake --build build",
		"c":          "cmake + gcc",
		"java":       "mvn package/exec or gradle build/run",
		"kotlin":     "gradle build/run",
		"ruby":       "bundle init/add + ruby main.rb",
		"dotnet":     "dotnet new/build/run + dotnet add package",
		"csharp":     "dotnet new/build/run + dotnet add package",
		"elixir":     "mix new/run/test + mix deps.get",
		"haskell":    "cabal init/build/run or stack new/build",
		"python":     "pip install + python main.py",
		"zig":        "zig init-exe + zig build/run",
		"scala":      "sbt new + sbt compile/run",
		"swift":      "swift package init + swift build/run",
	}
	return briefs[strings.ToLower(strings.TrimSpace(techStack))]
}
