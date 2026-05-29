import { useEffect, useState, useCallback } from 'react';
import { useSettingsStore } from '../../stores/settingsStore';

/**
 * JetBrains-style ambient edge lighting.
 *
 * Renders a fixed, pointer-events-none overlay that softly glows along the
 * screen perimeter with two random accent colors. A fresh color pair is
 * picked on every mount (so it differs each time the app opens) and can be
 * re-rolled on demand. The effect is intentionally subtle (masked to a thin
 * edge frame, low opacity) so it never interferes with the content.
 */

// Curated, pleasant palette — saturated but not harsh.
const PALETTE = [
  '#7c4dff', // violet
  '#00e5ff', // cyan
  '#f97316', // orange
  '#22c55e', // green
  '#ec4899', // pink
  '#3b82f6', // blue
  '#eab308', // amber
  '#14b8a6', // teal
  '#a855f7', // purple
  '#ef4444', // red
];

interface ColorPair {
  a: string;
  b: string;
}

/** Pick two distinct random colors, optionally differing from the previous pair. */
function pickColors(prev?: ColorPair): ColorPair {
  const a = PALETTE[Math.floor(Math.random() * PALETTE.length)];
  let b = PALETTE[Math.floor(Math.random() * PALETTE.length)];
  while (b === a) {
    b = PALETTE[Math.floor(Math.random() * PALETTE.length)];
  }
  const next = { a, b };
  // Avoid repeating the exact same pair on a re-roll.
  if (prev && prev.a === next.a && prev.b === next.b) {
    return pickColors(prev);
  }
  return next;
}

export function AmbientGlow() {
  const enabled = useSettingsStore((s) => s.ambientLighting);
  const [colors, setColors] = useState<ColorPair>(() => pickColors());

  // Re-roll colors (exposed via a global event so a settings button can trigger it).
  const reroll = useCallback(() => {
    setColors((prev) => pickColors(prev));
  }, []);

  useEffect(() => {
    const handler = () => reroll();
    window.addEventListener('octra:reroll-ambient', handler);
    return () => window.removeEventListener('octra:reroll-ambient', handler);
  }, [reroll]);

  if (!enabled) return null;

  return (
    <div
      className="ambient-glow"
      aria-hidden="true"
      style={{
        ['--ambient-color-a' as any]: colors.a,
        ['--ambient-color-b' as any]: colors.b,
      }}
    />
  );
}

/** Trigger a new random ambient color pair from anywhere in the app. */
export function rerollAmbientGlow() {
  window.dispatchEvent(new Event('octra:reroll-ambient'));
}
