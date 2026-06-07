package ruby

import "orchestrator/pkg/guids/core"

func init() {
	core.Register(core.Guide{
		Name:  "Gem",
		Tool:  "gem",
		Tools: []string{"ruby"},
		Desc:  "RubyGems — Ruby package manager (gem command)",
		Commands: []core.CommandExample{
			{"Install gem", "gem install <gem>"},
			{"Install specific version", "gem install <gem> -v <version>"},
			{"Uninstall gem", "gem uninstall <gem>"},
			{"List installed gems", "gem list"},
			{"Search remote gems", "gem search <query>"},
			{"Build gem from gemspec", "gem build <name>.gemspec"},
			{"Push gem to rubygems.org", "gem push <name>-<version>.gem"},
			{"Install dev dep", "gem install <gem> --development"},
			{"Update all gems", "gem update"},
			{"Update specific gem", "gem update <gem>"},
			{"Show gem environment", "gem environment"},
			{"Clean old gem versions", "gem cleanup"},
			{"Create new gem skeleton", "gem install gem-ctags"},
		},
		Structure: `<name>.gemspec
Gemfile
Gemfile.lock
lib/
  <name>.rb
  <name>/
    version.rb
bin/
  <name>
test/ or spec/
README.md
LICENSE.txt`,
	})
}
