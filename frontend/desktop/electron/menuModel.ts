'use strict';

// Single source of truth for the application-menu structure. The SAME model
// builds the native Electron menu (main process) and is exposed to the renderer
// so the in-window menu bar (DesktopTitleBar) never drifts from the OS menu.

export interface MenuItem {
  label?: string;
  command?: string;
  accelerator?: string;
  type?: 'separator';
  app?: boolean;
}

export interface MenuSection {
  label: string;
  app?: boolean;
  items: MenuItem[];
}

const SEP: MenuItem = { type: 'separator' };

// The leading "Octra" menu mirrors a Zed-style application menu, restyled for
// Octra (About / Check for Updates / Settings / Keymap / Themes / Quit).
export const MENU_MODEL: MenuSection[] = [
  {
    label: 'Octra',
    app: true,
    items: [
      { label: 'About Octra', command: 'app.about' },
      { label: 'Check for Updates', command: 'app.checkUpdates' },
      SEP,
      { label: 'Open Settings', accelerator: 'CmdOrCtrl+,', command: 'settings.open' },
      { label: 'Select Theme…', accelerator: 'CmdOrCtrl+K CmdOrCtrl+T', command: 'theme.select' },
      SEP,
      { label: 'Quit Octra', accelerator: 'CmdOrCtrl+Q', command: 'app.quit' },
    ],
  },
  {
    label: 'File',
    items: [
      { label: 'New Project…', accelerator: 'CmdOrCtrl+Shift+N', command: 'project.new' },
      { label: 'Open Folder…', accelerator: 'CmdOrCtrl+O', command: 'project.open' },
      { label: 'Open Recent', command: 'project.recent' },
      SEP,
      { label: 'Close Window', accelerator: 'CmdOrCtrl+Shift+W', command: 'window.close' },
    ],
  },
  {
    label: 'Edit',
    items: [
      { label: 'Undo', accelerator: 'CmdOrCtrl+Z', command: 'edit.undo' },
      { label: 'Redo', accelerator: 'CmdOrCtrl+Shift+Z', command: 'edit.redo' },
      SEP,
      { label: 'Cut', accelerator: 'CmdOrCtrl+X', command: 'edit.cut' },
      { label: 'Copy', accelerator: 'CmdOrCtrl+C', command: 'edit.copy' },
      { label: 'Paste', accelerator: 'CmdOrCtrl+V', command: 'edit.paste' },
    ],
  },
  {
    label: 'View',
    items: [
      { label: 'Toggle Explorer', accelerator: 'CmdOrCtrl+B', command: 'view.toggleExplorer' },
      { label: 'Reload', accelerator: 'CmdOrCtrl+R', command: 'view.reload' },
      { label: 'Toggle Developer Tools', accelerator: 'CmdOrCtrl+Alt+I', command: 'view.devtools' },
      SEP,
      { label: 'Zoom In', accelerator: 'CmdOrCtrl+=', command: 'view.zoomIn' },
      { label: 'Zoom Out', accelerator: 'CmdOrCtrl+-', command: 'view.zoomOut' },
      { label: 'Reset Zoom', accelerator: 'CmdOrCtrl+0', command: 'view.zoomReset' },
    ],
  },
  {
    label: 'Window',
    items: [
      { label: 'Minimize', accelerator: 'CmdOrCtrl+M', command: 'window.minimize' },
      { label: 'Zoom', command: 'window.toggleMaximize' },
    ],
  },
  {
    label: 'Help',
    items: [
      { label: 'Documentation', command: 'help.docs' },
      { label: 'Report Issue', command: 'help.report' },
    ],
  },
];
