package guids

func init() {
	register(Guide{
		Name:  "CMake",
		Tool:  "cmake",
		Tools: []string{"cmake", "gcc", "ninja"},
		Desc:  "Cross-platform build system generator for C/C++",
		Commands: []CommandExample{
			{"Configure (out-of-source)", "cmake -B build -G Ninja"},
			{"Configure with preset", "cmake --preset <preset>"},
			{"Build", "cmake --build build"},
			{"Build release", "cmake --build build --config Release"},
			{"Build target", "cmake --build build --target <target>"},
			{"Run tests", "ctest --test-dir build"},
			{"Install", "cmake --install build"},
			{"Clean", "cmake --build build --target clean"},
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
