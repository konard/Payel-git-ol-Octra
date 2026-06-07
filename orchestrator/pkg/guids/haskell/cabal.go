package haskell

import "orchestrator/pkg/guids/core"

func init() {
	core.Register(core.Guide{
		Name:  "Cabal",
		Tool:  "cabal",
		Tools: []string{"ghc", "cabal-install"},
		Desc:  "Haskell build system and package manager",
		Commands: []core.CommandExample{
			{Purpose: "Init new project", Command: "cabal init"},
			{Purpose: "Init with Stack", Command: "stack new <name>"},
			{Purpose: "Build", Command: "cabal build"},
			{Purpose: "Run", Command: "cabal run"},
			{Purpose: "Test", Command: "cabal test"},
			{Purpose: "REPL (ghci)", Command: "cabal repl"},
			{Purpose: "Add dependency", Command: "cabal install <pkg>  (then add to .cabal)"},
			{Purpose: "Format code", Command: "fourmizz <file>"},
			{Purpose: "Build docs", Command: "cabal haddock"},
			{Purpose: "Clean", Command: "cabal clean"},
		},
		Structure: `<name>.cabal
cabal.project
app/
  Main.hs
src/
  Lib.hs
test/
  Spec.hs
ChangeLog.md
README.md`,
	})
}
