import { useEffect } from 'react';
import { TopBar } from '../app/components/shell/TopBar';
import { SolutionViewer } from '../app/components/solution/SolutionViewer';
import { Chat, type ChatMessage } from '../app/components/chat/Chat';
import { useI18nStore } from '../stores/i18nStore';

interface HarnessProps {
  mode: 'canvas' | 'chat' | 'solution';
  messages: ChatMessage[];
}

// Harness renders the real production chrome (TopBar) plus the requested surface
// (SolutionViewer or Chat) in a fixed-size, single-pane (isDesktop=false) layout
// so the four landing screenshots are captured from the genuine components.
export function Harness({ mode, messages }: HarnessProps) {
  // The unauthenticated i18n store kicks off a geolocation request that can flip
  // the language away from English (the cause of the old mixed-language
  // screenshots). Pin it back to English so every capture is consistent.
  useEffect(() => {
    const unsubscribe = useI18nStore.subscribe((state) => {
      if (state.language !== 'en') {
        useI18nStore.setState({ language: 'en' });
        document.documentElement.lang = 'en';
      }
    });
    return unsubscribe;
  }, []);

  return (
    <div className="flex h-screen w-screen flex-col overflow-hidden bg-[var(--background)] text-[var(--text)]">
      <TopBar
        isAuthenticated={false}
        hasSubscription={false}
        onShowAuth={() => {}}
        onShowSubscription={() => {}}
        mode={mode}
        onModeChange={() => {}}
        hasUnreadMessages={false}
        onToggleSidebar={() => {}}
        isDesktop={false}
      />
      <div className="min-h-0 flex-1">
        {mode === 'chat' ? (
          <Chat messages={messages} onMarkAsRead={() => {}} />
        ) : (
          <SolutionViewer />
        )}
      </div>
    </div>
  );
}
