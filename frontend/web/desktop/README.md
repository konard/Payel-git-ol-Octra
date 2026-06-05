# Octra Desktop

A desktop IDE shell for **Octra** built with **Electron**. It visually mirrors the
Octra web app but adapts the UI for the desktop with a **Zed-style top panel**, an
**application menu**, a **project switcher**, and native filesystem integration for
opening and creating projects.

> Implements [issue #48](https://github.com/Payel-git-ol/Octra/issues/48).

## What it looks like

| Reference | Implemented here |
| --- | --- |
| Zed top bar (project · branch) | Frameless title bar with the `octra · main · master` breadcrumb, avatar, and window controls |
| "Open Application Menu" | Hamburger button + full **Octra / File / Edit / Selection / View / Go / Run / Window / Help** menu bar |
| Project switcher | Search box, **This Window** / **Recent Projects** sections, **Open Local Folder** / **New Project** |

The chrome reuses the Octra design tokens (`--background`, `--surface`, `--accent`
`#f97316`, `--text`, `--border`, the `Inter` font) from
[`frontend/web/src/styles/theme.css`](../src/styles/theme.css), so it reads as the
same product as the web app.

## Architecture

```
src/
  main/
    main.js       Electron main process — frameless window + IPC
    preload.js    contextBridge → window.octra (the only renderer ↔ Node surface)
    menu.js       builds the native menu from the shared model
    projects.js   filesystem integration (open / create / recent), pure Node
  renderer/
    index.html    the desktop chrome
    styles/       tokens.css (Octra design system) + chrome.css
    js/
      menuModel.js     single source of truth for the menu (shared with main)
      api.js           window.octra in Electron, in-memory mock in a browser
      dropdown.js      shared open/close controller
      appMenu.js       renders the menu bar + dropdowns (reference 3)
      projectSwitcher.js  renders the project switcher (reference 4)
      app.js           boots the chrome and routes commands
```

Security: the window runs with `contextIsolation: true` and `nodeIntegration:
false`. The renderer reaches the filesystem only through the explicit, whitelisted
channels in `preload.js`.

## Develop

```bash
cd frontend/web/desktop
npm install
npm start          # launch the Electron app
```

The renderer is plain HTML/CSS/JS, so it also runs in a normal browser for quick
visual iteration (it falls back to an in-memory project mock when `window.octra`
is absent):

```bash
# from this directory
npx serve src/renderer   # or open src/renderer/index.html directly
```

## Test

Structural and functional checks (no Electron runtime required):

```bash
npm test
```

- `check-projects` — round-trips open / remember / list / create / forget on a temp dir
- `check-menu-model` — the menu bar matches reference 3 and native+in-window share one model
- `check-topbar` — the Zed-style title bar has all required regions and Octra tokens
- `check-project-switcher` — the switcher matches reference 4 and is wired to the bridge
- `check-electron-main` — frameless, context-isolated window with the expected IPC surface
