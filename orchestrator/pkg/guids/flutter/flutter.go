package flutter

import "orchestrator/pkg/guids/core"

func init() {
	core.Register(core.Guide{
		Name:  "Flutter",
		Tool:  "flutter",
		Tools: []string{"flutter", "dart"},
		Desc:  "Flutter SDK — cross-platform UI toolkit",
		Commands: []core.CommandExample{
			{"Create new project", "flutter create <name>"},
			{"Create with org domain", "flutter create --org com.example ."},
			{"Create for web only", "flutter create --platforms=web ."},
			{"Run on device/emulator", "flutter run"},
			{"Run with release mode", "flutter run --release"},
			{"Build APK (Android)", "flutter build apk"},
			{"Build iOS", "flutter build ios"},
			{"Build web", "flutter build web"},
			{"Build Linux", "flutter build linux"},
			{"Build Windows", "flutter build windows"},
			{"Build macOS", "flutter build macos"},
			{"Test", "flutter test"},
			{"Add dependency", "flutter pub add <pkg>"},
			{"Get all deps", "flutter pub get"},
			{"Upgrade deps", "flutter pub upgrade"},
			{"Analyze code", "flutter analyze"},
			{"Format code", "dart format ."},
			{"Clean build", "flutter clean"},
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
