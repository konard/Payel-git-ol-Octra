package flutter

import "orchestrator/pkg/instralutions/core"

func init() {
	core.Register(core.InstallScript{
		Name:     "flutter",
		Requires: []string{"flutter"},
		Commands: []string{
			`flutter create --org com.octra --project-name app . 2>/dev/null || flutter create --org com.octra --project-name app`,
		},
	})
}
