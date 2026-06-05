/*
 * Renderer-side access to the desktop backend.
 *
 * Inside Electron, `window.octra` is injected by preload.js. When the same HTML
 * is opened in a plain browser (used for visual previews and the Playwright
 * screenshot test), `window.octra` is absent, so we fall back to an in-memory
 * mock that keeps the chrome fully interactive without touching the filesystem.
 */
(function (root) {
  const real = root.octra;

  // In-memory mock mirroring the main-process project store. It starts EMPTY —
  // there is no hardcoded project list. Recents only appear here once the user
  // actually opens or creates a project (matching the real Electron store).
  function createMock() {
    let recent = [];
    let counter = 0;
    const baseName = (p) => p.split(/[\\/]/).filter(Boolean).pop() || p;
    const remember = (path, name) => {
      recent = recent.filter((r) => r.path !== path);
      const entry = { name: name || baseName(path), path, openedAt: ++counter };
      recent = [entry, ...recent].slice(0, 12);
      return { project: entry, recent };
    };
    return {
      isElectron: false,
      platform: 'browser',
      window: {
        minimize() {},
        toggleMaximize() {},
        close() {},
        isMaximized: async () => false,
      },
      projects: {
        listRecent: async () => recent,
        open: async () => remember('/home/user/dev/new-folder', null),
        openPath: async (p) => remember(p, null),
        create: async (name) => remember('/home/user/dev/' + name, name),
        forget: async (p) => {
          recent = recent.filter((r) => r.path !== p);
          return recent;
        },
      },
      shell: { openExternal() {} },
      onMenuCommand() {},
    };
  }

  root.OctraApi = real && real.isElectron ? real : createMock();
})(window);
