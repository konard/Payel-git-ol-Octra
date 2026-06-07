package guids

func init() {
	register(Guide{
		Name:  "Bundler",
		Tool:  "bundler",
		Tools: []string{"ruby", "bundler"},
		Desc:  "Ruby dependency manager and project initializer",
		Commands: []CommandExample{
			{"Init Gemfile", "bundle init"},
			{"Install deps", "bundle install"},
			{"Add gem", "bundle add <gem>"},
			{"Remove gem", "bundle remove <gem>"},
			{"Run with bundle", "bundle exec ruby main.rb"},
			{"Update all gems", "bundle update"},
			{"New Rails app", "rails new ."},
			{"Run Rails server", "rails server"},
			{"Generate scaffold", "rails generate scaffold <Model> <field>:<type>"},
			{"Run migrations", "rails db:migrate"},
			{"Test (RSpec)", "bundle exec rspec"},
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
