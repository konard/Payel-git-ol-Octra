// Entry point for the landing-page screenshot harness (screenshot.html).
//
// It forces a deterministic environment — dark theme, English UI, Monaco wired
// to the locally installed package (no CDN) — then seeds the task store for the
// surface named in the `?surface=` query parameter and renders the real
// components. This is a DEV-only tool used by scripts/capture-landing.mjs to
// regenerate the four landing screenshots; it is not part of the app bundle.

import { createRoot } from 'react-dom/client';
import * as monaco from 'monaco-editor';
import { loader } from '@monaco-editor/react';

// Vite worker imports keep Monaco fully offline (no CDN, no network workers).
import editorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker';
import jsonWorker from 'monaco-editor/esm/vs/language/json/json.worker?worker';
import cssWorker from 'monaco-editor/esm/vs/language/css/css.worker?worker';
import htmlWorker from 'monaco-editor/esm/vs/language/html/html.worker?worker';
import tsWorker from 'monaco-editor/esm/vs/language/typescript/ts.worker?worker';

import '../styles/index.css';
import { useThemeStore } from '../stores/themeStore';
import { useI18nStore } from '../stores/i18nStore';
import { translationsCache } from '../hooks/useI18n';
import { Harness } from './Harness';
import { seedSurface, type Surface } from './fixtures';

(self as unknown as { MonacoEnvironment: unknown }).MonacoEnvironment = {
  getWorker(_workerId: string, label: string) {
    if (label === 'json') return new jsonWorker();
    if (label === 'css' || label === 'scss' || label === 'less') return new cssWorker();
    if (label === 'html' || label === 'handlebars' || label === 'razor') return new htmlWorker();
    if (label === 'typescript' || label === 'javascript') return new tsWorker();
    return new editorWorker();
  },
};

loader.config({ monaco });

const VALID_SURFACES: Surface[] = ['code-view', 'research-progress', 'document-reader', 'presentation-deck'];

function resolveSurface(): Surface {
  const requested = new URLSearchParams(window.location.search).get('surface');
  return VALID_SURFACES.includes(requested as Surface) ? (requested as Surface) : 'code-view';
}

async function boot() {
  // Force dark theme.
  useThemeStore.getState().setTheme(true);
  document.documentElement.classList.add('dark');

  // Pin the UI to English: preload the cache, persist the choice, set the store.
  try {
    const en = await fetch('/languages/en.json').then((res) => res.json());
    translationsCache.en = en;
  } catch (error) {
    console.warn('[screenshot] failed to preload English translations:', error);
  }
  try {
    localStorage.setItem('octra-language', 'en');
  } catch {
    // localStorage may be unavailable; the subscription in Harness still pins en.
  }
  useI18nStore.setState({ language: 'en' });
  document.documentElement.lang = 'en';

  const surface = resolveSurface();
  const config = seedSurface(surface);

  createRoot(document.getElementById('root')!).render(
    <Harness mode={config.mode} messages={config.messages} />,
  );
}

void boot();
