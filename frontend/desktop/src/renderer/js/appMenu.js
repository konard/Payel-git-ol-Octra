/* Renders the Zed-style application menu (reference 3): the "Open Application
   Menu" hamburger plus a full File / Edit / Selection / View / Go / Run / Window /
   Help menu bar, all driven by the shared OctraMenuModel. */
(function (root) {
  const { MENU_MODEL } = root.OctraMenuModel;
  const Dropdown = root.OctraDropdown;

  function escapeHtml(s) {
    return String(s).replace(/[&<>"']/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
  }

  // Build the dropdown panel for one menu definition.
  function buildDropdown(menu, dispatch) {
    const el = document.createElement('div');
    el.className = 'dropdown';
    el.setAttribute('role', 'menu');
    el.dataset.menu = menu.label;
    menu.items.forEach((item) => {
      if (item.type === 'separator') {
        const sep = document.createElement('div');
        sep.className = 'menu-sep';
        el.appendChild(sep);
        return;
      }
      const row = document.createElement('div');
      row.className = 'menu-item';
      row.setAttribute('role', 'menuitem');
      row.dataset.command = item.command;
      row.innerHTML =
        `<span>${escapeHtml(item.label)}</span>` +
        (item.accelerator ? `<span class="menu-item__accel">${escapeHtml(item.accelerator)}</span>` : '');
      row.addEventListener('click', () => {
        Dropdown.close();
        dispatch(item.command);
      });
      el.appendChild(row);
    });
    return el;
  }

  // Render the in-window menu bar. `dispatch(command)` runs the chosen command.
  function renderMenuBar(container, dispatch) {
    container.innerHTML = '';
    const dropdownLayer = document.getElementById('dropdown-layer');

    MENU_MODEL.forEach((menu) => {
      const trigger = document.createElement('div');
      trigger.className = 'menubar__item' + (menu.app ? ' menubar__item--app' : '');
      trigger.setAttribute('role', 'button');
      trigger.setAttribute('aria-haspopup', 'true');
      trigger.setAttribute('aria-expanded', 'false');
      trigger.dataset.menu = menu.label;
      trigger.textContent = menu.label;

      const dropdown = buildDropdown(menu, dispatch);
      dropdownLayer.appendChild(dropdown);

      trigger.addEventListener('click', () => Dropdown.open(dropdown, trigger));
      // Zed behaviour: once a menu is open, hovering a sibling switches to it.
      trigger.addEventListener('mouseenter', () => {
        if (Dropdown.current() && !Dropdown.isOpen(dropdown)) Dropdown.open(dropdown, trigger);
      });

      container.appendChild(trigger);
    });
  }

  // The hamburger "Open Application Menu" opens the leading (app) menu — matching
  // reference 3 where the button reveals the Octra/Zed menu.
  function openApplicationMenu(trigger) {
    const appMenu = document.querySelector('.dropdown[data-menu="' + MENU_MODEL[0].label + '"]');
    if (appMenu) Dropdown.open(appMenu, trigger);
  }

  root.OctraAppMenu = { renderMenuBar, openApplicationMenu };
})(window);
