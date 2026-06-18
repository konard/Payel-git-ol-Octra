import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';

// Regression test for issue #31: "still can't write in chat". The chat input
// only submitted on a button click — pressing Enter merely inserted newlines, so
// the message was never sent. The input must now send on Enter (and keep
// Shift+Enter for a newline) like every other chat client.

const here = dirname(fileURLToPath(import.meta.url));
const root = resolve(here, '..');
const read = (rel) => readFileSync(resolve(root, rel), 'utf8');

const chatInput = read('src/app/components/chat/ChatInput.tsx');

// The textarea must wire up a keydown handler.
assert.match(chatInput, /onKeyDown=\{handleKeyDown\}/, 'ChatInput textarea must have an onKeyDown handler');
assert.match(chatInput, /const handleKeyDown =/, 'ChatInput must define handleKeyDown');

// Enter (without Shift) sends; Shift+Enter must NOT send.
assert.match(chatInput, /e\.key === 'Enter'/, 'handleKeyDown must react to the Enter key');
assert.match(chatInput, /!e\.shiftKey/, 'Enter must only send when Shift is not held');
assert.match(chatInput, /e\.preventDefault\(\)/, 'Enter-to-send must prevent the default newline');

// IME composition (e.g. typing CJK) must not trigger an accidental send.
assert.match(chatInput, /isComposing/, 'handleKeyDown must ignore IME composition Enter');

// Both the form submit and the Enter key must funnel through one send helper so
// behaviour stays consistent.
assert.match(chatInput, /const submitMessage =/, 'ChatInput must share a submitMessage helper');
assert.match(chatInput, /onSendMessage\(inputValue\.trim\(\)\)/, 'submit must send the trimmed value');

console.log('check-chat-input: all assertions passed');
