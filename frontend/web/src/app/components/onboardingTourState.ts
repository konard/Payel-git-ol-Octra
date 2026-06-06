export const ONBOARDING_TOUR_STORAGE_KEY = 'octra:onboarding-tour-completed';

export function shouldShowOnboardingTour() {
  if (typeof window === 'undefined') {
    return false;
  }

  try {
    return window.localStorage.getItem(ONBOARDING_TOUR_STORAGE_KEY) !== 'true';
  } catch {
    return false;
  }
}

export function markOnboardingTourComplete() {
  if (typeof window === 'undefined') {
    return;
  }

  try {
    window.localStorage.setItem(ONBOARDING_TOUR_STORAGE_KEY, 'true');
  } catch {
    // Storage may be disabled in hardened browser contexts; closing still works.
  }
}
