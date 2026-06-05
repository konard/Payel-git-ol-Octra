'use strict';

// Builds the native application menu from the shared MENU_MODEL. Every leaf maps
// to a `command` string that is forwarded to the renderer over the menu:command
// channel, so the native menu and the in-window DesktopTitleBar menu trigger the
// exact same handlers.

import { Menu, BrowserWindow } from 'electron';
import type { MenuItemConstructorOptions } from 'electron';
import { MENU_MODEL } from './menuModel';

export type CommandDispatcher = (command: string) => void;

function toTemplate(dispatch: CommandDispatcher): MenuItemConstructorOptions[] {
  return MENU_MODEL.map((section) => ({
    label: section.label,
    submenu: section.items.map((item): MenuItemConstructorOptions => {
      if (item.type === 'separator') return { type: 'separator' };
      return {
        label: item.label,
        accelerator: item.accelerator,
        click: () => {
          if (item.command) dispatch(item.command);
        },
      };
    }),
  }));
}

// Build and install the application menu, routing every command to `window`'s
// renderer via the menu:command channel.
export function installApplicationMenu(window: BrowserWindow): void {
  const dispatch: CommandDispatcher = (command) => {
    if (!window.isDestroyed()) window.webContents.send('menu:command', command);
  };
  const menu = Menu.buildFromTemplate(toTemplate(dispatch));
  Menu.setApplicationMenu(menu);
}
