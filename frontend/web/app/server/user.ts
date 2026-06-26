export type MeResponse = {
  data?: {
    balance?: number;
    balance_credits?: number;
  };
};

export function fetchMe(token: string): Promise<Response> {
  return fetch('/me', {
    headers: { Authorization: `Bearer ${token}` },
  });
}
