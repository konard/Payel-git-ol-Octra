import { useEffect, useState } from 'react';

// useMediaQuery tracks whether a CSS media query currently matches. It powers the
// responsive workspace: on wide screens we show the docked Sessions / Canvas /
// Solution panes side by side, while narrow screens fall back to a single pane.
export function useMediaQuery(query: string): boolean {
  const getMatch = () =>
    typeof window !== 'undefined' && typeof window.matchMedia === 'function'
      ? window.matchMedia(query).matches
      : false;

  const [matches, setMatches] = useState<boolean>(getMatch);

  useEffect(() => {
    if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') {
      return;
    }
    const mql = window.matchMedia(query);
    const handler = (event: MediaQueryListEvent) => setMatches(event.matches);
    // Sync immediately in case the query changed between render and effect.
    setMatches(mql.matches);
    mql.addEventListener('change', handler);
    return () => mql.removeEventListener('change', handler);
  }, [query]);

  return matches;
}

// Tailwind's `lg` breakpoint. Above it we have room for the three docked panes.
export const useIsDesktop = () => useMediaQuery('(min-width: 1024px)');
