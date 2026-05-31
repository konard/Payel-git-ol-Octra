/**
 * Helpers for the Solution tab's Markdown handling.
 *
 * The Solution tab (formerly "Code") renders agent results. Research reports and
 * generated documents arrive as Markdown (`.md`) and are shown as a formatted
 * preview by default; source files keep the Monaco code view.
 */

/** Matches Markdown document extensions, case-insensitively. */
const MARKDOWN_EXTENSION = /\.(md|markdown|mdx)$/i;

/** Returns true when the given file path points to a Markdown document. */
export function isMarkdownPath(path: string): boolean {
  return MARKDOWN_EXTENSION.test(path);
}

/**
 * Matches binary documents/assets that must not be dumped into the text editor.
 * Presentations (`.pptx`), office documents and images are produced on the
 * server; their raw bytes cannot be edited or rendered as text in the browser.
 */
const BINARY_EXTENSION = /\.(pptx|docx|xlsx|pdf|png|jpe?g|gif|zip)$/i;

/** Returns true when the given file path points to a binary (non-text) file. */
export function isBinaryPath(path: string): boolean {
  return BINARY_EXTENSION.test(path);
}

/** Human-friendly label for a binary document type, used in placeholders. */
export function binaryFileLabel(path: string): string {
  const ext = path.toLowerCase().split('.').pop() ?? '';
  switch (ext) {
    case 'pptx':
      return 'PowerPoint presentation';
    case 'docx':
      return 'Word document';
    case 'xlsx':
      return 'Excel spreadsheet';
    case 'pdf':
      return 'PDF document';
    case 'png':
    case 'jpg':
    case 'jpeg':
    case 'gif':
      return 'Image';
    case 'zip':
      return 'Archive';
    default:
      return 'Binary file';
  }
}
