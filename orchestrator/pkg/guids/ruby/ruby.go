package ruby

import "orchestrator/pkg/guids/core"

func init() {
	core.Register(core.Guide{
		Name:  "Bundler",
		Tool:  "bundler",
		Tools: []string{"ruby", "bundler"},
		Desc:  "Ruby dependency manager and project initializer",
		Commands: []core.CommandExample{
			{Purpose: "Init Gemfile", Command: "bundle init"},
			{Purpose: "Install deps", Command: "bundle install"},
			{Purpose: "Add gem", Command: "bundle add <gem>"},
			{Purpose: "Remove gem", Command: "bundle remove <gem>"},
			{Purpose: "Run with bundle", Command: "bundle exec ruby main.rb"},
			{Purpose: "Update all gems", Command: "bundle update"},
			{Purpose: "New Rails app", Command: "rails new ."},
			{Purpose: "Run Rails server", Command: "rails server"},
			{Purpose: "Generate scaffold", Command: "rails generate scaffold <Model> <field>:<type>"},
			{Purpose: "Run migrations", Command: "rails db:migrate"},
			{Purpose: "Test (RSpec)", Command: "bundle exec rspec"},
		},
		Structure: `Gemfile
Gemfile.lock
main.rb
app/                 (Rails)
  controllers/
  models/
  views/
config/
  routes.rb
  database.yml
db/
  migrate/`,
	})
}
