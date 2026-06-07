package swift

import "orchestrator/pkg/guids/core"

func init() {
	core.Register(core.Guide{
		Name:  "SwiftPM",
		Tool:  "swift",
		Tools: []string{"swift"},
		Desc:  "Swift Package Manager — build, run, test Swift projects",
		Commands: []core.CommandExample{
			{Purpose: "New executable", Command: "swift package init --type executable"},
			{Purpose: "New library", Command: "swift package init --type library"},
			{Purpose: "New macro", Command: "swift package init --type macro"},
			{Purpose: "Build", Command: "swift build"},
			{Purpose: "Build release", Command: "swift build -c release"},
			{Purpose: "Run", Command: "swift run"},
			{Purpose: "Test", Command: "swift test"},
			{Purpose: "Format code", Command: "swift format ."},
			{Purpose: "Update deps", Command: "swift package update"},
			{Purpose: "Show deps graph", Command: "swift package show-dependencies"},
			{Purpose: "Generate Xcode project", Command: "swift package generate-xcodeproj"},
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
