// The API is served from the same origin as the app (dev server proxy in
// development, reverse proxy in production; see quasar.config.ts), so the
// HttpOnly session cookie is sent automatically and never exposed to scripts.
export const API_BASE_URL = '';
