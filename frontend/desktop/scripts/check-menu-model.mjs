#!/usr/bin/env node
/*
 * Validates electron/menuModel.ts — the single menu model shared by the native
 * menu and the in-window DesktopTitleBar. Asserts the expected sections exist,
 * every leaf has a command, and commands are unique.
 */
import assert from 'node:assert/strict';

const { MENU_MODEL } = await import('../dist-electron/menuModel.js');

assert.ok(Array.isArray(MENU_MODEL) && MENU_MODEL.length > 0);

const labels = MENU_MODEL.map((s) => s.label);
for (const expected of ['Octra', 'File', 'Edit', 'View', 'Window', 'Help']) {
  assert.ok(labels.includes(expected), `menu should have a ${expected} section`);
}

const commands = new Set();
for (const section of MENU_MODEL) {
  assert.ok(Array.isArray(section.items), `${section.label} needs items`);
  for (const item of section.items) {
    if (item.type === 'separator') continue;
    assert.ok(item.label, 'menu item needs a label');
    assert.ok(item.command, `menu item "${item.label}" needs a command`);
    assert.ok(!commands.has(item.command), `duplicate command ${item.command}`);
    commands.add(item.command);
  }
}

// Core commands the renderer relies on must exist.
for (const required of ['project.open', 'project.new', 'window.close', 'view.toggleExplorer']) {
  assert.ok(commands.has(required), `missing command ${required}`);
}

console.log('check-menu-model: OK');
