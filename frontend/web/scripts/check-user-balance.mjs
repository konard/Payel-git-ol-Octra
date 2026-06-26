import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';

// Regression checks for issue #105:
// - the top bar must expose a real user balance component backed by auth data;
// - the header's "new chat" control must open a new chat instead of only opening
//   the history sidebar;
// - the auth API types must carry the balance returned by /me.

const here = dirname(fileURLToPath(import.meta.url));
const root = resolve(here, '..');
const read = (rel) => readFileSync(resolve(root, rel), 'utf8');

const topbar = read('src/app/components/shell/TopBar.tsx');
assert.match(
  topbar,
  /import \{ UserBalance \} from ['"]\.\.\/\.\.\/\.\.\/components\/user\/UserBalance['"]/,
  'TopBar must import the reusable UserBalance component',
);
assert.match(
  topbar,
  /\{isAuthenticated && <UserBalance \/>}/,
  'TopBar must render UserBalance for authenticated users',
);
assert.match(
  topbar,
  /onNewChat\?: \(\) => void/,
  'TopBar props must expose a new-chat callback',
);
assert.match(
  topbar,
  /onClick=\{\(\) => onNewChat\?\.\(\)\}/,
  'the header new-chat button must call onNewChat instead of only opening the sidebar',
);

const app = read('src/app/App.tsx');
assert.match(
  app,
  /onNewChat=\{\(\) => \{\s*void handleNewChat\(\);\s*\}\}/,
  'App must wire TopBar onNewChat to the real new chat flow',
);

const authService = read('src/services/authService.ts');
assert.match(
  authService,
  /balance_credits\?: number/,
  'auth service user types must include balance_credits from the backend',
);

const userBalance = read('src/components/user/UserBalance.tsx');
assert.match(
  userBalance,
  /useAuthStore/,
  'UserBalance must read real auth store state',
);
assert.doesNotMatch(
  userBalance,
  /const\s+\w+\s*=\s*(?:100|1000|100000)/,
  'UserBalance must not hardcode the displayed balance metric',
);

const indexCss = read('src/styles/index.css');
assert.match(
  indexCss,
  /@import '\.\/components\/user-balance\.css';/,
  'UserBalance styles must live in a component-scoped stylesheet imported by the app',
);

console.log('check-user-balance: all assertions passed');
