<template>
  <q-page class="br-page">
    <p v-if="loading" class="br-status">{{ t('broadcaster.loading') }}</p>

    <div v-else-if="!broadcaster" class="br-status">
      <p>{{ t('broadcaster.notFound') }}</p>
      <router-link :to="homeRoute" class="br-link">{{ t('broadcaster.backHome') }}</router-link>
    </div>

    <main v-else class="br-layout">
      <!-- Galleria: foto principale con frecce e miniature -->
      <aside :aria-label="t('broadcaster.photoGallery')">
        <div class="br-hero">
          <img
            v-if="currentPhoto"
            :key="currentPhoto.image_url"
            :src="currentPhoto.image_url"
            :alt="currentPhoto.alt || broadcaster.display_name"
            class="br-hero__img"
          />
          <div v-else class="br-placeholder br-hero__img">{{ t('broadcaster.noPhotos') }}</div>
          <template v-if="photos.length > 1">
            <button class="br-nav br-nav--prev" type="button" :aria-label="t('broadcaster.prevPhoto')" @click="show(current - 1)">‹</button>
            <button class="br-nav br-nav--next" type="button" :aria-label="t('broadcaster.nextPhoto')" @click="show(current + 1)">›</button>
          </template>
          <LiveStatusBadge :short-id="broadcaster.short_id" />
        </div>

        <div v-if="photos.length > 1" class="br-thumbs">
          <button
            v-for="(photo, i) in photos"
            :key="photo.image_url"
            type="button"
            class="br-thumb"
            :class="{ 'br-thumb--active': i === current }"
            :aria-label="t('broadcaster.showPhoto', { n: i + 1 })"
            :aria-pressed="i === current"
            @click="show(i)"
          >
            <img :src="photo.image_url" :alt="''" loading="lazy" />
          </button>
        </div>
      </aside>

      <article>
        <header class="br-header">
          <h1>{{ broadcaster.display_name }}</h1>
          <div class="br-subtitle">
            {{ broadcaster.category }}
            <span v-if="broadcaster.certified" class="br-certified">✓ {{ t('broadcaster.certified') }}</span>
          </div>
          <router-link v-if="isLive" :to="liveRoute" class="br-live">
            <span class="br-live__dot">●</span> {{ t('live.liveNow') }} · {{ t('live.watch') }}
          </router-link>
        </header>

        <section v-if="broadcaster.bio">
          <h2>{{ t('broadcaster.introduction') }}</h2>
          <p class="br-intro">{{ broadcaster.bio }}</p>
        </section>

        <section v-if="availabilityLabel || hours.length > 0">
          <h2>{{ t('broadcaster.availabilityTitle') }}</h2>
          <div class="br-fields">
            <div v-if="availabilityLabel" class="br-field">
              <span class="br-field__label">{{ t('broadcaster.availability') }}</span>
              <div>{{ availabilityLabel }}</div>
            </div>
            <div v-if="hours.length > 0" class="br-field">
              <span class="br-field__label">{{ t('broadcaster.hours') }}</span>
              <div v-for="h in hours" :key="h.day">
                {{ h.day }}: {{ h.slots }}
              </div>
            </div>
          </div>
        </section>

        <section v-if="profileFields.length > 0">
          <h2>{{ t('broadcaster.profile') }}</h2>
          <div class="br-fields br-fields--three">
            <div v-for="field in profileFields" :key="field.label" class="br-field">
              <span class="br-field__label">{{ field.label }}</span>
              <div>{{ field.value }}</div>
            </div>
          </div>
        </section>

        <section v-if="broadcaster.services.length > 0">
          <h2>{{ t('broadcaster.services') }}</h2>
          <div class="br-services">
            <span v-for="service in broadcaster.services" :key="service">{{ service }}</span>
          </div>
        </section>
      </article>
    </main>
  </q-page>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { useMeta } from 'quasar';
import { useI18n } from 'vue-i18n';
import type { broadcasters } from '@/encore/client';
import { api } from '@/encore/api';
import LiveStatusBadge from '@/components/LiveStatusBadge.vue';

const { t } = useI18n();
const route = useRoute();

const broadcaster = ref<broadcasters.Broadcaster | null>(null);
const loading = ref(true);
const current = ref(0);

const isLive = ref(false);

const homeRoute = computed(() => (route.params.locale ? `/${String(route.params.locale)}` : '/'));
const liveRoute = computed(
  () => `${route.params.locale ? `/${String(route.params.locale)}` : ''}/br/${String(route.params.shortId || '').toLowerCase()}/live`,
);

useMeta(() => ({
  title: `${broadcaster.value?.display_name || t('broadcaster.profile')} - PublicSite`,
  meta: {
    description: {
      name: 'description',
      content: broadcaster.value?.bio?.slice(0, 160) || t('broadcaster.profile'),
    },
  },
}));

// Profilo caricato nel browser: la pagina non è pre-renderizzata (SSG).
async function load() {
  const shortId = String(route.params.shortId || '').toLowerCase();
  loading.value = true;
  current.value = 0;
  isLive.value = false;
  // Lo stato della diretta è facoltativo: se non arriva la pagina resta uguale.
  void api.videochat
    .GetLive(shortId)
    .then((status) => {
      if (String(route.params.shortId || '').toLowerCase() === shortId) isLive.value = status.live;
    })
    .catch(() => {});
  try {
    broadcaster.value = await api.broadcasters.GetBroadcasterByShortID(shortId);
  } catch {
    broadcaster.value = null;
  } finally {
    loading.value = false;
  }
}

onMounted(load);
watch(() => route.params.shortId, load);

