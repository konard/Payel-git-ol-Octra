export function choosePreferredCodeFilePath(
  paths: string[],
  activePath: string | null,
  latestPath: string | null,
): string | null {
  const availablePaths = new Set(paths);

  if (activePath && availablePaths.has(activePath)) {
    return activePath;
  }

  const presentationPath = paths.find((path) => /\.pptx$/i.test(path));
  if (latestPath && availablePaths.has(latestPath)) {
    if (/final-presentation\.md$/i.test(latestPath) && presentationPath) {
      return presentationPath;
    }
    return latestPath;
  }

  if (presentationPath && paths.some((path) => /final-presentation\.md$/i.test(path))) {
    return presentationPath;
  }

  return paths[0] ?? null;
}
