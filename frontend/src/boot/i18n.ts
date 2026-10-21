import type { App } from 'vue';
import { i18n } from '@/i18n';

export default ({ app }: { app: App<Element> }) => {
  app.use(i18n);
};
