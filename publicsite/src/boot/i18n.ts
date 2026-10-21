import type { App } from 'vue';
import { createI18n } from 'vue-i18n';
import messages, { type MessageSchema, type SupportedLocale } from '@/i18n';

export const i18n = createI18n<[MessageSchema], SupportedLocale>({
  legacy: false,
  locale: 'it',
  fallbackLocale: 'en',
  messages,
});

export default ({ app }: { app: App<Element> }) => {
  app.use(i18n);
};
