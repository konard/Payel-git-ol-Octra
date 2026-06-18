import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';

// Regression guard for PR #47 feedback: "когда нажимаю на эту кнопку уже когда
// меню открылось с целью закрыть оно закрывается и тут же открывается" — i.e.
// clicking the model pill to CLOSE the already-open model selector closed it and
// immediately reopened it.
//
// Root cause: the selector's document `mousedown` handler closed the menu when
// the click landed outside the menu. The trigger pill lives outside the menu
// element, so a click on it fired mousedown -> onClose() (menu unmounts) and
// THEN the trigger's own onClick toggled the now-closed menu back open. The fix
// teaches the outside-click handler to treat the anchor (trigger) as "inside",
// so the mousedown is ignored and only the trigger's onClick closes the menu.

const here = dirname(fileURLToPath(import.meta.url));
const root = resolve(here, '..');
const read = (rel) => readFileSync(resolve(root, rel), 'utf8');

const selector = read('src/app/components/chat/ModelSelector.tsx');

// The outside-click handler must consult the anchor ref, not just the menu, so a
// click on the trigger is not treated as an outside click.
assert.match(
  selector,
  /anchorRef\.current\s*&&\s*anchorRef\.current\.contains\(/,
  'the outside-click handler must check anchorRef.current.contains(target) so the trigger is treated as inside the menu',
);

// onClose must only fire when the click is outside BOTH the menu and the anchor.
assert.match(
  selector,
  /if\s*\(!insideMenu\s*&&\s*!insideAnchor\)\s*\{\s*onClose\(\)/,
  'onClose() must only run when the click is outside both the menu and the anchor',
);

// The effect must keep anchorRef in its dependency list so the handler always
// sees the current anchor.
assert.match(
  selector,
  /\}, \[isOpen, onClose, anchorRef\]\);/,
  'the outside-click effect must depend on anchorRef',
);

console.log('check-model-selector-toggle: all assertions passed');
