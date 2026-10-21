import { defineStore } from 'pinia';
import { ref } from 'vue';

export const useLayoutStore = defineStore('layout', () => {
  const bodyWidth = ref(0);
  const bodyHeight = ref(0);

  let bodyResizeObserver: ResizeObserver | null = null;
  let listeners = 0;

  function updateBodySize() {
    if (typeof document === 'undefined') {
      return;
    }

    const rect = document.body.getBoundingClientRect();
    bodyWidth.value = Math.floor(rect.width);
    bodyHeight.value = Math.floor(rect.height);
  }

  function startBodyResizeListener() {
    listeners += 1;

    if (typeof window === 'undefined' || typeof document === 'undefined') {
      return;
    }

    updateBodySize();

    if (bodyResizeObserver) {
      return;
    }

    bodyResizeObserver = new ResizeObserver(() => {
      updateBodySize();
    });
    bodyResizeObserver.observe(document.body);
    window.addEventListener('resize', updateBodySize);
  }

  function stopBodyResizeListener() {
    listeners = Math.max(0, listeners - 1);

    if (listeners > 0) {
      return;
    }

    bodyResizeObserver?.disconnect();
    bodyResizeObserver = null;

    if (typeof window !== 'undefined') {
      window.removeEventListener('resize', updateBodySize);
    }
  }

  return {
    bodyWidth,
    bodyHeight,
    startBodyResizeListener,
    stopBodyResizeListener,
  };
});
