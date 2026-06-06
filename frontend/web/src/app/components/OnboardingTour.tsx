import { useCallback, useEffect, useMemo, useState } from 'react';
import { X } from 'lucide-react';
import { useI18n } from '../../hooks/useI18n';
import { markOnboardingTourComplete } from './onboardingTourState';

type TourStep = {
  titleKey: string;
  titleFallback: string;
  bodyKey: string;
  bodyFallback: string;
  selectors: string[];
};

type TargetRect = {
  top: number;
  left: number;
  width: number;
  height: number;
};

const SPOTLIGHT_PADDING = 8;
const TOOLTIP_GAP = 16;
const TOOLTIP_WIDTH = 360;
const TOOLTIP_HEIGHT = 280;

const TOUR_STEPS: TourStep[] = [
  {
    titleKey: 'onboarding.workspace.title',
    titleFallback: 'Workspace',
    bodyKey: 'onboarding.workspace.body',
    bodyFallback:
      'This is the working area. Build a workflow on the canvas, talk with Octra Boss, and review generated results without leaving the app.',
    selectors: ['[data-tour="workflow-workspace"]', '[data-tour="canvas"]'],
  },
  {
    titleKey: 'onboarding.agents.title',
    titleFallback: 'Add agents',
    bodyKey: 'onboarding.agents.body',
    bodyFallback:
      'Use this button to add agents and workflow templates. Drag nodes onto the canvas, connect them, and describe each role.',
    selectors: ['[data-tour="add-agent"]'],
  },
  {
    titleKey: 'onboarding.taskInput.title',
    titleFallback: 'Describe the task',
    bodyKey: 'onboarding.taskInput.body',
    bodyFallback:
      'Write the goal here, choose a provider and model, attach files if needed, then send the task to start execution.',
    selectors: ['[data-tour="task-input"]'],
  },
  {
    titleKey: 'onboarding.settings.title',
    titleFallback: 'Settings',
    bodyKey: 'onboarding.settings.body',
    bodyFallback:
      'Open settings to manage API tokens, custom providers, models, language, interface visibility, and integrations.',
    selectors: ['[data-tour="settings-button"]'],
  },
  {
    titleKey: 'onboarding.chats.title',
    titleFallback: 'Chats',
    bodyKey: 'onboarding.chats.body',
    bodyFallback:
      'Create separate chats for different tasks. Each chat can keep its own conversation and saved workflow.',
    selectors: ['[data-tour="chat-sessions"]', '[data-tour="new-chat"]', '[data-tour="chat-history"]'],
  },
  {
    titleKey: 'onboarding.solution.title',
    titleFallback: 'Solution',
    bodyKey: 'onboarding.solution.body',
    bodyFallback:
      'The Solution area shows generated files, documents, search results, and pull request summaries after Octra finishes work.',
    selectors: ['[data-tour="solution-pane"]', '[data-tour="solution-tab"]'],
  },
];

function getViewportFallbackRect(): TargetRect {
  return {
    top: Math.max(24, window.innerHeight / 2 - 70),
    left: Math.max(16, window.innerWidth / 2 - 140),
    width: Math.min(280, window.innerWidth - 32),
    height: 140,
  };
}

function isVisibleElement(element: Element) {
  const style = window.getComputedStyle(element);
  const rect = element.getBoundingClientRect();

  return (
    style.display !== 'none' &&
    style.visibility !== 'hidden' &&
    rect.width > 0 &&
    rect.height > 0 &&
    rect.bottom > 0 &&
    rect.right > 0 &&
    rect.top < window.innerHeight &&
    rect.left < window.innerWidth
  );
}

function findTourTarget(selectors: string[]) {
  for (const selector of selectors) {
    const candidates = Array.from(document.querySelectorAll(selector));
    const visible = candidates.find(isVisibleElement);
    if (visible) {
      return visible as HTMLElement;
    }
  }

  return null;
}

function buildSpotlightRect(rect: TargetRect): TargetRect {
  const left = Math.max(8, rect.left - SPOTLIGHT_PADDING);
  const top = Math.max(8, rect.top - SPOTLIGHT_PADDING);
  const right = Math.min(window.innerWidth - 8, rect.left + rect.width + SPOTLIGHT_PADDING);
  const bottom = Math.min(window.innerHeight - 8, rect.top + rect.height + SPOTLIGHT_PADDING);

  return {
    top,
    left,
    width: Math.max(72, right - left),
    height: Math.max(48, bottom - top),
  };
}

