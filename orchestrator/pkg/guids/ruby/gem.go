package ruby

import "orchestrator/pkg/guids/core"

func init() {
	core.Register(core.Guide{
		Name:  "Gem",
		Tool:  "gem",
		Tools: []string{"ruby"},
		Desc:  "RubyGems — Ruby package manager (gem command)",
		Commands: []core.CommandExample{
			{Purpose: "Install gem", Command: "gem install <gem>"},
			{Purpose: "Install specific version", Command: "gem install <gem> -v <version>"},
			{Purpose: "Uninstall gem", Command: "gem uninstall <gem>"},
			{Purpose: "List installed gems", Command: "gem list"},
			{Purpose: "Search remote gems", Command: "gem search <query>"},
			{Purpose: "Build gem from gemspec", Command: "gem build <name>.gemspec"},
			{Purpose: "Push gem to rubygems.org", Command: "gem push <name>-<version>.gem"},
			{Purpose: "Install dev dep", Command: "gem install <gem> --development"},
			{Purpose: "Update all gems", Command: "gem update"},
			{Purpose: "Update specific gem", Command: "gem update <gem>"},
			{Purpose: "Show gem environment", Command: "gem environment"},
			{Purpose: "Clean old gem versions", Command: "gem cleanup"},
			{Purpose: "Create new gem skeleton", Command: "gem install gem-ctags"},
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
