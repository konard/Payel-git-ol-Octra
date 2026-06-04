// Deterministic isometric "voxel sculpture" avatar generator.
//
// Replaces the old flat identicon with a 3D-looking block structure that is
// unique per user (seeded by a hash of their id/email) yet stable across
// reloads. The descriptor generation is pure (no DOM), so it can be unit
// tested; drawing is split out so it can run against any 2D canvas context.

export interface AvatarPalette {
  /** Background radial gradient inner (center) color. */
  bgInner: string;
  /** Background radial gradient outer (edge) color. */
  bgOuter: string;
  /** Ring color drawn just inside the circular crop. */
  ring: string;
  /** Top-face color for each block "material". */
  top: string[];
  /** Left-face color for each block "material". */
  left: string[];
  /** Right-face color for each block "material". */
  right: string[];
}

export interface AvatarDescriptor {
  /** Grid size (grid x grid cells). */
  grid: number;
  /** Stacked-cube height for each cell, row-major (length grid*grid). */
  heights: number[];
  /** Material index into the palette face arrays, row-major. */
  materials: number[];
  palette: AvatarPalette;
}

const GRID = 5;
const MAX_HEIGHT = 4;

/** Small, fast, well-distributed PRNG seeded from a single 32-bit integer. */
function mulberry32(seed: number): () => number {
  let a = seed >>> 0;
  return () => {
    a = (a + 0x6d2b79f5) | 0;
    let t = Math.imul(a ^ (a >>> 15), 1 | a);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

/** Fold an arbitrary string (e.g. a hex hash) into a 32-bit seed (FNV-1a). */
export function seedFromHash(hash: string): number {
  let h = 0x811c9dc5;
  for (let i = 0; i < hash.length; i++) {
    h ^= hash.charCodeAt(i);
    h = Math.imul(h, 0x01000193);
  }
  return h >>> 0;
}

const hsl = (h: number, s: number, l: number) => `hsl(${h}, ${s}%, ${l}%)`;

function buildPalette(rng: () => number): AvatarPalette {
  const baseHue = Math.floor(rng() * 360);
  // A near-white accent material, gently tinted by a complementary hue.
  const accentHue = (baseHue + 150 + Math.floor(rng() * 60)) % 360;
  const sat = 45 + Math.floor(rng() * 25); // 45-70%

  return {
    bgInner: hsl(baseHue, Math.max(0, sat - 25), 16),
    bgOuter: hsl(baseHue, Math.max(0, sat - 30), 6),
    ring: hsl((baseHue + 20) % 360, 32, 40),
    // Material 0: the saturated base color. Material 1: light/near-white blocks.
    top: [hsl(baseHue, sat, 62), hsl(accentHue, 14, 92)],
    left: [hsl(baseHue, sat, 44), hsl(accentHue, 14, 74)],
    right: [hsl(baseHue, sat, 30), hsl(accentHue, 14, 60)],
  };
}

/**
 * Build a deterministic avatar descriptor from a hash string. The same hash
 * always yields the same sculpture; different hashes look distinct.
 */
export function generateAvatarDescriptor(hash: string): AvatarDescriptor {
  const rng = mulberry32(seedFromHash(hash));
  const palette = buildPalette(rng);

  const heights = new Array(GRID * GRID).fill(0);
  const materials = new Array(GRID * GRID).fill(0);
  const center = (GRID - 1) / 2;
  const half = Math.ceil(GRID / 2);

  // Generate the left half and mirror it horizontally so the structure reads
  // as an intentional, balanced sculpture rather than visual noise.
  for (let y = 0; y < GRID; y++) {
    for (let x = 0; x < half; x++) {
      // Pyramid bias: cells nearer the center tend to stack higher.
      const dist = Math.abs(x - center) + Math.abs(y - center);
      const bias = Math.max(0, MAX_HEIGHT - dist);
      let h = Math.round(bias * 0.5 + rng() * (MAX_HEIGHT - bias * 0.4));
      h = Math.max(0, Math.min(MAX_HEIGHT, h));
      const mat = rng() < 0.3 ? 1 : 0;

      const mx = GRID - 1 - x;
      heights[y * GRID + x] = h;
      heights[y * GRID + mx] = h;
      materials[y * GRID + x] = mat;
      materials[y * GRID + mx] = mat;
    }
  }

  // Guarantee a raised core so every avatar has a recognizable centerpiece.
  const ci = Math.floor(center) * GRID + Math.floor(center);
  if (heights[ci] < 2) heights[ci] = 2;

  return { grid: GRID, heights, materials, palette };
}

type Ctx2D = {
  fillStyle: string | CanvasGradient;
  strokeStyle: string | CanvasGradient;
  lineWidth: number;
  beginPath(): void;
  moveTo(x: number, y: number): void;
  lineTo(x: number, y: number): void;
  closePath(): void;
  fill(): void;
  stroke(): void;
  fillRect(x: number, y: number, w: number, h: number): void;
  arc(x: number, y: number, r: number, start: number, end: number): void;
  createRadialGradient(
    x0: number, y0: number, r0: number,
    x1: number, y1: number, r1: number,
  ): CanvasGradient;
};

function polygon(ctx: Ctx2D, points: number[][], fill: string) {
  ctx.beginPath();
  ctx.moveTo(points[0][0], points[0][1]);
  for (let i = 1; i < points.length; i++) ctx.lineTo(points[i][0], points[i][1]);
  ctx.closePath();
  ctx.fillStyle = fill;
  ctx.fill();
  // A faint matching outline hides sub-pixel seams between faces.
  ctx.strokeStyle = fill;
  ctx.lineWidth = 1;
  ctx.stroke();
}

/**
 * Draw the avatar described by {@link generateAvatarDescriptor} onto a square
 * 2D canvas context of the given size (in device pixels).
 */
export function drawAvatar(ctx: Ctx2D, size: number, d: AvatarDescriptor): void {
  // Dark, vignetted radial background.
  const bg = ctx.createRadialGradient(
    size * 0.42, size * 0.36, size * 0.04,
    size * 0.5, size * 0.5, size * 0.72,
  );
  bg.addColorStop(0, d.palette.bgInner);
  bg.addColorStop(1, d.palette.bgOuter);
  ctx.fillStyle = bg;
  ctx.fillRect(0, 0, size, size);

  const grid = d.grid;
  const tw = size * 0.17; // top-diamond full width
  const th = tw / 2; // top-diamond full height (2:1 isometric)
  const ch = tw * 0.62; // cube side height
  const ox = size / 2;
  // Vertical origin chosen so the bulk of the structure sits centered.
  const oy = size * 0.30;

  const project = (gx: number, gy: number, z: number): [number, number] => [
    ox + (gx - gy) * (tw / 2),
    oy + (gx + gy) * (th / 2) - z * ch,
  ];

  const drawCube = (gx: number, gy: number, z: number, mat: number) => {
    const [cx, cy] = project(gx, gy, z);
    // Top diamond.
    polygon(ctx, [
      [cx, cy],
      [cx + tw / 2, cy + th / 2],
      [cx, cy + th],
      [cx - tw / 2, cy + th / 2],
    ], d.palette.top[mat]);
    // Left face.
    polygon(ctx, [
      [cx - tw / 2, cy + th / 2],
      [cx, cy + th],
      [cx, cy + th + ch],
      [cx - tw / 2, cy + th / 2 + ch],
    ], d.palette.left[mat]);
    // Right face.
    polygon(ctx, [
      [cx, cy + th],
      [cx + tw / 2, cy + th / 2],
      [cx + tw / 2, cy + th / 2 + ch],
      [cx, cy + th + ch],
    ], d.palette.right[mat]);
  };

  // Painter's algorithm: draw back-to-front (smaller gx+gy first), and within
  // each column draw from the bottom cube up.
  const cells: { gx: number; gy: number }[] = [];
  for (let gy = 0; gy < grid; gy++) {
    for (let gx = 0; gx < grid; gx++) cells.push({ gx, gy });
  }
  cells.sort((a, b) => a.gx + a.gy - (b.gx + b.gy));

  for (const { gx, gy } of cells) {
    const idx = gy * grid + gx;
    const h = d.heights[idx];
    const mat = d.materials[idx];
    for (let z = 0; z < h; z++) drawCube(gx, gy, z, mat);
  }

  // Subtle ring just inside the circular crop, echoing the reference design.
  ctx.beginPath();
  ctx.arc(size / 2, size / 2, size / 2 - size * 0.025, 0, Math.PI * 2);
  ctx.strokeStyle = d.palette.ring;
  ctx.lineWidth = size * 0.03;
  ctx.stroke();
}

/**
 * Produce a PNG data URL for the avatar described by `hash`. Requires a DOM
 * (uses an offscreen <canvas>); returns an empty string if unavailable.
 */
export function createAvatarDataUrl(hash: string, size = 200): string {
  if (typeof document === 'undefined') return '';
  const canvas = document.createElement('canvas');
  canvas.width = size;
  canvas.height = size;
  const ctx = canvas.getContext('2d');
  if (!ctx) return '';
  drawAvatar(ctx as unknown as Ctx2D, size, generateAvatarDescriptor(hash));
  return canvas.toDataURL();
}