function buildTooltipPosition(rect: TargetRect) {
  const width = Math.min(TOOLTIP_WIDTH, window.innerWidth - 32);
  const spaceBelow = window.innerHeight - (rect.top + rect.height);
  const shouldPlaceAbove = spaceBelow < TOOLTIP_HEIGHT && rect.top > TOOLTIP_HEIGHT;
  const unclampedTop = shouldPlaceAbove
    ? rect.top - TOOLTIP_HEIGHT - TOOLTIP_GAP
    : rect.top + rect.height + TOOLTIP_GAP;

  return {
    top: Math.min(Math.max(16, unclampedTop), Math.max(16, window.innerHeight - TOOLTIP_HEIGHT - 16)),
    left: Math.min(
      Math.max(16, rect.left + rect.width / 2 - width / 2),
      Math.max(16, window.innerWidth - width - 16),
    ),
    width,
  };
}

export function OnboardingTour({ onClose }: { onClose: () => void }) {
  const { t } = useI18n();
  const [stepIndex, setStepIndex] = useState(0);
  const [targetRect, setTargetRect] = useState<TargetRect | null>(null);
  const step = TOUR_STEPS[stepIndex];
  const isLastStep = stepIndex === TOUR_STEPS.length - 1;

  const translate = useCallback(
    (key: string, fallback: string) => {
      const value = t(key);
      return value === key ? fallback : value;
    },
    [t],
  );

  const finishTour = useCallback(() => {
    markOnboardingTourComplete();
    onClose();
  }, [onClose]);

  const advanceTour = useCallback(() => {
    if (isLastStep) {
      finishTour();
      return;
    }

    setStepIndex((current) => current + 1);
  }, [finishTour, isLastStep]);

  const updateTargetRect = useCallback((shouldScroll = false) => {
    const target = findTourTarget(step.selectors);
    if (!target) {
      setTargetRect(getViewportFallbackRect());
      return;
    }

    if (shouldScroll) {
      target.scrollIntoView({ block: 'center', inline: 'center', behavior: 'smooth' });
    }

    window.requestAnimationFrame(() => {
      const rect = target.getBoundingClientRect();
      setTargetRect({
        top: rect.top,
        left: rect.left,
        width: rect.width,
        height: rect.height,
      });
    });
  }, [step]);

  useEffect(() => {
    updateTargetRect(true);

    const onViewportChange = () => updateTargetRect();
    window.addEventListener('resize', onViewportChange);
    window.addEventListener('scroll', onViewportChange, true);

    return () => {
      window.removeEventListener('resize', onViewportChange);
      window.removeEventListener('scroll', onViewportChange, true);
    };
  }, [updateTargetRect]);

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        finishTour();
      }

      if (event.key === 'Enter') {
        advanceTour();
      }
    };

    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, [advanceTour, finishTour]);

  const spotlightRect = useMemo(
    () => buildSpotlightRect(targetRect || getViewportFallbackRect()),
    [targetRect],
  );
  const tooltipPosition = useMemo(
    () => buildTooltipPosition(spotlightRect),
    [spotlightRect],
  );

  return (
    <div className="onboarding-tour" role="dialog" aria-modal="true" aria-live="polite">
      <div
        className="onboarding-tour__spotlight"
        style={{
          top: spotlightRect.top,
          left: spotlightRect.left,
          width: spotlightRect.width,
          height: spotlightRect.height,
        }}
      />

      <section
        className="onboarding-tour__card"
        style={{
          top: tooltipPosition.top,
          left: tooltipPosition.left,
          width: tooltipPosition.width,
        }}
      >
        <div className="onboarding-tour__header">
          <div>
            <div className="onboarding-tour__eyebrow">
              {translate('onboarding.progress', 'Training')} {stepIndex + 1} / {TOUR_STEPS.length}
            </div>
            <h2>{translate(step.titleKey, step.titleFallback)}</h2>
          </div>
          <button
            type="button"
            className="onboarding-tour__icon-button"
            onClick={finishTour}
            aria-label={translate('onboarding.skip', 'Skip training')}
          >
            <X size={16} />
          </button>
        </div>

        <p>{translate(step.bodyKey, step.bodyFallback)}</p>

        <div className="onboarding-tour__footer">
          <button type="button" className="onboarding-tour__skip" onClick={finishTour}>
            {translate('onboarding.skip', 'Skip training')}
          </button>
          <button type="button" className="onboarding-tour__next" onClick={advanceTour}>
            {isLastStep
              ? translate('onboarding.finish', 'Finish')
              : translate('onboarding.ok', 'OK')}
          </button>
        </div>
      </section>
    </div>
  );
}
