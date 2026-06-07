package guids

func init() {
	register(Guide{
		Name:  "pip",
		Tool:  "pip",
		Tools: []string{"python3", "pip"},
		Desc:  "Python package installer and virtual environment manager",
		Commands: []CommandExample{
			{"Create virtual env", "python -m venv .venv"},
			{"Activate venv (Linux/macOS)", "source .venv/bin/activate"},
			{"Activate venv (Windows)", ".venv\\Scripts\\activate"},
			{"Install package", "pip install <pkg>"},
			{"Install dev requirements", "pip install -r requirements.txt"},
			{"Freeze requirements", "pip freeze > requirements.txt"},
			{"Uninstall package", "pip uninstall <pkg>"},
			{"List installed", "pip list"},
			{"Upgrade pip", "pip install --upgrade pip"},
			{"Install Django", "pip install django"},
			{"Install FastAPI", "pip install fastapi uvicorn"},
			{"Install Flask", "pip install flask"},
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
