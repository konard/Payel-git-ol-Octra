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

/**
 * Returns true when a solution is made entirely of documents (Markdown reports,
 * presentations, office files) rather than a code project. In that case the
 * Solution tab shows a clean, full-width reader — no file explorer or tabs —
 * because a regular user just wants to read the finished document.
 */
export function isDocumentSolution(paths: string[]): boolean {
  if (!paths || paths.length === 0) return false;
  return paths.every((path) => isMarkdownPath(path) || isBinaryPath(path));
}

/** A standalone Markdown thematic break (`---`, `***`, `___`) acts as a page break. */
const THEMATIC_BREAK = /^\s{0,3}([-*_])(?:[ \t]*\1){2,}[ \t]*$/;

/** Soft target for characters per page when a document has no explicit breaks. */
const PAGE_TARGET_CHARS = 2200;

function splitByLength(text: string, target: number): string[] {
  const blocks = text.split(/\n{2,}/);
  const pages: string[] = [];
  let current = '';
  for (const block of blocks) {
    if (!current) {
      current = block;
      continue;
    }
    if (current.length + block.length + 2 > target) {
      pages.push(current);
      current = block;
    } else {
      current += `\n\n${block}`;
    }
  }
  if (current) pages.push(current);
  return pages;
}

/**
 * Splits a Markdown document into reader pages so large documents can be paged
 * through with a bottom page switcher. Explicit thematic breaks (`---`) are
 * honoured as page boundaries; otherwise long content is paginated by length.
 * Always returns at least one page.
 */
export function paginateMarkdown(content: string): string[] {
  const text = content ?? '';
  if (!text.trim()) return [''];

  const lines = text.split('\n');
  const segments: string[] = [];
  let buffer: string[] = [];
  for (const line of lines) {
    if (THEMATIC_BREAK.test(line)) {
      segments.push(buffer.join('\n'));
      buffer = [];
    } else {
      buffer.push(line);
    }
  }
  segments.push(buffer.join('\n'));

  const pages: string[] = [];
  for (const segment of segments) {
    const trimmed = segment.trim();
    if (!trimmed) continue;
    if (segment.length <= PAGE_TARGET_CHARS * 1.5) {
      pages.push(trimmed);
    } else {
      pages.push(...splitByLength(trimmed, PAGE_TARGET_CHARS));
    }
  }

  return pages.length > 0 ? pages : [text.trim()];
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
