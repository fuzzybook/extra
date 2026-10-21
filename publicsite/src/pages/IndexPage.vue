<template>
  <q-page class="home-page fill-width q-py-md">
    <div class="container row wrap q-col-gutter-md">
      <!-- Promo Left Column (col-2) -->
      <div class="promo__left col-2">
        <div class="promo-box q-pa-sm">
          <div v-if="promoLeftList.length > 0" class="q-mt-sm">
            <div v-for="promo in promoLeftList" :key="promo.id" class="q-mb-xs">
              <div class="text-caption text-weight-bold">{{ promo.section }}</div>
              <div v-for="(it, i) in promo.item" :key="i">
                <div v-for="(c, ci) in it.content" :key="ci" class="q-my-xs">
                  <PromoBroadcaster
                    v-if="c.broadcaster?.image_url"
                    :broadcaster="c.broadcaster"
                    :alt="c.broadcaster.display_name || it.title || promo.section"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Main Central Section: Broadcasters List / Home Card -->
      <div class="broadcasters__list col-12 col-md">
        <div class="text-center q-pa-xl home-card">
          <div class="text-h4 text-weight-bold text-white q-mb-sm">
            {{ t('home.title') }}
          </div>
          <p class="text-body1 text-grey-4 q-mb-lg">
            {{ t('home.subtitle') }}
          </p>

          <q-btn
            to="/dashboard"
            color="primary"
            size="lg"
            unelevated
            icon="dashboard"
            :label="t('actions.goToDashboard')"
            class="q-px-xl q-py-sm rounded-borders text-weight-bold"
          />
        </div>
      </div>

      <!-- Promo Right Column (col-2) - Visualizzato solo se ci sono contenuti -->
      <div v-if="hasRightPromo" class="promo__right-container col-2">
        <div class="row q-col-gutter-xs">
          <!-- promo_right1 -->
          <div
            v-if="promoRight1List.length > 0"
            class="promo__right1"
            :class="promoRight2List.length > 0 ? 'col-6' : 'col-12'"
          >
            <div class="promo-box q-pa-sm">
              <div v-for="promo in promoRight1List" :key="promo.id" class="q-mb-xs">
                <div class="text-caption text-weight-bold">{{ promo.section }}</div>
                <div v-for="(it, i) in promo.item" :key="i">
                  <div v-for="(c, ci) in it.content" :key="ci" class="q-my-xs">
                    <PromoBroadcaster
                      v-if="c.broadcaster?.image_url"
                      :broadcaster="c.broadcaster"
                      :alt="c.broadcaster.display_name || it.title || promo.section"
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- promo_right2 -->
          <div
            v-if="promoRight2List.length > 0"
            class="promo__right2"
            :class="promoRight1List.length > 0 ? 'col-6' : 'col-12'"
          >
            <div class="promo-box q-pa-sm">
              <div v-for="promo in promoRight2List" :key="promo.id" class="q-mb-xs">
                <div class="text-caption text-weight-bold">{{ promo.section }}</div>
                <div v-for="(it, i) in promo.item" :key="i">
                  <div v-for="(c, ci) in it.content" :key="ci" class="q-my-xs">
                    <PromoBroadcaster
                      v-if="c.broadcaster?.image_url"
                      :broadcaster="c.broadcaster"
                      :alt="c.broadcaster.display_name || it.title || promo.section"
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </q-page>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useMeta } from 'quasar';
import { useI18n } from 'vue-i18n';
import { type advertising } from '@/encore/client';
import { api } from '@/encore/api';
import PromoBroadcaster from '@/components/PromoBroadcaster.vue';

const { t } = useI18n();

useMeta(() => ({
  title: t('menu.home'),
  meta: {
    description: {
      name: 'description',
      content: t('home.subtitle'),
    },
  },
}));

type WeekPromoResponse = advertising.WeekPromoResponse;

const promoData = ref<WeekPromoResponse | null>(null);

// Promo della settimana corrente, caricate nel browser (non nel pre-render SSG).
async function loadPromo() {
  try {
    promoData.value = await api.advertising.GetCurrentWeekPromo({ week: 0, year: 0 });
  } catch (err: unknown) {
    console.warn('Impossibile caricare le promo:', err);
    promoData.value = null;
  }
}

onMounted(() => {
  void loadPromo();
});

const promoLeftList = computed<advertising.Promo[]>(() => {
  return promoData.value?.promo_left ?? [];
});

const promoRight1List = computed<advertising.Promo[]>(() => {
  return promoData.value?.promo_right1 ?? [];
});

const promoRight2List = computed<advertising.Promo[]>(() => {
  return promoData.value?.promo_right2 ?? [];
});

const hasRightPromo = computed(() => {
  return promoRight1List.value.length > 0 || promoRight2List.value.length > 0;
});
</script>

<style scoped>
.home-page {
  min-height: calc(100vh - 180px);
  background-color: #0b0b0e;
  color: #ffffff;
}

.container {
  max-width: 1400px;
  margin: 0 auto;
}

.home-card {
  max-width: 600px;
  margin: 0 auto;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 16px;
  backdrop-filter: blur(10px);
}

.promo__left,
.promo__right-container,
.promo__right1,
.promo__right2 {
  min-width: 0;
  max-width: 100%;
}

.promo-box {
  background: rgba(255, 255, 255, 0.02);
  border: 1px dashed rgba(255, 255, 255, 0.15);
  border-radius: 6px;
  min-height: 120px;
  width: 100%;
  max-width: 100%;
  box-sizing: border-box;
  overflow: hidden;
}

.promo-header {
  font-family: monospace;
  letter-spacing: 0.5px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  padding-bottom: 4px;
}

.json-dump {
  font-family: SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
  font-size: 10px;
  line-height: 1.35;
  color: #00e5ff;
  white-space: pre-wrap;
  word-break: break-all;
  margin: 0;
  max-height: 380px;
  overflow-y: auto;
}
</style>
