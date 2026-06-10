import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';

// Regression test for issue #70 ("Chat redesign"): the assistant's answers must
// stream in with a typewriter effect (text "being typed") instead of popping in
// all at once, and task-completion reports come from the backend over a `chat`
// message — so the frontend must no longer synthesise its own completion text
// from chatSummary (which would double-post).

const here = dirname(fileURLToPath(import.meta.url));
const root = resolve(here, '..');
const read = (rel) => readFileSync(resolve(root, rel), 'utf8');

// --- TypewriterText reveals text incrementally. ---
const typewriter = read('src/app/components/TypewriterText.tsx');
assert.match(typewriter, /export function TypewriterText/, 'TypewriterText component must be exported');
assert.match(typewriter, /animate\s*=\s*true/, 'TypewriterText must animate by default');
assert.match(typewriter, /setInterval/, 'TypewriterText must reveal characters over time');
assert.match(typewriter, /text\.slice\(0,\s*count\)/, 'TypewriterText must show a growing slice of the text');
assert.match(typewriter, /onTick/, 'TypewriterText must expose an onTick callback so the chat can keep scrolling');
// Non-animated messages (history / user) show the full text immediately.
assert.match(typewriter, /if \(!animate\)/, 'TypewriterText must render full text immediately when animate is false');

// --- Chat renders boss messages through TypewriterText. ---
const chat = read('src/app/components/Chat.tsx');
assert.match(chat, /import \{ TypewriterText \}/, 'Chat must import TypewriterText');
assert.match(chat, /<TypewriterText/, 'Chat must render boss messages via TypewriterText');
assert.match(chat, /animate=\{message\.animate\}/, 'Chat must pass the per-message animate flag to TypewriterText');
assert.match(chat, /animate\?: boolean/, 'ChatMessage must carry an optional animate flag');

// --- App marks fresh boss replies as animated, everything else as static. ---
const app = read('src/app/App.tsx');
assert.match(
  app,
  /animate:\s*sender === 'boss' && !showProgress/,
  'incoming boss chat messages must animate (but not progress placeholders)',
);
assert.match(app, /animate:\s*false/, 'history/user/progress messages must opt out of animation');

// --- The frontend no longer posts its own completion text from chatSummary. ---
const ws = read('src/hooks/useWebSocket.ts');
// Document that completion is now reported by the backend.
assert.match(
  ws,
  /reports task completion back in the chat/,
  'useWebSocket must document that the backend reports completion via a chat message',
);
// The `success` handler must not post a chat message itself (that would
// double-post alongside the backend completion report). Isolate the success
// case and assert it never calls onChatMessage.
const successCase = ws.slice(ws.indexOf("case 'success':"), ws.indexOf("case 'chat':"));
assert.ok(successCase.length > 0, 'could not locate the success case block');
assert.ok(
  !/onChatMessage\(/.test(successCase),
  'the success handler must not post its own completion chat message (backend reports completion now)',
);

console.log('check-chat-typewriter: all assertions passed');
