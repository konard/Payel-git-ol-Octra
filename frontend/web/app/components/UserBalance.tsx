'use client';

import { useEffect, useState } from 'react';
import { Wallet } from 'lucide-react';
import styles from './UserBalance.module.css';
import { fetchMe } from '../server/user';
import type { MeResponse } from '../server/user';
import { ROUTES } from '../config/routes';

const accessTokenKeys = ['octra_access_token', 'access_token'];
const refreshTokenKeys = ['octra_refresh_token', 'refresh_token'];

type BalanceState =
  | { status: 'loading' }
  | { status: 'signed-out' }
  | { status: 'ready'; balance: number }
  | { status: 'error' };

function readStoredToken(): string | null {
  for (const key of accessTokenKeys) {
    const token = window.localStorage.getItem(key);
    if (token) return token;
  }
  return null;
}

function storeTokenPair(accessToken: string, refreshToken: string | null): void {
  for (const key of accessTokenKeys) {
    window.localStorage.setItem(key, accessToken);
  }
  if (!refreshToken) return;
  for (const key of refreshTokenKeys) {
    window.localStorage.setItem(key, refreshToken);
  }
}

function consumeTokenFromURL(): string | null {
  const params = new URLSearchParams(window.location.search);
  const accessToken = params.get('token');
  const refreshToken = params.get('refresh_token');
  if (!accessToken) return null;

  storeTokenPair(accessToken, refreshToken);
  params.delete('token');
  params.delete('refresh_token');
  const nextSearch = params.toString();
  const nextURL = `${window.location.pathname}${nextSearch ? `?${nextSearch}` : ''}${window.location.hash}`;
  window.history.replaceState({}, '', nextURL);
  return accessToken;
}

function balanceFromResponse(payload: MeResponse): number | null {
  const balance = payload.data?.balance ?? payload.data?.balance_credits;
  return typeof balance === 'number' ? balance : null;
}

export function UserBalance() {
  const [state, setState] = useState<BalanceState>({ status: 'loading' });

  useEffect(() => {
    const token = consumeTokenFromURL() ?? readStoredToken();
    if (!token) {
      setState({ status: 'signed-out' });
      return;
    }

    let cancelled = false;
    fetchMe(token)
      .then(async (response) => {
        if (!response.ok) {
          throw new Error(`balance request failed with ${response.status}`);
        }
        return response.json() as Promise<MeResponse>;
      })
      .then((payload) => {
        if (cancelled) return;
        const user = payload.data;
        const userId = user?.user_id ?? user?.id;
        if (userId) window.localStorage.setItem('octra_user_id', userId);
        if (user?.username) window.localStorage.setItem('octra_username', user.username);
        const balance = balanceFromResponse(payload);
        setState(typeof balance === 'number' ? { status: 'ready', balance } : { status: 'error' });
      })
      .catch(() => {
        if (!cancelled) setState({ status: 'error' });
      });

    return () => {
      cancelled = true;
    };
  }, []);

  const value =
    state.status === 'ready'
      ? `${state.balance.toLocaleString()} credits`
      : state.status === 'signed-out'
        ? 'Sign in'
        : state.status === 'error'
          ? 'Unavailable'
          : 'Loading';

  return (
    <a className={styles.balance} href={state.status === 'signed-out' ? ROUTES.LOGIN : ROUTES.DASHBOARD} title="Account balance">
      <span className={styles.icon} aria-hidden="true">
        <Wallet size={16} />
      </span>
      <span className={styles.copy}>
        <span className={styles.label}>Balance</span>
        <span className={styles.value}>{value}</span>
      </span>
    </a>
  );
}