const photos = computed(() => (broadcaster.value?.photos ?? []).filter((p) => p.image_url));
const currentPhoto = computed(() => photos.value[current.value] ?? null);

function show(i: number) {
  const n = photos.value.length;
  if (n > 0) current.value = (i + n) % n;
}

// "available" -> "Available"
const availabilityLabel = computed(() => {
  const value = broadcaster.value?.availability?.trim() ?? '';
  return value ? value.charAt(0).toUpperCase() + value.slice(1) : '';
});

const hours = computed(() =>
  (broadcaster.value?.hours ?? [])
    .filter((h) => h.hours.length > 0)
    .map((h) => ({ day: h.day, slots: h.hours.map((s) => `${s.from}–${s.to}`).join(', ') })),
);

// Attributi (es. età, città) e lingue
const profileFields = computed(() => {
  const b = broadcaster.value;
  if (!b) return [];
  const fields = b.attributes.filter((a) => a.label && a.value).map((a) => ({ label: a.label, value: a.value }));
  if (b.languages.length > 0) {
    fields.push({ label: t('broadcaster.languages'), value: b.languages.join(', ') });
  }
  return fields;
});
</script>

<style scoped>
/* Stile dal modello profilo.html */
.br-page {
  --br-text: #e9e9e9;
  --br-muted: #929292;
  --br-line: #252525;
  --br-accent: #f4cf00;
  background: #050505;
  color: var(--br-text);
  font: 20px/1.55 Arial, Helvetica, sans-serif;
}

.br-status {
  padding: 80px 16px;
  text-align: center;
}

.br-link {
  color: var(--br-accent);
}

.br-layout {
  display: grid;
  grid-template-columns: 348px minmax(0, 1fr);
  gap: 27px;
  max-width: 1200px;
  margin: 0 auto;
  padding: 40px 26px 75px;
}

.br-hero {
  position: relative;
  height: 480px;
}

.br-hero__img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.br-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #343434;
  background: linear-gradient(145deg, #292929, #151515);
  color: #888888;
  font-size: 15px;
}

.br-nav {
  position: absolute;
  top: 45%;
  width: 50px;
  height: 55px;
  border: 0;
  background: #ffffff30;
  color: #ffffff;
  font-size: 35px;
  line-height: 1;
  cursor: pointer;
}

.br-nav:focus-visible {
  outline: 2px solid #bbbbbb;
}

.br-nav--prev {
  left: 12px;
}

.br-nav--next {
  right: 12px;
}

.br-thumbs {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 22px 30px;
  margin-top: 44px;
  padding: 0 10px;
}

.br-thumb {
  width: 100%;
  aspect-ratio: 0.8;
  padding: 0;
  border: 1px solid #343434;
  background: #151515;
  cursor: pointer;
}

.br-thumb img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.br-thumb--active,
.br-thumb:focus-visible {
  outline: 2px solid #bbbbbb;
  outline-offset: 2px;
}

.br-header {
  margin-bottom: 20px;
  padding-left: 20px;
  border-left: 20px solid var(--br-line);
}

h1 {
  margin: 0 0 7px;
  font-size: 43px;
  font-weight: 400;
  line-height: 1.1;
  letter-spacing: 0;
}

.br-subtitle {
  color: var(--br-accent);
  font-size: 22px;
}

.br-live {
  display: inline-block;
  margin-top: 10px;
  padding: 4px 12px;
  border-radius: 4px;
  background: #d32f2f;
  color: #ffffff;
  font-size: 16px;
  font-weight: 700;
  text-decoration: none;
}

.br-live:hover,
.br-live:focus-visible {
  background: #b71c1c;
  outline: none;
}

.br-live__dot {
  animation: br-blink 1.4s ease-in-out infinite;
}

@keyframes br-blink {
  50% {
    opacity: 0.3;
  }
}

@media (prefers-reduced-motion: reduce) {
  .br-live__dot {
    animation: none;
  }
}

.br-certified {
  margin-left: 12px;
  color: #45b34e;
  font-size: 17px;
}

h2 {
  margin: 24px 0 18px;
  color: #8a8a8a;
  font-size: 34px;
  font-weight: 400;
  line-height: 1.2;
  letter-spacing: 0;
}

.br-intro {
  margin: 0 0 22px;
  line-height: 1.65;
  white-space: pre-line;
}

.br-fields {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 26px;
}

.br-fields--three {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.br-field {
  min-width: 0;
  margin-bottom: 14px;
  padding: 0 0 10px;
  border-bottom: 1px dotted var(--br-line);
  overflow-wrap: anywhere;
}

.br-field__label {
  display: block;
  margin-bottom: 5px;
  color: var(--br-muted);
  font-size: 21px;
}

.br-services {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px 24px;
}

@media (max-width: 800px) {
  .br-layout {
    grid-template-columns: 1fr;
    max-width: 620px;
    padding-top: 20px;
  }

  .br-hero {
    height: 430px;
  }

  .br-thumbs {
    gap: 12px;
    margin-top: 18px;
  }

  .br-fields--three {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 450px) {
  .br-page {
    font-size: 17px;
  }

  .br-layout {
    padding-right: 16px;
    padding-left: 16px;
  }

  .br-hero {
    height: 360px;
  }

  h1 {
    font-size: 36px;
  }

  h2 {
    font-size: 29px;
  }

  .br-fields,
  .br-fields--three {
    grid-template-columns: 1fr;
  }

  .br-services {
    grid-template-columns: repeat(2, 1fr);
  }

  .br-field__label {
    font-size: 18px;
  }
}
</style>
