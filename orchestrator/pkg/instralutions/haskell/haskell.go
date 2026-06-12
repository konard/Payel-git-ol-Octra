package haskell

import "orchestrator/pkg/instralutions/core"

func init() {
	core.Register(core.InstallScript{
		Name:     "haskell",
		Requires: []string{"ghc", "cabal"},
		Commands: []string{
			`cabal init --non-interactive --package-name=app --libandexe --license=NONE`,
		},
	})
	core.Register(core.InstallScript{
		Name:     "haskell-servant",
		Requires: []string{"ghc", "cabal"},
		Commands: []string{
			`cabal init --non-interactive --package-name=app --libandexe --license=NONE`,
			`cabal install servant servant-server`,
		},
	})
}
