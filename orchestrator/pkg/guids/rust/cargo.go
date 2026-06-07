package rust

import "orchestrator/pkg/guids/core"

func init() {
	core.Register(core.Guide{
		Name:  "Cargo",
		Tool:  "cargo",
		Tools: []string{"rustc", "cargo"},
		Desc:  "Rust package manager and build tool",
		Commands: []core.CommandExample{
			{"New binary project", "cargo init --name <name>"},
			{"New library", "cargo init --lib --name <name>"},
			{"Build", "cargo build"},
			{"Release build", "cargo build --release"},
			{"Run", "cargo run"},
			{"Test", "cargo test"},
			{"Check without codegen", "cargo check"},
			{"Add dependency", "cargo add <crate>"},
			{"Remove dependency", "cargo rm <crate>"},
			{"Update dependencies", "cargo update"},
			{"Format code", "cargo fmt"},
			{"Lint", "cargo clippy"},
			{"Build docs", "cargo doc --open"},
			{"Publish crate", "cargo publish"},
		},
		Structure: `Cargo.toml
src/main.rs         (binary target)
src/lib.rs          (library target)
tests/              (integration tests)
examples/           (example programs)
benches/            (benchmarks)`,
	})
}
