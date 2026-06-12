package python

import "orchestrator/pkg/instralutions/core"

func init() {
	core.Register(core.InstallScript{
		Name:     "python",
		Requires: []string{"python3", "pip"},
		Commands: []string{
			`python3 -m venv venv`,
		},
	})
	core.Register(core.InstallScript{
		Name:     "python-django",
		Requires: []string{"python3", "pip"},
		Commands: []string{
			`python3 -m venv venv`,
			`. venv/bin/activate && pip install django && django-admin startproject app .`,
		},
	})
	core.Register(core.InstallScript{
		Name:     "python-flask",
		Requires: []string{"python3", "pip"},
		Commands: []string{
			`python3 -m venv venv`,
			`. venv/bin/activate && pip install flask`,
		},
	})
}
