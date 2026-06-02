export interface PresentationPreviewFile {
  path: string;
  content: string;
  encoding?: string;
}

export interface PresentationImage {
  url: string;
  alt: string;
}

export interface PresentationSlide {
  title: string;
  bullets: string[];
  visual: string;
  image: PresentationImage | null;
  sources: string[];
  notes: string;
}

export interface PresentationDeck {
  title: string;
  slides: PresentationSlide[];
}

const MARKDOWN_EXTENSION = /\.(md|markdown|mdx)$/i;

function stripExtension(path: string): string {
  return path.replace(/\.[^/.]+$/, '');
}

function fileName(path: string): string {
  return path.split('/').filter(Boolean).at(-1) || path;
}

export function findPresentationPreviewFile<T extends PresentationPreviewFile>(
  files: T[],
  pptxPath: string,
): T | null {
  if (!/\.pptx$/i.test(pptxPath)) return null;

  const markdownFiles = files.filter((file) =>
    MARKDOWN_EXTENSION.test(file.path) &&
    file.encoding !== 'base64' &&
    typeof file.content === 'string' &&
    file.content.trim() !== '',
  );
  if (markdownFiles.length === 0) return null;

  const pptxBase = stripExtension(pptxPath).toLowerCase();
  const exact = markdownFiles.find((file) => stripExtension(file.path).toLowerCase() === pptxBase);
  if (exact) return exact;

  const pptxName = fileName(pptxBase);
  const sameName = markdownFiles.find((file) => fileName(stripExtension(file.path)).toLowerCase() === pptxName);
  if (sameName) return sameName;

  return markdownFiles.find((file) => /final-presentation\.(md|markdown|mdx)$/i.test(file.path)) ??
    markdownFiles.find((file) => /(presentation|deck|slides)/i.test(file.path)) ??
    null;
}

function emptySlide(title = ''): PresentationSlide {
  return {
    title,
    bullets: [],
    visual: '',
    image: null,
    sources: [],
    notes: '',
  };
}

function isHttpUrl(value: string): boolean {
  return /^https?:\/\//i.test(value.trim());
}

// setSlideImage записывает реальную ссылку на картинку в слайд, оставляя visual
// как текстовое описание (alt используется как fallback, если visual пуст).
function setSlideImage(slide: PresentationSlide, url: string, alt: string): void {
  const trimmed = url.trim();
  if (!isHttpUrl(trimmed)) return;
  slide.image = { url: trimmed, alt: alt.trim() };
  if (!slide.visual && alt.trim()) slide.visual = alt.trim();
}

function splitStructuredLine(line: string): [string, string] | null {
  const index = line.indexOf(':');
  if (index < 0) return null;

  const key = line.slice(0, index).trim().toLowerCase();
  const value = line.slice(index + 1).trim();
  switch (key) {
    case 'visual':
    case 'visual direction':
    case 'image':
    case 'image idea':
    case 'illustration':
    case 'source':
    case 'sources':
      return [key, value];
    default:
      return null;
  }
}

function applyStructuredLine(slide: PresentationSlide, line: string): boolean {
  const structured = splitStructuredLine(line);
  if (!structured) return false;

  const [key, value] = structured;
  if (!value) return true;

  if (key === 'source' || key === 'sources') {
    slide.sources.push(value);
  } else if (key === 'image' || key === 'image idea' || key === 'illustration') {
    // Прямая ссылка на картинку встраивается как изображение; текстовое описание
    // (идея картинки) идёт в visual.
    if (isHttpUrl(value)) {
      setSlideImage(slide, value, slide.image?.alt ?? '');
    } else {
      slide.visual = value;
    }
  } else {
    slide.visual = value;
  }
  return true;
}

// parseMarkdownImage разбирает строку вида ![alt](url). Возвращает alt и url
// отдельно, чтобы вызывающий код мог встроить реальную картинку.
function parseMarkdownImage(line: string): { alt: string; url: string } | null {
  const match = line.match(/!\[([^\]]*)\]\(([^)]+)\)/);
  if (!match) return null;
  return { alt: match[1].trim(), url: match[2].trim() };
}

export function parsePresentationMarkdown(markdown: string): PresentationDeck {
  const deck: PresentationDeck = { title: '', slides: [] };
  let current: PresentationSlide | null = null;

  const flush = () => {
    if (!current) return;
    if (current.title || current.bullets.length > 0 || current.visual || current.image || current.sources.length > 0 || current.notes) {
      deck.slides.push(current);
    }
    current = null;
  };

  const ensureSlide = () => {
    if (!current) current = emptySlide(deck.title);
    return current;
  };

  for (const raw of (markdown || '').split('\n')) {
    const line = raw.trim();
    if (!line) continue;

    if (line.startsWith('## ')) {
      flush();
      current = emptySlide(line.slice(3).trim());
      continue;
    }

    if (line.startsWith('# ')) {
      deck.title = line.slice(2).trim();
      continue;
    }

    if (line.startsWith('- ') || line.startsWith('* ')) {
      const slide = ensureSlide();
      const value = line.slice(2).trim();
      if (!applyStructuredLine(slide, value)) {
        slide.bullets.push(value);
      }
      continue;
    }

    if (splitStructuredLine(line)) {
      applyStructuredLine(ensureSlide(), line);
      continue;
    }

    if (line.startsWith('![')) {
      const parsed = parseMarkdownImage(line);
      if (parsed) {
        const slide = ensureSlide();
        if (isHttpUrl(parsed.url)) {
          setSlideImage(slide, parsed.url, parsed.alt);
        } else if (parsed.alt) {
          slide.visual = parsed.alt;
        }
      }
      continue;
    }

    if (line.startsWith('> ')) {
      const slide = ensureSlide();
      const note = line.slice(2).trim();
      slide.notes = slide.notes ? `${slide.notes} ${note}` : note;
    }
  }

  flush();

  if (deck.slides.length === 0 && deck.title) {
    deck.slides.push(emptySlide(deck.title));
  }

  return deck;
}
