import { API } from '../config/routes'

export async function login(email: string, password: string): Promise<Response> {
  return fetch(API.LOGIN, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  });
}

export async function register(username: string, email: string, password: string): Promise<Response> {
  return fetch(API.REGISTER, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, email, password }),
  });
}
