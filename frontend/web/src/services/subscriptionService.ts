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
