'use strict';

// The preload bridge exposes a tiny, explicit `window.octra` API to the renderer.
// contextIsolation is on and nodeIntegration is off, so the renderer can only
// reach the main process through the channels whitelisted here — never raw Node.

const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('octra', {
  // Marks that we are running inside Electron (the renderer falls back to a mock
  // when this is absent, e.g. when previewing the chrome in a plain browser).
  isElectron: true,
  platform: process.platform,

  window: {
    minimize: () => ipcRenderer.send('window:minimize'),
    toggleMaximize: () => ipcRenderer.send('window:toggle-maximize'),
    close: () => ipcRenderer.send('window:close'),
    isMaximized: () => ipcRenderer.invoke('window:is-maximized'),
  },

  projects: {
    listRecent: () => ipcRenderer.invoke('projects:list-recent'),
    open: () => ipcRenderer.invoke('projects:open'),
    openPath: (p) => ipcRenderer.invoke('projects:open-path', p),
    create: (name) => ipcRenderer.invoke('projects:create', name),
    forget: (p) => ipcRenderer.invoke('projects:forget', p),
  },

  shell: {
    openExternal: (url) => ipcRenderer.send('shell:open-external', url),
  },

  // Native menu / accelerator commands are forwarded here so the renderer can run
  // the same handler whether the command came from the menu bar or a shortcut.
  onMenuCommand: (handler) => {
    ipcRenderer.on('menu:command', (_event, command) => handler(command));
  },
});
