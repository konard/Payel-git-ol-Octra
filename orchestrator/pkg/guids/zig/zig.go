package zig

import "orchestrator/pkg/guids/core"

func init() {
	core.Register(core.Guide{
		Name:  "Zig",
		Tool:  "zig",
		Tools: []string{"zig"},
		Desc:  "Zig build system and toolchain",
		Commands: []core.CommandExample{
			{Purpose: "New executable", Command: "zig init-exe"},
			{Purpose: "New library", Command: "zig init-lib"},
			{Purpose: "Build", Command: "zig build"},
			{Purpose: "Run", Command: "zig build run"},
			{Purpose: "Test", Command: "zig build test"},
			{Purpose: "Run tests directly", Command: "zig test src/main.zig"},
			{Purpose: "Format code", Command: "zig fmt ."},
			{Purpose: "Cross-compile list", Command: "zig targets"},
			{Purpose: "Build for target", Command: "zig build -Dtarget=<triple>"},
		},
		Structure: `build.zig
build.zig.zon
src/
  main.zig
  root.zig`,
	})
}
