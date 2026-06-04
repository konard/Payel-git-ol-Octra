import assert from 'node:assert/strict';
import { createServer } from 'vite';

// The avatar descriptor generator must be deterministic per seed (so a user's
// avatar is stable across reloads) yet produce distinct sculptures for
// different seeds (so it is "always different" between users, per issue #42).

const server = await createServer({
  root: process.cwd(),
  logLevel: 'error',
  mode: 'test',
  envFile: false,
  server: { middlewareMode: true },
  optimizeDeps: { noDiscovery: true, include: [] },
  appType: 'custom',
});

try {
  const { generateAvatarDescriptor, seedFromHash } =
    await server.ssrLoadModule('/src/utils/avatar.ts');

  // Determinism: the same seed yields an identical descriptor.
  const a = generateAvatarDescriptor('0x4358f7abf391');
  const b = generateAvatarDescriptor('0x4358f7abf391');
  assert.deepEqual(a, b, 'same seed must produce an identical avatar descriptor');

  // Structure sanity: 5x5 grid, full-length arrays, heights within [0, 4].
  assert.equal(a.grid, 5, 'avatar grid must be 5');
  assert.equal(a.heights.length, 25, 'heights must cover every grid cell');
  assert.equal(a.materials.length, 25, 'materials must cover every grid cell');
  for (const h of a.heights) {
    assert.ok(Number.isInteger(h) && h >= 0 && h <= 4, `height ${h} out of range`);
  }

  // Horizontal symmetry: the sculpture mirrors left-to-right for a balanced look.
  for (let y = 0; y < 5; y++) {
    for (let x = 0; x < 5; x++) {
      assert.equal(
        a.heights[y * 5 + x],
        a.heights[y * 5 + (4 - x)],
        'avatar heights must be horizontally symmetric',
      );
    }
  }

  // Always a visible centerpiece.
  assert.ok(a.heights[2 * 5 + 2] >= 2, 'avatar must keep a raised central block');

  // Palette must expose two block materials and gradient/ring colors.
  assert.equal(a.palette.top.length, 2, 'palette must define two top-face materials');
  assert.ok(/^hsl\(/.test(a.palette.bgInner), 'background colors must be hsl strings');
  assert.ok(/^hsl\(/.test(a.palette.ring), 'ring color must be an hsl string');

  // Distinctness: different seeds should not collapse to the same sculpture.
  const seeds = ['user-1', 'alice@example.com', 'bob', 'octra', 'zoe@mail.io'];
  const signatures = new Set(
    seeds.map((s) => {
      const d = generateAvatarDescriptor(s);
      return JSON.stringify([d.heights, d.materials, d.palette]);
    }),
  );
  assert.equal(signatures.size, seeds.length, 'different seeds must look different');

  // seedFromHash is a stable 32-bit unsigned integer.
  const seed = seedFromHash('octra');
  assert.ok(Number.isInteger(seed) && seed >= 0, 'seedFromHash must return a uint32');
  assert.equal(seed, seedFromHash('octra'), 'seedFromHash must be deterministic');

  console.log('check-profile-avatar: all assertions passed');
} finally {
  await server.close();
}
