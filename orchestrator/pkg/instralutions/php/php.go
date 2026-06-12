package php

import "orchestrator/pkg/instralutions/core"

func init() {
	core.Register(core.InstallScript{
		Name:     "php",
		Requires: []string{"php", "composer"},
		Commands: []string{
			`composer init --name=octra/app --type=project --no-interaction`,
		},
	})
	core.Register(core.InstallScript{
		Name:     "php-laravel",
		Requires: []string{"php", "composer"},
		Commands: []string{
			`composer create-project laravel/laravel app --no-interaction`,
		},
	})
}
