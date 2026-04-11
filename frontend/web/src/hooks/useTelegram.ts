import { useCallback } from 'react';
import { useIntegrationStore } from '../stores/integrationStore';
import {
  openTelegramBot,
  sendTaskStatusUpdate,
  sendTaskToTelegram,
} from '../services/telegramService';

/**
 * React hook for Telegram integration
 * 
 * Usage:
 * ```tsx
 * const { isConnected, connect, notifyUser, sendTask } = useTelegram();
 * ```
 */
export function useTelegram() {
  const telegramIntegration = useIntegrationStore((state) => state.integrations.telegram);

  // Connect to Telegram bot
  const connect = useCallback(() => {
    openTelegramBot(telegramIntegration.config.workflowId);
  }, [telegramIntegration.config.workflowId]);

  // Notify user via Telegram
  const notifyUser = useCallback(
    async (chatId: string, message: string, botToken: string) => {
      if (!telegramIntegration.connected) {
        console.warn('Telegram: Not connected');
        return false;
      }

      return sendTaskStatusUpdate(chatId, message, 'info', botToken);
    },
    [telegramIntegration.connected]
  );

  // Send completed task to Telegram
  const sendTask = useCallback(
    async (chatId: string, taskId: string, zipData: Blob, botToken: string) => {
      if (!telegramIntegration.connected) {
        console.warn('Telegram: Not connected');
        return false;
      }

      return sendTaskToTelegram(chatId, taskId, zipData, botToken);
    },
    [telegramIntegration.connected]
  );

  return {
    isConnected: telegramIntegration.connected,
    connect,
    notifyUser,
    sendTask,
    config: telegramIntegration.config,
  };
}
