package flutter

import "orchestrator/pkg/guids/core"

func init() {
	core.Register(core.Guide{
		Name:  "Flutter",
		Tool:  "flutter",
		Tools: []string{"flutter", "dart"},
		Desc:  "Flutter SDK — cross-platform UI toolkit",
		Commands: []core.CommandExample{
			{Purpose: "Create new project", Command: "flutter create <name>"},
			{Purpose: "Create with org domain", Command: "flutter create --org com.example ."},
			{Purpose: "Create for web only", Command: "flutter create --platforms=web ."},
			{Purpose: "Run on device/emulator", Command: "flutter run"},
			{Purpose: "Run with release mode", Command: "flutter run --release"},
			{Purpose: "Build APK (Android)", Command: "flutter build apk"},
			{Purpose: "Build iOS", Command: "flutter build ios"},
			{Purpose: "Build web", Command: "flutter build web"},
			{Purpose: "Build Linux", Command: "flutter build linux"},
			{Purpose: "Build Windows", Command: "flutter build windows"},
			{Purpose: "Build macOS", Command: "flutter build macos"},
			{Purpose: "Test", Command: "flutter test"},
			{Purpose: "Add dependency", Command: "flutter pub add <pkg>"},
			{Purpose: "Get all deps", Command: "flutter pub get"},
			{Purpose: "Upgrade deps", Command: "flutter pub upgrade"},
			{Purpose: "Analyze code", Command: "flutter analyze"},
			{Purpose: "Format code", Command: "dart format ."},
			{Purpose: "Clean build", Command: "flutter clean"},
		},
		Structure: `pubspec.yaml
lib/
  main.dart
  app.dart
test/
android/
ios/
web/
linux/
macos/
windows/
assets/
  images/
  fonts/`,
	})
}
