package guids

func init() {
	register(Guide{
		Name:  "Cabal",
		Tool:  "cabal",
		Tools: []string{"ghc", "cabal-install"},
		Desc:  "Haskell build system and package manager",
		Commands: []CommandExample{
			{"Init new project", "cabal init"},
			{"Init with Stack", "stack new <name>"},
			{"Build", "cabal build"},
			{"Run", "cabal run"},
			{"Test", "cabal test"},
			{"REPL (ghci)", "cabal repl"},
			{"Add dependency", "cabal install <pkg>  (then add to .cabal)"},
			{"Format code", "fourmizz <file>"},
			{"Build docs", "cabal haddock"},
			{"Clean", "cabal clean"},
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
