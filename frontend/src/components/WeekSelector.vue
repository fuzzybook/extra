<template>
  <div class="week-selector q-pa-sm bg-grey-1 rounded-borders">
    <!-- Intestazione con Selettore Anno e Statistiche -->
    <div class="row items-center justify-between q-mb-sm">
      <div class="row items-center q-gutter-sm">
        <q-icon name="date_range" color="primary" size="24px" />
        <span class="text-subtitle2 text-weight-bold text-grey-9">
          Settimane Lavorative (Anno {{ currentYear }})
        </span>
      </div>

      <div class="row items-center q-gutter-xs">
        <q-btn
          dense
          flat
          round
          icon="chevron_left"
          size="sm"
          :disable="currentYear <= minSelectableYear"
          @click="currentYear--"
        >
          <q-tooltip>Anno precedente</q-tooltip>
        </q-btn>
        <span class="text-weight-bold q-px-xs text-primary">{{ currentYear }}</span>
        <q-btn
          dense
          flat
          round
          icon="chevron_right"
          size="sm"
          @click="currentYear++"
        >
          <q-tooltip>Anno successivo</q-tooltip>
        </q-btn>
      </div>
    </div>

    <!-- Toolbar pulsanti rapidi -->
    <div class="row items-center q-gutter-xs q-mb-sm">
      <q-btn
        size="xs"
        unelevated
        color="primary"
        label="Tutte (1-52)"
        :disable="allWeeksPast"
        @click="selectAll"
      />
      <q-btn
        size="xs"
        flat
        color="grey-8"
        label="Nessuna"
        @click="clearAll"
      />
      <q-separator vertical class="q-mx-xs" />
      <q-btn
        size="xs"
        outline
        color="primary"
        label="Q1 (1-13)"
        :disable="isQuarterPast(1)"
        @click="toggleQuarter(1)"
      />
      <q-btn
        size="xs"
        outline
        color="primary"
        label="Q2 (14-26)"
        :disable="isQuarterPast(2)"
        @click="toggleQuarter(2)"
      />
      <q-btn
        size="xs"
        outline
        color="primary"
        label="Q3 (27-39)"
        :disable="isQuarterPast(3)"
        @click="toggleQuarter(3)"
      />
      <q-btn
        size="xs"
        outline
        color="primary"
        label="Q4 (40-52)"
        :disable="isQuarterPast(4)"
        @click="toggleQuarter(4)"
      />
      <q-separator vertical class="q-mx-xs" />
      <q-btn
        size="xs"
        flat
        color="secondary"
        label="Pari"
        :disable="allWeeksPast"
        @click="selectEven"
      />
      <q-btn
        size="xs"
        flat
        color="secondary"
        label="Dispari"
        :disable="allWeeksPast"
        @click="selectOdd"
      />
    </div>

    <!-- Griglia trimestrale delle 52 settimane -->
    <div class="quarters-container q-gutter-y-xs">
      <div
        v-for="q in quarters"
        :key="q.name"
        class="quarter-row row items-center q-col-gutter-xs"
      >
        <div class="col-12 col-sm-auto text-caption text-weight-bold text-grey-7" style="min-width: 32px">
          {{ q.name }}
        </div>
        <div class="col row items-center q-gutter-xs">
          <q-btn
            v-for="w in q.weeks"
            :key="w"
            :color="isWeekSelected(w) ? 'primary' : 'white'"
            :text-color="isWeekSelected(w) ? 'white' : (isWeekPast(w) ? 'grey-5' : (isCurrentWeek(w) ? 'primary' : 'grey-9'))"
            :outline="!isWeekSelected(w)"
            size="xs"
            dense
            :class="[
              'week-chip text-weight-medium',
              isCurrentWeek(w) ? 'current-week-chip' : '',
              isWeekPast(w) ? 'past-week-chip' : ''
            ]"
            style="min-width: 28px; padding: 2px 4px"
            :disable="isWeekPast(w)"
            @click="toggleWeek(w)"
          >
            {{ w }}
            <q-tooltip>
              Settimana {{ w }} ({{ getWeekDateRange(w, currentYear) }})
              <template v-if="isCurrentWeek(w)"> • Settimana corrente</template>
              <template v-else-if="isWeekPast(w)"> • Passata (non selezionabile)</template>
            </q-tooltip>
          </q-btn>
        </div>
      </div>
    </div>

    <!-- Riepilogo selezione -->
    <div class="row items-center justify-between q-mt-sm text-caption text-grey-8">
      <div>
        <span v-if="selectedWeeks.length === 0" class="text-italic text-grey-6">
          Nessuna settimana selezionata
        </span>
        <span v-else>
          <strong>{{ selectedWeeks.length }}</strong> settiman{{ selectedWeeks.length === 1 ? 'a' : 'e' }} selezionat{{ selectedWeeks.length === 1 ? 'a' : 'e' }}:
          <span class="text-primary text-weight-bold">{{ formatSelectedWeeksSummary() }}</span>
        </span>
      </div>

      <q-badge
        :color="selectedWeeks.length > 0 ? 'positive' : 'grey-5'"
        :label="selectedWeeks.length > 0 ? 'Valida nel calendario' : 'Inattiva'"
        rounded
      />
    </div>
  </div>
</template>

<style scoped>
.week-selector {
  border: 1px solid #e0e0e0;
}
.week-chip {
  transition: all 0.15s ease;
}
.current-week-chip {
  border: 1.5px solid var(--q-primary) !important;
  font-weight: 700;
}
.past-week-chip {
  opacity: 0.45;
  cursor: not-allowed;
}
</style>

