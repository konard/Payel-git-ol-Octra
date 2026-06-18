import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';

// Regression test for issue #59: attaching files only showed a tiny count badge
// on the paperclip button. The unified task input must show the selected images
// and other files inside the composer above the text area, with a way to remove
// an accidental attachment before submitting.

const here = dirname(fileURLToPath(import.meta.url));
const root = resolve(here, '..');
const read = (rel) => readFileSync(resolve(root, rel), 'utf8');

const bottomInput = read('src/app/components/chat/BottomInput.tsx');

assert.match(bottomInput, /type AttachedFileItem =/, 'BottomInput must keep attachment metadata for rendering previews');
assert.match(bottomInput, /previewUrl/, 'image attachments must get object URLs for thumbnails');
assert.match(bottomInput, /URL\.createObjectURL/, 'image previews must be created from selected files');
assert.match(bottomInput, /URL\.revokeObjectURL/, 'image preview object URLs must be released');
assert.match(bottomInput, /attachedFiles\.map/, 'the composer must render each selected attachment');
assert.match(bottomInput, /<img/, 'image attachments must render thumbnail previews');
assert.match(bottomInput, /FileText/, 'non-image attachments must render a file icon');
assert.match(bottomInput, /removeAttachedFile/, 'users must be able to remove a selected attachment');
assert.match(bottomInput, /attachedFiles\.map\(\(attachment\) => attachment\.file\)/, 'submitted task data must still include the raw File objects');

const previewStripPosition = bottomInput.indexOf('aria-label="Attached files"');
assert.ok(previewStripPosition >= 0, 'the composer must expose an attachment preview strip');
const previewBeforeText = bottomInput.indexOf('attachedFiles.map', previewStripPosition);
const textAreaPosition = bottomInput.indexOf('<textarea');
assert.ok(previewBeforeText >= 0 && textAreaPosition >= 0 && previewBeforeText < textAreaPosition, 'attachment previews must appear above the text area');

console.log('check-bottom-input-attachments: all assertions passed');
