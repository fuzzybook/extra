<template>
  <div class="promo-calendar-container column q-gutter-md">
    <!-- Header Calendario e Controlli Globali -->
    <q-card flat bordered class="q-pa-md bg-white">
      <div class="row items-center justify-between">
        <div class="row items-center q-gutter-sm">
          <q-icon name="calendar_month" size="28px" color="primary" />
          <div>
            <div class="text-h6 text-weight-bold">{{ t('promos.calendar.title', { year: selectedYear }) }}</div>
            <div class="text-caption text-grey-7">{{ t('promos.calendar.subtitle') }}</div>
          </div>
        </div>

        <div class="row items-center q-gutter-md">
          <!-- Selettore Anno -->
          <div class="row items-center q-gutter-xs bg-grey-1 q-px-sm q-py-xs rounded-borders">
            <q-btn flat round dense icon="chevron_left" size="sm" @click="selectedYear--" />
            <span class="text-subtitle1 text-weight-bolder text-primary q-px-sm">{{ selectedYear }}</span>
            <q-btn flat round dense icon="chevron_right" size="sm" @click="selectedYear++" />
          </div>
          <q-btn
            v-if="selectedYear !== currentYear"
            flat
            dense
            size="sm"
            color="primary"
            :label="t('promos.calendar.currentYear')"
            @click="selectedYear = currentYear"
          />
        </div>
      </div>
    </q-card>

    <!-- 3 Calendari: uno per ciascuna PromoPosition (Promo Left, Promo Right 1, Promo Right 2) -->
    <q-card
      v-for="pos in calendarPositions"
      :key="pos.value"
      flat
      bordered
      class="q-pa-md bg-white"
    >
      <!-- Intestazione del Calendario Posizione -->
      <div class="row items-center justify-between q-mb-md">
        <div class="row items-center q-gutter-sm">
          <q-avatar :icon="pos.icon" :color="pos.color" text-color="white" size="32px" rounded />
          <div>
            <div class="text-subtitle1 text-weight-bold row items-center q-gutter-xs">
              <span>{{ pos.label }}</span>
              <q-badge
                :color="pos.color"
                :label="t('promos.calendar.promosCount', { count: getPromosForPosition(pos.value).length })"
                rounded
                class="q-ml-sm"
              />
            </div>
            <div class="text-caption text-grey-7">
              {{ t('promos.calendar.posCoverage', { pos: pos.label }) }}
            </div>
          </div>
        </div>
      </div>

      <!-- Tabella Grafica delle 52 Settimane per la Posizione Corrente -->
      <div class="calendar-scroll-container scroll">
        <table class="calendar-table">
          <thead>
            <tr>
              <th class="sticky-col-promo text-left q-pa-xs">{{ t('promos.calendar.columnPromo') }}</th>
              <th
                v-for="w in 52"
                :key="w"
                :class="['week-col-header text-center', isCurrentWeek(w) ? 'current-week-header' : '']"
              >
                W{{ w }}
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="getPromosForPosition(pos.value).length === 0">
              <td colspan="53" class="text-center q-pa-lg text-grey-6 text-italic">
                {{ t('promos.calendar.emptyPos', { pos: pos.label, year: selectedYear }) }}
              </td>
            </tr>
            <tr
              v-for="p in getPromosForPosition(pos.value)"
              :key="p.id"
              class="promo-row cursor-pointer"
              @click="emit('edit', p)"
            >
              <!-- Colonna sticky con dettagli promo -->
              <td class="sticky-col-promo q-pa-xs">
                <div class="row items-center no-wrap q-gutter-xs">
                  <q-badge
                    :color="pos.color"
                    rounded
                    :label="`#${p.id}`"
                    class="q-mr-xs text-weight-bold"
                  />
                  <span class="text-weight-medium ellipsis col">
                    {{ getPromoTitle(p) }}
                  </span>
                  <!-- Badge settimane selezionate con colore sezione e tooltip con nome sezione -->
                  <q-badge
                    :color="getSectionMeta(p.position, p.section).color"
                    :label="getWeeksCountLabel(p)"
                    rounded
                    class="q-ml-xs text-weight-bold cursor-pointer"
                  >
                    <q-tooltip anchor="top middle" self="bottom middle">
                      {{ t('promos.calendar.sectionLabel', { section: getSectionMeta(p.position, p.section).label }) }}
                    </q-tooltip>
                  </q-badge>
                  <q-tooltip>
                    <div class="text-weight-bold">#{{ p.id }} - {{ getPromoTitle(p) }}</div>
                    <div>{{ t('promos.calendar.positionLabel', { pos: pos.label }) }}</div>
                    <div>{{ t('promos.calendar.sectionLabel', { section: getSectionMeta(p.position, p.section).label }) }}</div>
                    <div>{{ t('promos.calendar.weeksLabel', { weeks: getWeeksCountLabel(p) }) }}</div>
                    <div>{{ t('promos.calendar.clickToEdit') }}</div>
                  </q-tooltip>
                </div>
              </td>

              <!-- 52 celle settimanali con barre di copertura -->
              <td
                v-for="w in 52"
                :key="w"
                :class="['week-cell text-center', isCurrentWeek(w) ? 'current-week-cell' : '']"
              >
                <div
                  v-if="isPromoActiveInWeek(p, w)"
                  :class="['week-active-bar', `bg-${pos.color}`]"
                >
                  <q-tooltip>
                    <div class="text-weight-bold">#{{ p.id }} - {{ getPromoTitle(p) }}</div>
                    <div>{{ t('promos.calendar.positionLabel', { pos: pos.label }) }} | {{ t('promos.calendar.sectionLabel', { section: getSectionMeta(p.position, p.section).label }) }}</div>
                    <div>{{ t('promos.calendar.activeInWeek', { week: w, year: selectedYear }) }}</div>
                  </q-tooltip>
                </div>
                <div v-else class="week-empty-dot"></div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </q-card>
  </div>
