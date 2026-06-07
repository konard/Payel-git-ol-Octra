package php

import "orchestrator/pkg/guids/core"

func init() {
	core.Register(core.Guide{
		Name:  "Composer",
		Tool:  "composer",
		Tools: []string{"php83", "php83Packages.composer"},
		Desc:  "PHP dependency manager",
		Commands: []core.CommandExample{
			{"Init new project", "composer init"},
			{"Create new Laravel project", "composer create-project laravel/laravel ."},
			{"Install dependencies", "composer install"},
			{"Add production dependency", "composer require <pkg>"},
			{"Add dev dependency", "composer require --dev <pkg>"},
			{"Remove dependency", "composer remove <pkg>"},
			{"Update all deps", "composer update"},
			{"Show outdated deps", "composer outdated"},
			{"Dump autoloader", "composer dump-autoload"},
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
