'use strict';

// Preload bridge: the ONLY surface the renderer (the real Octra web app) uses to
// reach the desktop main process. Exposed on window.octra via contextBridge with
// contextIsolation on, so the renderer keeps running as a normal web app and only
// lights up desktop features (window chrome, filesystem, recent projects) when
// window.octra.isElectron is true. The web build degrades gracefully without it.

import { contextBridge, ipcRenderer } from 'electron';
import type { FileTreeNode, ReadFileResult } from './fileSystem';
import type { Project, RememberResult } from './projects';

export interface DesktopBridgeConfig {
  apiBaseUrl: string;
  wsUrl: string;
  platform: NodeJS.Platform;
}

const api = {
  isElectron: true,
  platform: process.platform,

  // Frameless-window controls driven by the custom DesktopTitleBar.
  window: {
    minimize: (): Promise<void> => ipcRenderer.invoke('window:minimize'),
    toggleMaximize: (): Promise<boolean> => ipcRenderer.invoke('window:toggle-maximize'),
    close: (): Promise<void> => ipcRenderer.invoke('window:close'),
    isMaximized: (): Promise<boolean> => ipcRenderer.invoke('window:is-maximized'),
    onMaximizeChange: (cb: (maximized: boolean) => void): (() => void) => {
      const listener = (_e: unknown, maximized: boolean) => cb(maximized);
      ipcRenderer.on('window:maximize-change', listener);
      return () => ipcRenderer.removeListener('window:maximize-change', listener);
    },
  },

  // Recent-projects store + folder pickers.
  projects: {
    listRecent: (): Promise<Project[]> => ipcRenderer.invoke('projects:list-recent'),
    open: (): Promise<RememberResult | null> => ipcRenderer.invoke('projects:open'),
    openPath: (p: string): Promise<RememberResult> => ipcRenderer.invoke('projects:open-path', p),
    create: (): Promise<RememberResult | null> => ipcRenderer.invoke('projects:create'),
    forget: (p: string): Promise<Project[]> => ipcRenderer.invoke('projects:forget', p),
  },

  // Filesystem reads — the headline missing feature (issue #50.4).
  fs: {
    readTree: (root: string): Promise<FileTreeNode> => ipcRenderer.invoke('fs:read-tree', root),
    readFile: (root: string, rel: string): Promise<ReadFileResult> =>
      ipcRenderer.invoke('fs:read-file', root, rel),
  },

  shell: {
    openExternal: (url: string): Promise<void> => ipcRenderer.invoke('shell:open-external', url),
  },

  // Backend endpoints + platform info, so the web renderer can target the right
  // server without guessing from window.location when packaged.
  getConfig: (): Promise<DesktopBridgeConfig> => ipcRenderer.invoke('app:get-config'),

  // Native-menu / accelerator commands forwarded from the main process.
  onMenuCommand: (cb: (command: string, payload?: unknown) => void): (() => void) => {
    const listener = (_e: unknown, command: string, payload?: unknown) => cb(command, payload);
    ipcRenderer.on('menu:command', listener);
    return () => ipcRenderer.removeListener('menu:command', listener);
  },
};

export type OctraBridge = typeof api;

contextBridge.exposeInMainWorld('octra', api);
