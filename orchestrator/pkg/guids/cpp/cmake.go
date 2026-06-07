package cpp

import "orchestrator/pkg/guids/core"

func init() {
	core.Register(core.Guide{
		Name:  "CMake",
		Tool:  "cmake",
		Tools: []string{"cmake", "gcc", "ninja"},
		Desc:  "Cross-platform build system generator for C/C++",
		Commands: []core.CommandExample{
			{Purpose: "Configure (out-of-source)", Command: "cmake -B build -G Ninja"},
			{Purpose: "Configure with preset", Command: "cmake --preset <preset>"},
			{Purpose: "Build", Command: "cmake --build build"},
			{Purpose: "Build release", Command: "cmake --build build --config Release"},
			{Purpose: "Build target", Command: "cmake --build build --target <target>"},
			{Purpose: "Run tests", Command: "ctest --test-dir build"},
			{Purpose: "Install", Command: "cmake --install build"},
			{Purpose: "Clean", Command: "cmake --build build --target clean"},
		},
		Structure: `CMakeLists.txt
CMakePresets.json
src/
  main.cpp
include/
  <lib>/
    lib.h
tests/
  CMakeLists.txt
build/               (generated)`,
	})
}
