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

	core.Register(core.Guide{
		Name:  "Artisan",
		Tool:  "artisan",
		Tools: []string{"php83", "php83Packages.composer"},
		Desc:  "Laravel command-line interface (php artisan)",
		Commands: []core.CommandExample{
			{"Start dev server", "php artisan serve"},
			{"Make controller", "php artisan make:controller <Name>Controller"},
			{"Make model", "php artisan make:model <Name>"},
			{"Make migration", "php artisan make:migration create_<table>_table"},
			{"Run migrations", "php artisan migrate"},
			{"Rollback migrations", "php artisan migrate:rollback"},
			{"Make seeder", "php artisan make:seeder <Name>Seeder"},
			{"Seed database", "php artisan db:seed"},
			{"Make factory", "php artisan make:factory <Name>Factory"},
			{"Make resource controller", "php artisan make:controller <Name>Controller --resource"},
			{"List all routes", "php artisan route:list"},
			{"Clear cache", "php artisan cache:clear"},
			{"Clear config cache", "php artisan config:clear"},
			{"Storage link", "php artisan storage:link"},
			{"Tinker REPL", "php artisan tinker"},
			{"Make middleware", "php artisan make:middleware <Name>"},
			{"Make mail", "php artisan make:mail <Name>"},
			{"Make notification", "php artisan make:notification <Name>"},
			{"Make request (validation)", "php artisan make:request <Name>Request"},
		},
		Structure: `app/Http/Controllers/
app/Models/
app/Http/Middleware/
database/migrations/
database/seeders/
routes/web.php
routes/api.php
resources/views/`,
	})
}
