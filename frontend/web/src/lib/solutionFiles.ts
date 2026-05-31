export function choosePreferredCodeFilePath(
  paths: string[],
  activePath: string | null,
  latestPath: string | null,
): string | null {
  const availablePaths = new Set(paths);

  if (activePath && availablePaths.has(activePath)) {
    return activePath;
  }

  if (latestPath && availablePaths.has(latestPath)) {
    return latestPath;
  }

  return paths[0] ?? null;
}
