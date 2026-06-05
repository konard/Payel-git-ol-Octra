'use strict';

// Electron main process for Octra Desktop.
//
// The renderer is the REAL Octra web app (frontend/web): in dev we load the Vite
// dev server, in prod we serve the built dist over a loopback http server. On top
// of that we add a frameless window with a custom title bar, native menu, recent-
// projects store and filesystem bridge. This is what makes the desktop IDE look
// and behave like the web app (issue #50) instead of a separate vanilla shell.

import { app, BrowserWindow, ipcMain, dialog, shell } from 'electron';
import * as path from 'path';
import { resolveConfig } from './config';
import { startStaticServer, type RunningServer } from './staticServer';
import { installApplicationMenu } from './menu';
import { readProjectTree, readProjectFile } from './fileSystem';
import {
  readRecent,
  rememberProject,
  forgetProject,
  describeProject,
  type RememberResult,
} from './projects';

const config = resolveConfig();

let mainWindow: BrowserWindow | null = null;
let staticServer: RunningServer | null = null;

function userDataDir(): string {
  return app.getPath('userData');
}

async function resolveRendererUrl(): Promise<string> {
  if (config.devServerUrl) return config.devServerUrl;
  staticServer = await startStaticServer(config.webDistDir);
  return staticServer.url;
}

async function createWindow(): Promise<void> {
  mainWindow = new BrowserWindow({
    width: 1280,
    height: 800,
    minWidth: 880,
    minHeight: 560,
    show: false,
    backgroundColor: '#0b0d12',
    frame: false,
    titleBarStyle: 'hidden',
    // Keep the OS traffic-light buttons on macOS, inset to align with our bar.
    trafficLightPosition: { x: 14, y: 12 },
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      contextIsolation: true,
      nodeIntegration: false,
      sandbox: false,
    },
  });

  // Reflect maximize/unmaximize so the title bar can swap its restore/maximize
  // icon to match the real window state.
  const emitMaximize = () => {
    if (mainWindow && !mainWindow.isDestroyed()) {
      mainWindow.webContents.send('window:maximize-change', mainWindow.isMaximized());
    }
  };
  mainWindow.on('maximize', emitMaximize);
  mainWindow.on('unmaximize', emitMaximize);

  mainWindow.once('ready-to-show', () => {
    mainWindow?.show();
    if (config.openDevtools) mainWindow?.webContents.openDevTools({ mode: 'detach' });
  });

  mainWindow.on('closed', () => {
    mainWindow = null;
  });

  installApplicationMenu(mainWindow);

  const url = await resolveRendererUrl();
  await mainWindow.loadURL(url);
}

// ---------------------------------------------------------------------------
// IPC handlers — the implementation behind the preload bridge.
// ---------------------------------------------------------------------------

function registerIpc(): void {
  ipcMain.handle('app:get-config', () => ({
    apiBaseUrl: config.apiBaseUrl,
    wsUrl: config.wsUrl,
    platform: process.platform,
  }));

  ipcMain.handle('window:minimize', () => mainWindow?.minimize());
  ipcMain.handle('window:toggle-maximize', () => {
    if (!mainWindow) return false;
    if (mainWindow.isMaximized()) mainWindow.unmaximize();
    else mainWindow.maximize();
    return mainWindow.isMaximized();
  });
  ipcMain.handle('window:close', () => mainWindow?.close());
  ipcMain.handle('window:is-maximized', () => mainWindow?.isMaximized() ?? false);

  ipcMain.handle('projects:list-recent', () => readRecent(userDataDir()));

  ipcMain.handle('projects:open', async (): Promise<RememberResult | null> => {
    if (!mainWindow) return null;
    const result = await dialog.showOpenDialog(mainWindow, {
      title: 'Open Project Folder',
      properties: ['openDirectory'],
    });
    if (result.canceled || result.filePaths.length === 0) return null;
    return rememberProject(userDataDir(), result.filePaths[0], Date.now());
  });

  ipcMain.handle('projects:open-path', (_e, projectPath: string): RememberResult =>
    rememberProject(userDataDir(), projectPath, Date.now()),
  );

  ipcMain.handle('projects:create', async (): Promise<RememberResult | null> => {
    if (!mainWindow) return null;
    const result = await dialog.showOpenDialog(mainWindow, {
      title: 'Choose a Folder for the New Project',
      properties: ['openDirectory', 'createDirectory'],
    });
    if (result.canceled || result.filePaths.length === 0) return null;
    return rememberProject(userDataDir(), result.filePaths[0], Date.now());
  });

  ipcMain.handle('projects:forget', (_e, projectPath: string) =>
    forgetProject(userDataDir(), projectPath),
  );

  // Filesystem reads (issue #50.4).
  ipcMain.handle('fs:read-tree', (_e, root: string) => readProjectTree(root));
  ipcMain.handle('fs:read-file', (_e, root: string, rel: string) => readProjectFile(root, rel));

  ipcMain.handle('shell:open-external', (_e, url: string) => shell.openExternal(url));
}

// describeProject is part of the public projects API; reference it so tree-shaking
// linters don't flag the import while keeping the module surface intact.
void describeProject;

app.whenReady().then(() => {
  registerIpc();
  void createWindow();

  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) void createWindow();
  });
});

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') app.quit();
});

app.on('before-quit', () => {
  void staticServer?.close();
});
