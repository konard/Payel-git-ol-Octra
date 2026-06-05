/* Boots the desktop chrome: assembles the Zed-style title bar (reference 2),
   mounts the application menu and project switcher, and routes menu commands. */
(function (root) {
  const Api = root.OctraApi;
  const Icons = root.OctraIcons;
  const Dropdown = root.OctraDropdown;
  const AppMenu = root.OctraAppMenu;
  const Switcher = root.OctraProjectSwitcher;

  // Route a menu/accelerator command to the right action. Most commands are
  // stubs for now — the issue scopes this PR to the chrome plus project I/O.
  function dispatch(command) {
    switch (command) {
      case 'project.new':
        Switcher.newProject();
        break;
      case 'project.open':
        Switcher.openFolder();
        break;
      case 'project.recent':
        toggleSwitcher();
        break;
      case 'window.close':
        Api.window.close();
        break;
      case 'window.minimize':
        Api.window.minimize();
        break;
      case 'app.quit':
        Api.window.close();
        break;
      case 'help.docs':
        Api.shell.openExternal('https://github.com/Payel-git-ol/Octra');
        break;
      default:
        // Surface the command in the status area so the UI feels responsive.
        setStatus(command);
    }
  }

  let switcherTrigger = null;
  function toggleSwitcher() {
    if (switcherTrigger) Switcher.toggle(switcherTrigger);
  }

  function setStatus(text) {
    const el = document.getElementById('status-command');
    if (el) el.textContent = text;
  }

  // Reflect the active project in the breadcrumb (reference 2: "octra · main · …").
  function renderBreadcrumb(project) {
    const nameEl = document.getElementById('project-name');
    if (nameEl) nameEl.textContent = project ? project.name : 'No Project';
    const titleEl = document.getElementById('workspace-title');
    if (titleEl) titleEl.textContent = project ? project.name : 'No project open';
    const pathEl = document.getElementById('workspace-path');
    if (pathEl) pathEl.textContent = project ? project.path : 'Open or create a project to get started';
  }

  function buildTitleBar() {
    // Hamburger — "Open Application Menu" (reference 3 entry point).
    const hamburger = document.getElementById('app-menu-button');
    hamburger.innerHTML = Icons.menu;
    hamburger.addEventListener('click', () => AppMenu.openApplicationMenu(hamburger));

    // Project breadcrumb chip opens the project switcher (reference 4).
    switcherTrigger = document.getElementById('project-chip');
    switcherTrigger.addEventListener('click', () => toggleSwitcher());

    document.getElementById('branch-remote').innerHTML = Icons.git + '<span>main</span>';
    document.getElementById('branch-local').innerHTML = Icons.branch + '<span>master</span>';

    // Window controls.
    const min = document.getElementById('win-min');
    const max = document.getElementById('win-max');
    const close = document.getElementById('win-close');
    min.innerHTML = Icons.minimize;
    max.innerHTML = Icons.maximize;
    close.innerHTML = Icons.close;
    min.addEventListener('click', () => Api.window.minimize());
    max.addEventListener('click', () => Api.window.toggleMaximize());
    close.addEventListener('click', () => Api.window.close());

    // The in-window menu bar (reference 3).
    AppMenu.renderMenuBar(document.getElementById('menubar'), dispatch);

    // Backdrop closes any open dropdown on outside click.
    document.querySelector('.backdrop').addEventListener('click', () => Dropdown.close());
  }

  function buildWorkspaceBody() {
    // A small sample file tree so the window reads as an IDE (reference 1).
    const tree = document.getElementById('file-tree');
    if (!tree) return;
    const rows = ['.github', 'src', 'docs', 'README.md', 'package.json'];
    tree.innerHTML = rows
      .map((name) => {
        const isFile = name.includes('.');
        const icon = isFile ? Icons.file : Icons.folder;
        return `<div class="tree-row">${icon}<span>${name}</span></div>`;
      })
      .join('');

    document.getElementById('cta-open').addEventListener('click', () => Switcher.openFolder());
    document.getElementById('cta-new').addEventListener('click', () => Switcher.newProject());
  }

  async function boot() {
    buildTitleBar();
    buildWorkspaceBody();
    await Switcher.init(renderBreadcrumb);
    Api.onMenuCommand(dispatch);
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', boot);
  } else {
    boot();
  }
})(window);
