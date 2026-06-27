import { API } from '../config/routes'

export type MeResponse = {
  data?: {
    balance?: number;
    balance_credits?: number;
  };
};

export function fetchMe(token: string): Promise<Response> {
  return fetch(API.ME, {
    headers: { Authorization: `Bearer ${token}` },
  });
}
