/**
 * Subscription API Service
 */

const AUTH_API_URL = import.meta.env.VITE_AUTH_URL || '/auth';

export interface SubscriptionPlan {
  id: string;
  name: string;
  duration: number;
  price: string;
}

export interface SubscriptionStatus {
  has_subscription: boolean;
  subscription_end?: string;
}

export async function getSubscriptionPlans(): Promise<SubscriptionPlan[]> {
  try {
    const response = await fetch(`${AUTH_API_URL}/plans`, {
      method: 'GET',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error('Failed to fetch plans');
    }

    const data = await response.json();
    return data.data;
  } catch (error) {
    // Fallback to default plans if API is not available
    console.warn('API not available, using default plans:', error);
    return [
      { id: 'monthly', name: '1 месяц', duration: 30, price: '100000' },
      { id: 'quarterly', name: '3 месяца', duration: 90, price: '250000' },
      { id: 'semi_annually', name: '6 месяцев', duration: 180, price: '500000' },
      { id: 'annually', name: '1 год', duration: 365, price: '1000000' },
    ];
  }
}

export async function subscribeUser(userId: string, plan: string): Promise<void> {
  const response = await fetch(`${AUTH_API_URL}/subscribe`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ user_id: userId, plan }),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to subscribe');
  }
}

export async function activatePromoCode(userId: string, code: string): Promise<void> {
  const response = await fetch(`${AUTH_API_URL}/subscribe/promo`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ user_id: userId, code }),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to activate promo code');
  }
}

export interface PaymentSession {
  payment_id: string;
  confirmation_url: string;
  amount: {
    value: string;
    currency: string;
  };
}

export async function createPaymentSession(planId: string, returnUrl: string): Promise<PaymentSession> {
  const response = await fetch(`${AUTH_API_URL}/payments/create`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ plan_id: planId, return_url: returnUrl }),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to create payment session');
  }

  const data = await response.json();
  return data.data;
}

export async function getSubscriptionStatus(userId: string): Promise<SubscriptionStatus> {
  const response = await fetch(`${AUTH_API_URL}/subscription/status?user_id=${userId}`, {
    method: 'GET',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
  });

  if (!response.ok) {
    throw new Error('Failed to fetch subscription status');
  }

  const data = await response.json();
  return data.data;
}

// Test function to simulate successful payment (for development)
export async function simulatePaymentSuccess(userId: string, planId: string): Promise<void> {
  const response = await fetch(`${AUTH_API_URL}/payments/simulate-success`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ user_id: userId, plan_id: planId }),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to simulate payment');
  }
}