<script setup lang="ts">
import { computed } from 'vue';

const props = withDefaults(
  defineProps<{
    modelValue: number[];
    year?: number;
  }>(),
  {
    modelValue: () => [],
    year: () => new Date().getFullYear(),
  }
);

const emit = defineEmits<{
  'update:modelValue': [weeks: number[]];
  'update:year': [year: number];
}>();

const currentYear = computed({
  get: () => props.year,
  set: (val: number) => emit('update:year', val),
});

const selectedWeeks = computed({
  get: () => props.modelValue,
  set: (val: number[]) => emit('update:modelValue', val),
});

interface Quarter {
  name: string;
  weeks: number[];
}

const quarters: Quarter[] = [
  { name: 'Q1', weeks: Array.from({ length: 13 }, (_, i) => i + 1) },
  { name: 'Q2', weeks: Array.from({ length: 13 }, (_, i) => i + 14) },
  { name: 'Q3', weeks: Array.from({ length: 13 }, (_, i) => i + 27) },
  { name: 'Q4', weeks: Array.from({ length: 13 }, (_, i) => i + 40) },
];

function getISOWeekInfo(date: Date) {
  const d = new Date(Date.UTC(date.getFullYear(), date.getMonth(), date.getDate()));
  const dayNum = d.getUTCDay() || 7;
  d.setUTCDate(d.getUTCDate() + 4 - dayNum);
  const isoYear = d.getUTCFullYear();
  const yearStart = new Date(Date.UTC(isoYear, 0, 1));
  const weekNo = Math.ceil(((d.getTime() - yearStart.getTime()) / 86400000 + 1) / 7);
  return { isoYear, weekNo };
}

const minSelectableYear = computed(() => {
  const currentIsoYear = getISOWeekInfo(new Date()).isoYear;
  return Math.min(props.year ?? currentIsoYear, currentIsoYear);
});

function isWeekPast(week: number, year = currentYear.value): boolean {
  const nowInfo = getISOWeekInfo(new Date());
  if (year < nowInfo.isoYear) return true;
  if (year > nowInfo.isoYear) return false;
  return week < nowInfo.weekNo;
}

function isCurrentWeek(week: number, year = currentYear.value): boolean {
  const nowInfo = getISOWeekInfo(new Date());
  return year === nowInfo.isoYear && week === nowInfo.weekNo;
}

function isQuarterPast(qNum: number): boolean {
  const q = quarters[qNum - 1];
  if (!q) return false;
  return q.weeks.every((w) => isWeekPast(w));
}

const allWeeksPast = computed(() => {
  return Array.from({ length: 52 }, (_, i) => i + 1).every((w) => isWeekPast(w));
});

function isWeekSelected(w: number): boolean {
  return selectedWeeks.value.includes(w);
}

function toggleWeek(w: number) {
  if (isWeekPast(w)) return;
  const current = [...selectedWeeks.value];
  const idx = current.indexOf(w);
  if (idx >= 0) {
    current.splice(idx, 1);
  } else {
    current.push(w);
    current.sort((a, b) => a - b);
  }
  emit('update:modelValue', current);
}

function selectAll() {
  const available = Array.from({ length: 52 }, (_, i) => i + 1).filter((w) => !isWeekPast(w));
  emit('update:modelValue', available);
}

function clearAll() {
  emit('update:modelValue', []);
}

function toggleQuarter(qNum: number) {
  const q = quarters[qNum - 1];
  if (!q) return;

  const available = q.weeks.filter((w) => !isWeekPast(w));
  if (available.length === 0) return;

  const current = new Set(selectedWeeks.value);
  const allAvailableSelected = available.every((w) => current.has(w));

  if (allAvailableSelected) {
    available.forEach((w) => current.delete(w));
  } else {
    available.forEach((w) => current.add(w));
  }

  const sorted = Array.from(current).sort((a, b) => a - b);
  emit('update:modelValue', sorted);
}

function selectEven() {
  const evens = Array.from({ length: 52 }, (_, i) => i + 1).filter((w) => w % 2 === 0 && !isWeekPast(w));
  emit('update:modelValue', evens);
}

function selectOdd() {
  const odds = Array.from({ length: 52 }, (_, i) => i + 1).filter((w) => w % 2 !== 0 && !isWeekPast(w));
  emit('update:modelValue', odds);
}

function getMondayOfISOWeek(week: number, year: number): Date {
  const date = new Date(year, 0, 1 + (week - 1) * 7);
  const day = date.getDay();
  const diff = date.getDate() - day + (day === 0 ? -6 : 1);
  return new Date(date.setDate(diff));
}

function getWeekDateRange(week: number, year: number): string {
  const monday = getMondayOfISOWeek(week, year);
  const sunday = new Date(monday);
  sunday.setDate(monday.getDate() + 6);
  return `${monday.getDate()} ${monday.toLocaleString('it', { month: 'short' })} - ${sunday.getDate()} ${sunday.toLocaleString('it', { month: 'short' })}`;
}

function formatSelectedWeeksSummary(): string {
  const sorted = [...selectedWeeks.value].sort((a, b) => a - b);
  if (sorted.length <= 8) {
    return sorted.map((w) => `W${w}`).join(', ');
  }
  return `${sorted.slice(0, 6).map((w) => `W${w}`).join(', ')} ... (+${sorted.length - 6})`;
}
</script>
