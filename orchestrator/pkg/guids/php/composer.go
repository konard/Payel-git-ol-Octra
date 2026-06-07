package php

import "orchestrator/pkg/guids/core"

func init() {
	core.Register(core.Guide{
		Name:  "Composer",
		Tool:  "composer",
		Tools: []string{"php83", "php83Packages.composer"},
		Desc:  "PHP dependency manager",
		Commands: []core.CommandExample{
			{Purpose: "Init new project", Command: "composer init"},
			{Purpose: "Create new Laravel project", Command: "composer create-project laravel/laravel ."},
			{Purpose: "Install dependencies", Command: "composer install"},
			{Purpose: "Add production dependency", Command: "composer require <pkg>"},
			{Purpose: "Add dev dependency", Command: "composer require --dev <pkg>"},
			{Purpose: "Remove dependency", Command: "composer remove <pkg>"},
			{Purpose: "Update all deps", Command: "composer update"},
			{Purpose: "Show outdated deps", Command: "composer outdated"},
			{Purpose: "Dump autoloader", Command: "composer dump-autoload"},
		},
		Structure: `composer.json
composer.lock
vendor/
src/
public/index.php
    (PHP project)

composer.json
artisan
bootstrap/
config/
database/
public/index.php
resources/
routes/
storage/
    (Laravel project)`,
	})
}
