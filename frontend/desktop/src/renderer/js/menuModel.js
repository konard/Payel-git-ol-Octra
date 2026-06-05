/*
 * Single source of truth for the application-menu structure (reference 3).
 *
 * Written as a UMD-style module so the SAME data drives both sides:
 *   - the Electron main process (CommonJS `require`) builds the native menu;
 *   - the renderer (classic <script>) draws the in-window Zed-style menu bar.
 * Sharing one model guarantees the native and in-window menus never drift.
 */
(function (root, factory) {
  const exported = factory();
  if (typeof module !== 'undefined' && module.exports) {
    module.exports = exported;
  } else {
    root.OctraMenuModel = exported;
  }
})(typeof globalThis !== 'undefined' ? globalThis : this, function () {
  const SEP = { type: 'separator' };

  // The leading "Octra" menu mirrors reference 3's "Zed" menu, restyled for
  // Octra (About Octra / Check for Updates / Settings / Keymap / Themes / Quit).
  const MENU_MODEL = [
    {
      label: 'Octra',
      app: true,
      items: [
        { label: 'About Octra', command: 'app.about' },
        { label: 'Check for Updates', command: 'app.checkUpdates' },
        SEP,
        { label: 'Open Settings', accelerator: 'Ctrl-,', command: 'settings.open' },
        { label: 'Open Settings File', accelerator: 'Ctrl-Alt-,', command: 'settings.openFile' },
        { label: 'Open Project Settings', command: 'settings.openProject' },
        { label: 'Open Project Settings File', command: 'settings.openProjectFile' },
        { label: 'Open Default Settings', command: 'settings.openDefaults' },
        SEP,
        { label: 'Open Keymap', accelerator: 'Ctrl-K Ctrl-S', command: 'keymap.open' },
        { label: 'Open Keymap File', command: 'keymap.openFile' },
        { label: 'Open Default Key Bindings', command: 'keymap.openDefaults' },
        SEP,
        { label: 'Select Theme…', accelerator: 'Ctrl-K Ctrl-T', command: 'theme.select' },
        { label: 'Select Icon Theme…', command: 'theme.selectIcons' },
        SEP,
        { label: 'Extensions', accelerator: 'Ctrl-Shift-X', command: 'extensions.open' },
        SEP,
        { label: 'Quit Octra', accelerator: 'Ctrl-Q', command: 'app.quit' },
      ],
    },
    {
      label: 'File',
      items: [
        { label: 'New Project…', accelerator: 'Ctrl-Shift-N', command: 'project.new' },
        { label: 'Open Folder…', accelerator: 'Ctrl-K Ctrl-O', command: 'project.open' },
        { label: 'Open Recent', command: 'project.recent' },
        SEP,
        { label: 'Save', accelerator: 'Ctrl-S', command: 'file.save' },
        { label: 'Save All', accelerator: 'Ctrl-Alt-S', command: 'file.saveAll' },
        SEP,
        { label: 'Close Window', accelerator: 'Ctrl-Shift-W', command: 'window.close' },
      ],
    },
    {
      label: 'Edit',
      items: [
        { label: 'Undo', accelerator: 'Ctrl-Z', command: 'edit.undo' },
        { label: 'Redo', accelerator: 'Ctrl-Shift-Z', command: 'edit.redo' },
        SEP,
        { label: 'Cut', accelerator: 'Ctrl-X', command: 'edit.cut' },
        { label: 'Copy', accelerator: 'Ctrl-C', command: 'edit.copy' },
        { label: 'Paste', accelerator: 'Ctrl-V', command: 'edit.paste' },
      ],
    },
    {
      label: 'Selection',
      items: [
        { label: 'Select All', accelerator: 'Ctrl-A', command: 'selection.all' },
        { label: 'Expand Selection', accelerator: 'Ctrl-Shift-Right', command: 'selection.expand' },
        { label: 'Shrink Selection', accelerator: 'Ctrl-Shift-Left', command: 'selection.shrink' },
      ],
    },
    {
      label: 'View',
      items: [
        { label: 'Toggle Left Dock', accelerator: 'Ctrl-B', command: 'view.toggleLeftDock' },
        { label: 'Toggle Right Dock', accelerator: 'Ctrl-Alt-B', command: 'view.toggleRightDock' },
        { label: 'Toggle Bottom Dock', accelerator: 'Ctrl-J', command: 'view.toggleBottomDock' },
        SEP,
        { label: 'Zoom In', accelerator: 'Ctrl-=', command: 'view.zoomIn' },
        { label: 'Zoom Out', accelerator: 'Ctrl--', command: 'view.zoomOut' },
      ],
    },
    {
      label: 'Go',
      items: [
        { label: 'Go to File…', accelerator: 'Ctrl-P', command: 'go.file' },
        { label: 'Go to Symbol…', accelerator: 'Ctrl-Shift-O', command: 'go.symbol' },
        { label: 'Go to Line…', accelerator: 'Ctrl-G', command: 'go.line' },
      ],
    },
    {
      label: 'Run',
      items: [
        { label: 'Run Task…', accelerator: 'Ctrl-Shift-R', command: 'run.task' },
        { label: 'Spawn Task', command: 'run.spawn' },
      ],
    },
    {
      label: 'Window',
      items: [
        { label: 'Minimize', accelerator: 'Ctrl-M', command: 'window.minimize' },
        { label: 'Zoom', command: 'window.zoom' },
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

  return { MENU_MODEL, SEP };
});
