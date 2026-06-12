package ruby

import "orchestrator/pkg/instralutions/core"

func init() {
	core.Register(core.InstallScript{
		Name:     "ruby",
		Requires: []string{"ruby", "bundler"},
		Commands: []string{
			`bundle init`,
		},
	})
	core.Register(core.InstallScript{
		Name:     "ruby-rails",
		Requires: []string{"ruby", "bundler"},
		Commands: []string{
			`gem install rails --no-doc`,
			`rails new app --skip-git --skip-test --skip-action-mailer --skip-action-mailbox --skip-active-storage --skip-action-text`,
		},
	})
}