</template>

<style scoped>
.calendar-scroll-container {
  overflow-x: auto;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
}

.calendar-table {
  width: 100%;
  border-collapse: collapse;
  min-width: 1300px;
  font-size: 11px;
}

.calendar-table th,
.calendar-table td {
  border: 1px solid #eeeeee;
  padding: 4px;
}

.sticky-col-promo {
  position: sticky;
  left: 0;
  background: #fafafa;
  z-index: 2;
  min-width: 220px;
  max-width: 260px;
  box-shadow: 2px 0 4px rgba(0, 0, 0, 0.05);
}

.week-col-header {
  min-width: 24px;
  max-width: 26px;
  font-size: 9px;
  font-weight: 600;
  background: #f5f5f5;
  color: #616161;
}

.current-week-header {
  background: #e3f2fd !important;
  color: #1976d2 !important;
  font-weight: bold;
}

.week-cell {
  height: 28px;
  padding: 2px !important;
}

.current-week-cell {
  background: #f9fbff;
}

.week-active-bar {
  height: 18px;
  border-radius: 3px;
  transition: transform 0.1s ease;
}

.week-active-bar:hover {
  transform: scaleY(1.2);
}

.week-empty-dot {
  height: 4px;
  width: 4px;
  background: #e0e0e0;
  border-radius: 50%;
  margin: 0 auto;
}

.promo-row:hover {
  background: #f0f7ff;
}
.promo-row:hover .sticky-col-promo {
  background: #f0f7ff;
}
</style>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import type { advertising } from '@/encore/client';
import { PromoPosition, PROMO_POSITION_MAP, getSectionMeta } from '@/types/promo';

type Promo = advertising.Promo;

const { t } = useI18n();

const props = defineProps<{
  promos: Promo[];
}>();

const emit = defineEmits<{
  'edit': [promo: Promo];
}>();

const currentYear = new Date().getFullYear();
const selectedYear = ref<number>(currentYear);

const calendarPositions = [
  { value: PromoPosition.PromoLeft, ...PROMO_POSITION_MAP[PromoPosition.PromoLeft] },
  { value: PromoPosition.PromoRight1, ...PROMO_POSITION_MAP[PromoPosition.PromoRight1] },
  { value: PromoPosition.PromoRight2, ...PROMO_POSITION_MAP[PromoPosition.PromoRight2] },
];

const filteredPromos = computed(() => {
  return props.promos.filter((p) => {
    const promoYear = p.year || (p.active_from ? new Date(p.active_from).getFullYear() : selectedYear.value);
    return promoYear === selectedYear.value;
  });
});

function getPromosForPosition(position: PromoPosition): Promo[] {
  return filteredPromos.value.filter((p) => Number(p.position) === Number(position));
}

function isPromoActiveInWeek(promo: Promo, week: number): boolean {
  if (promo.weeks && promo.weeks.length > 0) {
    return promo.weeks.includes(week);
  }
  // Se non ha settimane esplicite, calcola la settimana da active_from
  if (promo.active_from) {
    const d = new Date(promo.active_from);
    if (d.getFullYear() === selectedYear.value) {
      const w = getISOWeekNumber(d);
      return w === week;
    }
  }
  return false;
}

function getISOWeekNumber(date: Date): number {
  const d = new Date(Date.UTC(date.getFullYear(), date.getMonth(), date.getDate()));
  const dayNum = d.getUTCDay() || 7;
  d.setUTCDate(d.getUTCDate() + 4 - dayNum);
  const yearStart = new Date(Date.UTC(d.getUTCFullYear(), 0, 1));
  return Math.ceil(((d.getTime() - yearStart.getTime()) / 86400000 + 1) / 7);
}

function isCurrentWeek(week: number): boolean {
  const now = new Date();
  if (now.getFullYear() !== selectedYear.value) return false;
  return getISOWeekNumber(now) === week;
}

function getPromoTitle(p: Promo): string {
  if (p.item && p.item.length > 0 && p.item[0]?.title) {
    return p.item[0].title;
  }
  return `Promo #${p.id}`;
}

function getWeeksCount(p: Promo): number {
  if (p.weeks && p.weeks.length > 0) {
    return p.weeks.length;
  }
  if (p.active_from) {
    const d = new Date(p.active_from);
    if (d.getFullYear() === selectedYear.value) {
      return 1;
    }
  }
  return 0;
}

function getWeeksCountLabel(p: Promo): string {
  const count = getWeeksCount(p);
  return t('promos.calendar.weeksCount', { count });
}
</script>
