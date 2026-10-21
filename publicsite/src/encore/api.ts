import Client from './client';
import { API_BASE_URL } from './base-url';

// Single shared API client for the public site (no authentication needed).
export const api = new Client(API_BASE_URL);
