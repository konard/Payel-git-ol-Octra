/**
 * Detects if the app is running in Electron desktop environment
 */
export function isElectron(): boolean {
  // Renderer process
  if (typeof window !== 'undefined' && typeof window.process === 'object' && (window.process as any).type === 'renderer') {
    return true;
  }

  // Main process
  if (typeof process !== 'undefined' && typeof process.versions === 'object' && !!(process.versions as any).electron) {
    return true;
  }

  // Detect the user agent when the `nodeIntegration` option is set to true
  if (
    typeof navigator === 'object' &&
    typeof navigator.userAgent === 'string' &&
    navigator.userAgent.indexOf('Electron') >= 0
  ) {
    return true;
  }

  return false;
}

/**
 * Platform info available only in desktop mode
 */
export interface DesktopInfo {
  isDesktop: boolean;
  platform?: string;
  appVersion?: string;
}

export async function getDesktopInfo(): Promise<DesktopInfo> {
  if (!isElectron()) {
    return { isDesktop: false };
  }

  const api = (window as any).electronAPI;
  if (!api) {
    return { isDesktop: false };
  }

  return {
    isDesktop: true,
    platform: await api.getPlatform(),
    appVersion: await api.getAppVersion(),
  };
}
