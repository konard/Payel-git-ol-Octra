import { useEffect, useState } from 'react';
import { getFileIconName } from './fileIcon';

const cache = new Map<string, string>();

export function FileTypeIcon({
  filePath,
  size = 15,
}: {
  filePath: string;
  size?: number;
}) {
  const iconName = getFileIconName(filePath);
  const [svgContent, setSvgContent] = useState<string | null>(
    () => cache.get(iconName) ?? null,
  );

  useEffect(() => {
    if (cache.has(iconName)) {
      setSvgContent(cache.get(iconName)!);
      return;
    }

    const url = `/node_modules/material-icon-theme/icons/${iconName}.svg`;
    let cancelled = false;

    fetch(url)
      .then((res) => {
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        return res.text();
      })
      .then((svg) => {
        if (!cancelled) {
          cache.set(iconName, svg);
          setSvgContent(svg);
        }
      })
      .catch(() => {
        if (!cancelled) setSvgContent(null);
      });

    return () => {
      cancelled = true;
    };
  }, [iconName]);

  if (!svgContent) {
    return (
      <span
        className="shrink-0 inline-flex items-center justify-center rounded"
        style={{
          width: size,
          height: size,
          background: 'var(--text-muted, #999)',
          opacity: 0.3,
        }}
      />
    );
  }

  return (
    <span
      className="shrink-0 inline-flex items-center justify-center"
      style={{ width: size, height: size }}
      dangerouslySetInnerHTML={{
        __html: svgContent.replace(
          '<svg',
          `<svg width="${size}" height="${size}"`,
        ),
      }}
    />
  );
}
