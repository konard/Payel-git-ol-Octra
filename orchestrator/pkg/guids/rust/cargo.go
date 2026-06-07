package rust

import "orchestrator/pkg/guids/core"

func init() {
	core.Register(core.Guide{
		Name:  "Cargo",
		Tool:  "cargo",
		Tools: []string{"rustc", "cargo"},
		Desc:  "Rust package manager and build tool",
		Commands: []core.CommandExample{
			{Purpose: "New binary project", Command: "cargo init --name <name>"},
			{Purpose: "New library", Command: "cargo init --lib --name <name>"},
			{Purpose: "Build", Command: "cargo build"},
			{Purpose: "Release build", Command: "cargo build --release"},
			{Purpose: "Run", Command: "cargo run"},
			{Purpose: "Test", Command: "cargo test"},
			{Purpose: "Check without codegen", Command: "cargo check"},
			{Purpose: "Add dependency", Command: "cargo add <crate>"},
			{Purpose: "Remove dependency", Command: "cargo rm <crate>"},
			{Purpose: "Update dependencies", Command: "cargo update"},
			{Purpose: "Format code", Command: "cargo fmt"},
			{Purpose: "Lint", Command: "cargo clippy"},
			{Purpose: "Build docs", Command: "cargo doc --open"},
			{Purpose: "Publish crate", Command: "cargo publish"},
		},
		Structure: `Cargo.toml
src/main.rs         (binary target)
src/lib.rs          (library target)
tests/              (integration tests)
examples/           (example programs)
benches/            (benchmarks)`,
	})
}
