import { useEffect, useRef, useCallback } from 'react';
import { useIntegrationStore } from '../stores/integrationStore';
import {
  initializeLefineIntegration,
  sendTaskToInbox,
  type TaskFromLefine,
} from '../services/lefineProService';

/**
 * React hook for Lefine.pro ActivityPub integration
 * 
 * Usage:
 * ```tsx
 * const { isInitialized, sendCompletedTask } = useLefinePro();
 * ```
 */
export function useLefinePro() {
  const lefineIntegration = useIntegrationStore((state) => state.integrations.lefine);
  const unsubscribeRef = useRef<(() => void) | null>(null);

  // Initialize subscription when integration is connected
  useEffect(() => {
    if (!lefineIntegration.connected) {
      // Cleanup if disconnected
      if (unsubscribeRef.current) {
        unsubscribeRef.current();
        unsubscribeRef.current = null;
      }
      return;
    }

    const { config } = lefineIntegration;
    const workflowId = config.workflowId;
    const apiKey = config.useDefaultKey ? '' : config.apiKey;
    const outboxUrl = config.outboxEndpoint || 'https://exchange.lefine.pro/outbox';
    const inboxUrl = config.inboxEndpoint || 'https://exchange.lefine.pro/inbox';

    if (!workflowId) {
      console.warn('Lefine.pro: No workflow selected');
      return;
    }

    // Initialize integration
    const { unsubscribe } = initializeLefineIntegration(
      workflowId,
      apiKey,
      outboxUrl,
      inboxUrl
    );

    unsubscribeRef.current = unsubscribe;

    // Cleanup on unmount or disconnect
    return () => {
      if (unsubscribeRef.current) {
        unsubscribeRef.current();
        unsubscribeRef.current = null;
      }
    };
  }, [lefineIntegration.connected, lefineIntegration.config]);

  // Send completed task to Lefine.pro inbox
  const sendCompletedTask = useCallback(
    async (taskResult: {
      taskId: string;
      zipUrl: string;
      zipData?: Blob;
      status: string;
      metadata?: Record<string, any>;
    }) => {
      if (!lefineIntegration.connected) {
        console.warn('Lefine.pro: Not connected');
        return false;
      }

      const { config } = lefineIntegration;
      const apiKey = config.useDefaultKey ? '' : config.apiKey;
      const inboxUrl = config.inboxEndpoint || 'https://exchange.lefine.pro/inbox';

      return await sendTaskToInbox(taskResult, inboxUrl, apiKey);
    },
    [lefineIntegration.connected, lefineIntegration.config]
  );

  return {
    isInitialized: lefineIntegration.connected,
    sendCompletedTask,
    config: lefineIntegration.config,
  };
}

export { sendTaskToInbox, type TaskFromLefine } from '../services/lefineProService';
