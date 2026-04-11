/**
 * Telegram Integration Service
 * 
 * This service handles Telegram bot integration for:
 * - User authentication via Telegram bot
 * - Sending notifications to Telegram
 * - Receiving tasks from Telegram
 */

const TELEGRAM_BOT_URL = 'https://t.me/CrewAIBot'; // Replace with your actual bot URL

export interface TelegramUser {
  id: number;
  first_name: string;
  username?: string;
  language_code?: string;
}

export interface TelegramAuthPayload {
  id: number;
  first_name: string;
  last_name?: string;
  username?: string;
  photo_url?: string;
  auth_date: string;
  hash: string;
}

/**
 * Get Telegram bot URL for user authorization
 * Users click this link to open the bot and start authentication
 */
export function getTelegramAuthUrl(workflowId?: string): string {
  const params = new URLSearchParams();
  
  if (workflowId) {
    params.set('workflow', workflowId);
  }
  
  // Add start parameter with auth token
  const state = generateStateToken();
  params.set('start', state);
  
  return `${TELEGRAM_BOT_URL}?${params.toString()}`;
}

/**
 * Generate a state token for Telegram auth flow
 * This is used to correlate the auth request with the user session
 */
function generateStateToken(): string {
  const array = new Uint8Array(16);
  crypto.getRandomValues(array);
  return Array.from(array, (b) => b.toString(16).padStart(2, '0')).join('');
}

/**
 * Verify Telegram authentication hash
 * According to Telegram's auth widget specification
 */
export async function verifyTelegramAuth(payload: TelegramAuthPayload, botToken: string): Promise<boolean> {
  try {
    const { hash, ...data } = payload;
    
    // Create data check string
    const dataCheckString = Object.keys(data)
      .sort()
      .map((key) => `${key}=${data[key as keyof Omit<TelegramAuthPayload, 'hash'>]}`)
      .join('\n');

    // Create secret key using SHA256 of bot token
    const encoder = new TextEncoder();
    const secretKeyBuffer = await crypto.subtle.digest(
      'SHA-256',
      encoder.encode(botToken)
    );
    const secretKey = await crypto.subtle.importKey(
      'raw',
      secretKeyBuffer,
      { name: 'HMAC', hash: 'SHA-256' },
      false,
      ['sign']
    );

    // Calculate hash
    const signatureBuffer = await crypto.subtle.sign(
      'HMAC',
      secretKey,
      encoder.encode(dataCheckString)
    );
    const signatureArray = Array.from(new Uint8Array(signatureBuffer));
    const calculatedHash = signatureArray
      .map((b) => b.toString(16).padStart(2, '0'))
      .join('');

    return calculatedHash === hash;
  } catch (error) {
    console.error('Error verifying Telegram auth:', error);
    return false;
  }
}

/**
 * Send a notification message to a Telegram user
 * 
 * @param chatId - Telegram chat ID
 * @param message - Message text
 * @param botToken - Telegram bot token
 */
export async function sendTelegramNotification(
  chatId: string,
  message: string,
  botToken: string
): Promise<boolean> {
  try {
    const response = await fetch(
      `https://api.telegram.org/bot${botToken}/sendMessage`,
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          chat_id: chatId,
          text: message,
          parse_mode: 'HTML',
        }),
      }
    );

    if (!response.ok) {
      const error = await response.json();
      console.error('Telegram API error:', error);
      return false;
    }

    return true;
  } catch (error) {
    console.error('Error sending Telegram notification:', error);
    return false;
  }
}

/**
 * Send task status update to Telegram
 */
export async function sendTaskStatusUpdate(
  chatId: string,
  taskId: string,
  status: string,
  botToken: string
): Promise<boolean> {
  const message = `
📋 <b>Task Status Update</b>

🆔 Task: ${taskId}
📊 Status: ${status}
⏰ Time: ${new Date().toLocaleString()}
  `.trim();

  return sendTelegramNotification(chatId, message, botToken);
}

/**
 * Send completed task (zip) to Telegram
 */
export async function sendTaskToTelegram(
  chatId: string,
  taskId: string,
  zipData: Blob,
  botToken: string
): Promise<boolean> {
  try {
    const formData = new FormData();
    formData.append('chat_id', chatId);
    formData.append(
      'document',
      zipData,
      `${taskId}-result.zip`
    );
    formData.append('caption', `✅ Task completed: ${taskId}`);

    const response = await fetch(
      `https://api.telegram.org/bot${botToken}/sendDocument`,
      {
        method: 'POST',
        body: formData,
      }
    );

    if (!response.ok) {
      const error = await response.json();
      console.error('Telegram send document error:', error);
      return false;
    }

    return true;
  } catch (error) {
    console.error('Error sending task to Telegram:', error);
    return false;
  }
}

/**
 * Initialize Telegram integration
 * Opens the Telegram bot URL for user authorization
 */
export function openTelegramBot(workflowId?: string): void {
  const authUrl = getTelegramAuthUrl(workflowId);
  window.open(authUrl, '_blank', 'noopener,noreferrer');
}

/**
 * Check if user came back from Telegram auth
 * This would typically check with your backend to verify auth
 */
export async function checkTelegramAuth(userId: string): Promise<boolean> {
  try {
    // Check with your backend if user has authenticated via Telegram
    // This is a placeholder - implement based on your backend setup
    const response = await fetch(`/api/auth/telegram/status/${userId}`);
    
    if (!response.ok) {
      return false;
    }

    const data = await response.json();
    return data.connected || false;
  } catch (error) {
    console.error('Error checking Telegram auth:', error);
    return false;
  }
}

/**
 * Hook for Telegram integration
 */
export function useTelegramIntegration() {
  const connectTelegram = (workflowId?: string) => {
    openTelegramBot(workflowId);
  };

  const notifyUser = async (
    chatId: string,
    message: string,
    botToken: string
  ) => {
    return sendTelegramNotification(chatId, message, botToken);
  };

  return {
    connectTelegram,
    notifyUser,
    botUrl: TELEGRAM_BOT_URL,
  };
}
