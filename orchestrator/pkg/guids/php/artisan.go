package php

import "orchestrator/pkg/guids/core"

func init() {
	core.Register(core.Guide{
		Name:  "Artisan",
		Tool:  "artisan",
		Tools: []string{"php83", "php83Packages.composer"},
		Desc:  "Laravel command-line interface (php artisan)",
		Commands: []core.CommandExample{
			{Purpose: "Start dev server", Command: "php artisan serve"},
			{Purpose: "Make controller", Command: "php artisan make:controller <Name>Controller"},
			{Purpose: "Make model", Command: "php artisan make:model <Name>"},
			{Purpose: "Make migration", Command: "php artisan make:migration create_<table>_table"},
			{Purpose: "Run migrations", Command: "php artisan migrate"},
			{Purpose: "Rollback migrations", Command: "php artisan migrate:rollback"},
			{Purpose: "Make seeder", Command: "php artisan make:seeder <Name>Seeder"},
			{Purpose: "Seed database", Command: "php artisan db:seed"},
			{Purpose: "Make factory", Command: "php artisan make:factory <Name>Factory"},
			{Purpose: "Make resource controller", Command: "php artisan make:controller <Name>Controller --resource"},
			{Purpose: "List all routes", Command: "php artisan route:list"},
			{Purpose: "Clear cache", Command: "php artisan cache:clear"},
			{Purpose: "Clear config cache", Command: "php artisan config:clear"},
			{Purpose: "Storage link", Command: "php artisan storage:link"},
			{Purpose: "Tinker REPL", Command: "php artisan tinker"},
			{Purpose: "Make middleware", Command: "php artisan make:middleware <Name>"},
			{Purpose: "Make mail", Command: "php artisan make:mail <Name>"},
			{Purpose: "Make notification", Command: "php artisan make:notification <Name>"},
			{Purpose: "Make request (validation)", Command: "php artisan make:request <Name>Request"},
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
