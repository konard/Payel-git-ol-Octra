package guids

func init() {
	register(Guide{
		Name:  "Zig",
		Tool:  "zig",
		Tools: []string{"zig"},
		Desc:  "Zig build system and toolchain",
		Commands: []CommandExample{
			{"New executable", "zig init-exe"},
			{"New library", "zig init-lib"},
			{"Build", "zig build"},
			{"Run", "zig build run"},
			{"Test", "zig build test"},
			{"Run tests directly", "zig test src/main.zig"},
			{"Format code", "zig fmt ."},
			{"Cross-compile list", "zig targets"},
			{"Build for target", "zig build -Dtarget=<triple>"},
		},
		Structure: `build.zig
build.zig.zon
src/
  main.zig
  root.zig`,
	})
}
