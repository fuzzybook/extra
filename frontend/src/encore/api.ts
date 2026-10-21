import Client from './client';
import { API_BASE_URL } from './base-url';

// Single shared API client. Authentication travels in the HttpOnly session
// cookie set by /auth/login, so no token is attached here.
export const api = new Client(API_BASE_URL);
