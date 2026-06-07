package swift

import "orchestrator/pkg/guids/core"

func init() {
	core.Register(core.Guide{
		Name:  "SwiftPM",
		Tool:  "swift",
		Tools: []string{"swift"},
		Desc:  "Swift Package Manager — build, run, test Swift projects",
		Commands: []core.CommandExample{
			{"New executable", "swift package init --type executable"},
			{"New library", "swift package init --type library"},
			{"New macro", "swift package init --type macro"},
			{"Build", "swift build"},
			{"Build release", "swift build -c release"},
			{"Run", "swift run"},
			{"Test", "swift test"},
			{"Format code", "swift format ."},
			{"Update deps", "swift package update"},
			{"Show deps graph", "swift package show-dependencies"},
			{"Generate Xcode project", "swift package generate-xcodeproj"},
		},
		Structure: `Package.swift
Sources/
  <Target>/
    main.swift       (executable)
    <Target>.swift   (library)
Tests/
  <Target>Tests/
    <Target>Tests.swift`,
	})
}
