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
