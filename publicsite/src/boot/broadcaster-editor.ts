import type { App } from 'vue';
import { createBroadcasterEditor } from '@extra/broadcaster-editor';
import { api } from '@/encore/api';

// Dialoghi del profilo broadcaster (packages/broadcaster-editor) con il client del sito.
export default ({ app }: { app: App<Element> }) => {
  app.use(createBroadcasterEditor(api));
};
