'use strict';

// Electron main process for Octra Desktop. It opens a frameless window (so we can
// draw the Zed-style top panel ourselves — reference 2) and wires the renderer to
// the filesystem through a small, explicit IPC surface defined in preload.js.

const { app, BrowserWindow, ipcMain, dialog, Menu, shell } = require('electron');
const path = require('path');
const projects = require('./projects');
const { toElectronTemplate } = require('./menu');

let mainWindow = null;

function userDataDir() {
  return app.getPath('userData');
}

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1280,
    height: 800,
    minWidth: 880,
    minHeight: 560,
    backgroundColor: '#111111', // Octra dark --background, avoids white flash on load
    show: false,
    // Frameless: the renderer draws the title bar (window controls + menu).
    frame: false,
    titleBarStyle: 'hidden',
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      contextIsolation: true,
      nodeIntegration: false,
      sandbox: false,
    },
  });

  // The native menu stays available for accelerators / macOS, sharing the model
  // with the in-window menu so the two never diverge.
  Menu.setApplicationMenu(
    Menu.buildFromTemplate(toElectronTemplate((command) => {
      if (command === 'app.quit') app.quit();
      else if (mainWindow) mainWindow.webContents.send('menu:command', command);
    })),
  );

  mainWindow.loadFile(path.join(__dirname, '..', 'renderer', 'index.html'));
  mainWindow.once('ready-to-show', () => mainWindow.show());
  mainWindow.on('closed', () => {
    mainWindow = null;
  });
}

// ---- IPC: window controls (reference 2 — minimize / maximize / close) ----------

ipcMain.on('window:minimize', () => mainWindow && mainWindow.minimize());
ipcMain.on('window:toggle-maximize', () => {
  if (!mainWindow) return;
  if (mainWindow.isMaximized()) mainWindow.unmaximize();
  else mainWindow.maximize();
});
ipcMain.on('window:close', () => mainWindow && mainWindow.close());
ipcMain.handle('window:is-maximized', () => !!(mainWindow && mainWindow.isMaximized()));

// ---- IPC: projects -------------------------------------------------------------

ipcMain.handle('projects:list-recent', () => projects.readRecent(userDataDir()));

ipcMain.handle('projects:open', async () => {
  const result = await dialog.showOpenDialog(mainWindow, {
    title: 'Open Project Folder',
    properties: ['openDirectory', 'createDirectory'],
  });
  if (result.canceled || result.filePaths.length === 0) return null;
  const { project, recent } = projects.rememberProject(
    userDataDir(),
    result.filePaths[0],
    Date.now(),
  );
  return { project, recent };
});

ipcMain.handle('projects:create', async (_event, name) => {
  const result = await dialog.showOpenDialog(mainWindow, {
    title: 'Choose Location for New Project',
    properties: ['openDirectory', 'createDirectory'],
  });
  if (result.canceled || result.filePaths.length === 0) return null;
  const project = projects.createProject(result.filePaths[0], name);
  const remembered = projects.rememberProject(userDataDir(), project.path, Date.now());
  return remembered;
});

// Re-open a project already present in the recent list (clicking it in the switcher).
ipcMain.handle('projects:open-path', (_event, projectPath) => {
  return projects.rememberProject(userDataDir(), projectPath, Date.now());
});

ipcMain.handle('projects:forget', (_event, projectPath) => {
  return projects.forgetProject(userDataDir(), projectPath);
});

ipcMain.on('shell:open-external', (_event, url) => {
  if (typeof url === 'string' && /^https?:\/\//.test(url)) shell.openExternal(url);
});

// ---- App lifecycle -------------------------------------------------------------

app.whenReady().then(() => {
  createWindow();
  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) createWindow();
  });
});

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') app.quit();
});
