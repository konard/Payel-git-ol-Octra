package python

import "orchestrator/pkg/guids/core"

func init() {
	core.Register(core.Guide{
		Name:  "pip",
		Tool:  "pip",
		Tools: []string{"python3", "pip"},
		Desc:  "Python package installer and virtual environment manager",
		Commands: []core.CommandExample{
			{Purpose: "Create virtual env", Command: "python -m venv .venv"},
			{Purpose: "Activate venv (Linux/macOS)", Command: "source .venv/bin/activate"},
			{Purpose: "Activate venv (Windows)", Command: ".venv\\Scripts\\activate"},
			{Purpose: "Install package", Command: "pip install <pkg>"},
			{Purpose: "Install dev requirements", Command: "pip install -r requirements.txt"},
			{Purpose: "Freeze requirements", Command: "pip freeze > requirements.txt"},
			{Purpose: "Uninstall package", Command: "pip uninstall <pkg>"},
			{Purpose: "List installed", Command: "pip list"},
			{Purpose: "Upgrade pip", Command: "pip install --upgrade pip"},
			{Purpose: "Install Django", Command: "pip install django"},
			{Purpose: "Install FastAPI", Command: "pip install fastapi uvicorn"},
			{Purpose: "Install Flask", Command: "pip install flask"},
		},
		Structure: `requirements.txt
.venv/
main.py             (app entry)
app/
  __init__.py
  routes.py
  models.py
manage.py           (Django)
project_name/       (Django settings)
  settings.py
  urls.py`,
	})
}
