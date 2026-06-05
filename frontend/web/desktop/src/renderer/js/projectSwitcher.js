/* The project switcher dropdown (reference 4): a search box, the currently open
   project under "This Window", a "Recent Projects" list, and the Open Local
   Folders / New Project actions — wired to the filesystem through OctraApi. */
(function (root) {
  const Api = root.OctraApi;
  const Dropdown = root.OctraDropdown;
  const Icons = root.OctraIcons;

  function escapeHtml(s) {
    return String(s).replace(/[&<>"']/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
  }

  // State the switcher renders from. `active` is the project shown in the
  // breadcrumb (This Window); `recent` is the persisted recents list.
  const state = { active: null, recent: [], filter: '' };
  let onChange = () => {};

  let dropdownEl = null;
  let listEl = null;

  function matches(project) {
    const q = state.filter.trim().toLowerCase();
    if (!q) return true;
    return project.name.toLowerCase().includes(q) || project.path.toLowerCase().includes(q);
  }

  // Render the list portion (everything below the search box).
  function renderList() {
    const active = state.active;
    const others = state.recent.filter((p) => !active || p.path !== active.path);
    const thisWindow = active ? [active] : [];
    const filteredThis = thisWindow.filter(matches);
    const filteredRecent = others.filter(matches);

    let html = '';
    if (filteredThis.length) {
      html += '<div class="menu-section">This Window</div>';
      html += filteredThis
        .map(
          (p) =>
            `<div class="switcher__item switcher__item--active" data-path="${escapeHtml(p.path)}">` +
            `<span class="switcher__check">${Icons.check}</span>` +
            `<span class="switcher__name">${escapeHtml(p.name)}</span></div>`,
        )
        .join('');
    }
    if (filteredRecent.length) {
      html += '<div class="menu-section">Recent Projects</div>';
      html += filteredRecent
        .map(
          (p) =>
            `<div class="switcher__item" data-path="${escapeHtml(p.path)}" title="${escapeHtml(p.path)}">` +
            `<span class="switcher__check"></span>` +
            `<span class="switcher__name">${escapeHtml(p.name)}</span></div>`,
        )
        .join('');
    }
    if (!filteredThis.length && !filteredRecent.length) {
      html += '<div class="switcher__empty">No matching projects</div>';
    }

    // Fixed actions footer (reference 4).
    html +=
      '<div class="menu-sep"></div>' +
      '<div class="menu-item" data-action="open"><span>Open Local Folder…</span>' +
      '<span class="menu-item__accel">Ctrl-K Ctrl-O</span></div>' +
      '<div class="menu-item" data-action="new"><span>New Project…</span>' +
      '<span class="menu-item__accel">Ctrl-Shift-N</span></div>';

    listEl.innerHTML = html;
  }

  async function openFolder() {
    const result = await Api.projects.open();
    if (result && result.project) {
      setActive(result.project, result.recent);
    }
    Dropdown.close();
  }

  async function newProject() {
    const name = root.prompt ? root.prompt('New project name') : null;
    if (!name) return;
    const result = await Api.projects.create(name);
    if (result && result.project) {
      setActive(result.project, result.recent);
    }
    Dropdown.close();
  }

  async function selectPath(path) {
    if (state.active && state.active.path === path) {
      Dropdown.close();
      return;
    }
    const result = await Api.projects.openPath(path);
    if (result && result.project) {
      setActive(result.project, result.recent);
    }
    Dropdown.close();
  }

  function setActive(project, recent) {
    state.active = project;
    if (Array.isArray(recent)) state.recent = recent;
    renderList();
    onChange(project);
  }

  // Build the dropdown element once and attach its event handlers.
  function build() {
    dropdownEl = document.createElement('div');
    dropdownEl.className = 'dropdown switcher';
    dropdownEl.dataset.menu = 'project-switcher';
    dropdownEl.innerHTML =
      '<input class="switcher__search" type="text" placeholder="Search projects…" aria-label="Search projects">' +
      '<div class="switcher__list"></div>';

    listEl = dropdownEl.querySelector('.switcher__list');
    const search = dropdownEl.querySelector('.switcher__search');
    search.addEventListener('input', () => {
      state.filter = search.value;
      renderList();
    });

    listEl.addEventListener('click', (e) => {
      const actionEl = e.target.closest('[data-action]');
      if (actionEl) {
        if (actionEl.dataset.action === 'open') openFolder();
        else if (actionEl.dataset.action === 'new') newProject();
        return;
      }
      const itemEl = e.target.closest('[data-path]');
      if (itemEl) selectPath(itemEl.dataset.path);
    });

    document.getElementById('dropdown-layer').appendChild(dropdownEl);
    return dropdownEl;
  }

  // Initialise: load recents, pick the most recent as the active project.
  async function init(handler) {
    onChange = handler || (() => {});
    build();
    const recent = await Api.projects.listRecent();
    state.recent = Array.isArray(recent) ? recent : [];
    state.active = state.recent[0] || null;
    renderList();
    onChange(state.active);
  }

  function toggle(trigger) {
    state.filter = '';
    const search = dropdownEl.querySelector('.switcher__search');
    if (search) search.value = '';
    renderList();
    Dropdown.open(dropdownEl, trigger);
    if (Dropdown.isOpen(dropdownEl) && search) search.focus();
  }

  root.OctraProjectSwitcher = {
    init,
    toggle,
    openFolder,
    newProject,
    getActive: () => state.active,
  };
})(window);
