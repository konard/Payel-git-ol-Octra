/* Boots the desktop shell: the Zed-style title bar (window chrome) plus the
   Octra web-style workspace (appbar, welcome screen, three-pane layout and the
   console dock). The whole UI mirrors the live web app (frontend/web), reuses
   the real Octra mascot, and only reveals the File/Edit/… menu bar and the
   three-pane workspace once a project is actually open. */
(function (root) {
  const Api = root.OctraApi;
  const Icons = root.OctraIcons;
  const Dropdown = root.OctraDropdown;
  const AppMenu = root.OctraAppMenu;
  const Switcher = root.OctraProjectSwitcher;

  // The real Octra mascot artwork (same PNG the web app ships in TopBar/Chat).
  const MASCOT_SRC = 'assets/octra-mascot.png';

  function $(id) {
    return document.getElementById(id);
  }

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
        break;
    }
  }

  let switcherTrigger = null;
  function toggleSwitcher() {
    if (switcherTrigger) Switcher.toggle(switcherTrigger);
  }

  // Reflect the active project across the chrome AND flip the whole shell between
  // the welcome screen (no project) and the three-pane workspace. Hiding the
  // menu bar until a project is open is an explicit PR #49 review requirement
  // ("если не открыт проект то меню файлов быть не должно").
  function applyProject(project) {
    const hasProject = !!project;
    document.body.setAttribute('data-has-project', hasProject ? 'true' : 'false');

    const nameEl = $('project-name');
    if (nameEl) nameEl.textContent = hasProject ? project.name : 'No Project';

    // Branch chips only carry meaning once a project is open.
    const remote = $('branch-remote');
    const local = $('branch-local');
    if (hasProject) {
      if (remote) remote.innerHTML = Icons.git + '<span>origin</span>';
      if (local) local.innerHTML = Icons.branch + '<span>main</span>';
    } else {
      if (remote) remote.innerHTML = '';
      if (local) local.innerHTML = '';
    }
  }

  function buildTitleBar() {
    // Hamburger — "Open Application Menu" (reference 3 entry point).
    const hamburger = $('app-menu-button');
    hamburger.innerHTML = Icons.menu;
    hamburger.addEventListener('click', () => AppMenu.openApplicationMenu(hamburger));

    // Project breadcrumb chip opens the project switcher (reference 4).
    switcherTrigger = $('project-chip');
    switcherTrigger.addEventListener('click', () => toggleSwitcher());

    // Window controls.
    const min = $('win-min');
    const max = $('win-max');
    const close = $('win-close');
    min.innerHTML = Icons.minimize;
    max.innerHTML = Icons.maximize;
    close.innerHTML = Icons.close;
    min.addEventListener('click', () => Api.window.minimize());
    max.addEventListener('click', () => Api.window.toggleMaximize());
    close.addEventListener('click', () => Api.window.close());

    // The in-window menu bar (reference 3) — CSS keeps it hidden until a project
    // is open via the body[data-has-project] gate.
    AppMenu.renderMenuBar($('menubar'), dispatch);

    // Backdrop closes any open dropdown on outside click.
    document.querySelector('.backdrop').addEventListener('click', () => Dropdown.close());
  }

  // The Octra application toolbar that mirrors the web TopBar.
  function buildAppbar() {
    $('appbar-mascot').src = MASCOT_SRC;

    const setIcon = (id, svg) => {
      const el = $(id);
      if (el) el.innerHTML = svg;
    };
    setIcon('toggle-sessions', Icons.panelLeft);
    setIcon('new-chat', Icons.messagePlus);
    setIcon('theme-toggle', Icons.sun);
    setIcon('open-settings', Icons.settings);
    setIcon('account', Icons.user);

    // Toggle the Sessions dock the same way the web header does.
    const toggle = $('toggle-sessions');
    toggle.addEventListener('click', () => {
      const open = document.body.getAttribute('data-sessions') !== 'closed';
      document.body.setAttribute('data-sessions', open ? 'closed' : 'open');
      toggle.classList.toggle('is-active', !open);
      toggle.setAttribute('aria-pressed', String(!open));
    });

    // Light/dark theme toggle (the web app defaults to dark too).
    $('theme-toggle').addEventListener('click', () => {
      const dark = document.documentElement.classList.toggle('dark');
      document.documentElement.classList.toggle('light', !dark);
      setIcon('theme-toggle', dark ? Icons.sun : Icons.moon);
    });

    // A new chat only makes sense with a project open; otherwise nudge to open.
    $('new-chat').addEventListener('click', () => {
      if (document.body.getAttribute('data-has-project') !== 'true') toggleSwitcher();
    });
  }

  // The welcome / no-project screen: real mascot, open/new actions, recents.
  function buildWelcome() {
    $('welcome-mascot').src = MASCOT_SRC;
    $('welcome-open').addEventListener('click', () => Switcher.openFolder());
    $('welcome-new').addEventListener('click', () => Switcher.newProject());
  }

  // Render the recent-projects list on the welcome card straight from the store
  // (no hardcoded entries — the list is empty until the user opens a project).
  async function renderRecents() {
    const host = $('welcome-recents');
    if (!host) return;
    let recent = [];
    try {
      recent = (await Api.projects.listRecent()) || [];
    } catch {
      recent = [];
    }
    if (!recent.length) {
      host.innerHTML = '<div class="welcome__recents-empty">No recent projects yet.</div>';
      return;
    }
    host.innerHTML =
      '<div class="welcome__recents-title">Recent projects</div>' +
      recent
        .map(
          (p) =>
            `<button class="recent-row" data-path="${escapeHtml(p.path)}" title="${escapeHtml(p.path)}">` +
            `<span class="recent-row__icon">${Icons.folder}</span>` +
            `<span class="recent-row__name">${escapeHtml(p.name)}</span>` +
            `<span class="recent-row__path">${escapeHtml(p.path)}</span></button>`,
        )
        .join('');
    host.querySelectorAll('.recent-row').forEach((row) => {
      row.addEventListener('click', async () => {
        const result = await Api.projects.openPath(row.dataset.path);
        if (result && result.project) onProjectChange(result.project);
      });
    });
  }

  function escapeHtml(s) {
    return String(s).replace(/[&<>"']/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
  }

  // Wire the three-pane workspace icons + the shared composer (mirrors the web
  // Sessions / Canvas+Chat / Solution layout and the bottom Console dock).
  function buildWorkspace() {
    $('chat-mascot').src = MASCOT_SRC;

    const setIcon = (id, svg) => {
      const el = $(id);
      if (el) el.innerHTML = svg;
    };
    setIcon('sessions-icon', Icons.history);
    setIcon('sessions-new-icon', Icons.plus);
    setIcon('solution-icon', Icons.file);
    setIcon('solution-empty-icon', Icons.box);
    setIcon('console-icon', Icons.terminal);
    setIcon('console-chevron', Icons.chevronUp);
    setIcon('env-icon', Icons.box);
    setIcon('env-caret', Icons.chevronDown);
    setIcon('model-icon', Icons.cpu);
    setIcon('model-caret', Icons.chevronDown);
    setIcon('attach', Icons.paperclip);
    setIcon('globe', Icons.globe);
    setIcon('send', Icons.send);
    setIcon('zoom-in', Icons.zoomIn);
    setIcon('zoom-out', Icons.zoomOut);
    setIcon('zoom-fit', Icons.maximize2);
    setIcon('zoom-export', Icons.download);

    // Auto-grow the composer textarea like the web BottomInput.
    const input = document.querySelector('.composer__input');
    if (input) {
      input.addEventListener('input', () => {
        input.style.height = 'auto';
        input.style.height = Math.min(input.scrollHeight, 160) + 'px';
      });
    }

    // Collapse / expand the console dock.
    const consoleToggle = $('console-toggle');
    if (consoleToggle) {
      consoleToggle.addEventListener('click', () => {
        const collapsed = document.body.toggleAttribute('data-console-collapsed');
        setIcon('console-chevron', collapsed ? Icons.chevronDown : Icons.chevronUp);
      });
    }

    // "Новый чат" in the Sessions dock opens the switcher when no project yet.
    const sessionsNew = $('sessions-new');
    if (sessionsNew) sessionsNew.addEventListener('click', () => dispatch('project.new'));
  }

  // Single entry point used whenever the active project changes.
  function onProjectChange(project) {
    applyProject(project);
    Dropdown.close();
  }

  async function boot() {
    buildTitleBar();
    buildAppbar();
    buildWelcome();
    buildWorkspace();
    await Switcher.init(applyProject);
    await renderRecents();
    Api.onMenuCommand(dispatch);
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', boot);
  } else {
    boot();
  }
})(window);
