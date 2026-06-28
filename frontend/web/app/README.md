# Frontend Architecture

## Tech Stack
- **Framework:** Next 16 + TypeScript
- **UI:** React 18 + Lucide icons
- **Styling:** Centralized `globals.css` + CSS custom properties + 2 CSS modules

---

## CSS Variable System (`globals.css :root`)

### Animation (from `styles/tokens.css`)
| Token | Value |
|---|---|
| `--anim-duration` | `0.2s` |
| `--anim-ease` | `cubic-bezier(0.4, 0, 0.2, 1)` |

### Surface
| Token | Usage |
|---|---|
| `--bg` `#050609` | Page background |
| `--bg-2` `#0b0b0b` | Secondary background |
| `--panel` `rgba(10,10,10,0.94)` | Default surface |
| `--panel-2` `rgba(16,16,16,0.92)` | Elevated surface |
| `--panel-3` `rgba(26,26,26,0.84)` | Deeper surface |
| `--panel-elevated` | Gradient for highest surfaces |
| `--titlebar-bg` | Desktop title bar |
| `--rail-bg` | Left rail (navigation) |
| `--sidebar-bg` | Side panels |
| `--card-bg` | Card components |
| `--modal-bg` | Modal overlays |
| `--section-bg` | Section containers |

### Lines & Borders
| Token | Value |
|---|---|
| `--line` `rgba(255,255,255,0.11)` | Default border |
| `--line-strong` `rgba(255,255,255,0.2)` | Emphasis border |

### Text
| Token | Value |
|---|---|
| `--text` `#f5f7f2` | Primary text |
| `--muted` `#b0b0b0` | Secondary text |
| `--quiet` `#787878` | Tertiary / placeholder |

### White Opacity Scale (naming: `--white-{opacity*100}`)
`micro`(0.03), `subtle`(0.04), `02`, `035`, `045`, `05`, `055`, `06`, `mid`(0.07), `08`, `10`, `12`, `13`, `18`, `19`, `22`, `24`, `26`, `28`, `30`, `34`, `35`, `36`, `42`, `48`, `58`, `strong`(0.25)

Used for: backgrounds, borders, gradients, overlays.

### Black Opacity Scale
`--black-22` through `--black-78` — used for shadows and mask overlays.

### Metrics (status colors)
| Token | Color |
|---|---|
| `--metric-success` `#32d583` | Green |
| `--metric-warning` `#f7c948` | Yellow |
| `--metric-danger` `#ff5b66` | Red |
| `--metric-{name}-bg-{opacity}` | Tinted backgrounds for metric cards |

### Light Theme (`.theme-light`)
| Token | Value |
|---|---|
| `--light-bg` | `rgba(246,246,246,0.96)` |
| `--light-text` | `rgba(245,247,242,0.07)` |
| `--light-border` | `rgba(245,247,242,0.48)` |
| `--light-gradient-from/to` | Primary button gradient colors |
| `--light-stroke` | SVG stroke in light mode |
| `--dark-text` | Text on light backgrounds (dark) |
| `--dark-bg` / `--dark-bg-alt` / `--dark-bg-soft` | Dark surface variants in light mode |

### Shadows
| Token | Value |
|---|---|
| `--shadow` | `0 24px 80px rgba(0,0,0,0.48)` |

### Border Radius
| Token | Usage |
|---|---|
| `--radius-sm` `8px` | 40 uses — cards, inputs, modals |
| `--radius-xs` `6px` | 4 uses — small elements |

---

## Component Tree

```
layout.tsx                          # Root: globals.css + tokens + canvas bg
├── DesktopTitleBar                 # Draggable title bar
├── page.tsx                        # Landing page
├── login/page.tsx                  # Auth (Google, GitHub, Email)
├── app/page.tsx                    # Workspace (/app)
├── dashboard/                      # /dashboard
│   ├── DashboardShell              # Sidebar + topbar + tabs layout
│   ├── page.tsx                    # Overview: metrics + canvas + policy
│   └── [section]/page.tsx          # Dynamic section pages
├── settings/page.tsx               # User settings
├── profile/                        # /profile
│   └── ProfileSidebar              # Profile navigation
└── components/
    ├── IconButton                  # Reusable icon button
    ├── Select                      # Custom dropdown
    ├── EmptyDataPanel              # Empty state placeholder
    ├── CreateEnvironmentModal      # Environment creation dialog
    ├── CreateKeyModal              # API key creation dialog
    ├── UserBalance                 # Balance display
    ├── EnvironmentPanel            # Environment list (legacy)
    ├── WelcomeModal                # First-time user modal
    └── WorkflowCanvas              # Node-based workflow editor
```

## Conventions

- **"use client"** — interactive components only (no server components with client logic)
- **CSS classes** — all defined in `globals.css` as flat selectors (`.class-name`); no CSS-in-JS
- **CSS modules** — only for truly isolated component styles (`EmptyDataPanel.module.css`, `UserBalance.module.css`)
- **Icons** — Lucide React (`lucide-react`)
- **API calls** — separate `server/` directory with async functions
- **Config** — `config/` for routes, images, mock data
- **Typography** — `Inter` font, `font-weight: 800` for emphasis, rem-based sizes (most common: `0.82rem`)

## Routes

| Path | Page |
|---|---|
| `/` | Landing |
| `/auth` | Sign-in |
| `/app` | Workspace |
| `/dashboard` | Dashboard |
| `/dashboard/:section` | Dashboard section |
| `/login` | Login (legacy) |
| `/settings` | Settings |
| `/profile` | Profile |
| `/profile/api-keys` | API keys |
| `/profile/settings` | Profile settings |
