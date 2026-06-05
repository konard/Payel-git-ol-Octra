'use strict';

// Builds the native Electron menu from the shared MENU_MODEL so the native menu
// and the in-window Zed-style menu bar (reference 3) stay identical.

const { MENU_MODEL } = require('../renderer/js/menuModel');

// Translate the data model into an Electron Menu template. `dispatch` receives the
// command id when an item is clicked so the main process can route it.
function toElectronTemplate(dispatch) {
  return MENU_MODEL.map((menu) => ({
    label: menu.label,
    submenu: menu.items.map((item) =>
      item.type === 'separator'
        ? { type: 'separator' }
        : {
            label: item.label,
            // Single-chord accelerators (e.g. "Ctrl-S") map cleanly to Electron's
            // "Ctrl+S"; multi-chord ones (e.g. "Ctrl-K Ctrl-S") aren't supported
            // as global accelerators, so we leave them visual-only.
            accelerator:
              item.accelerator && !item.accelerator.includes(' ')
                ? item.accelerator.replace(/-/g, '+')
                : undefined,
            click: () => dispatch(item.command),
          },
    ),
  }));
}

module.exports = { MENU_MODEL, toElectronTemplate };
