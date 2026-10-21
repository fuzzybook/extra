<template>
  <q-dialog
    :model-value="modelValue"
    persistent
    @update:model-value="(val: boolean) => emit('update:modelValue', val)"
    @show="onShow"
  >
    <q-card style="width: 780px; max-width: 95vw; max-height: 90vh" class="column no-wrap">
      <!-- Header Principale -->
      <q-card-section class="row items-center justify-between bg-teal-7 text-white q-py-sm shrink-0" style="flex-shrink: 0">
        <div class="row items-center q-gutter-sm">
          <q-icon name="schedule" size="26px" />
          <div>
            <div class="text-h6 text-weight-bold">{{ t('broadcasters.hoursDialog.title') }}</div>
            <div class="text-caption opacity-90" v-if="broadcaster">
              {{ broadcaster.display_name }} (ID #{{ broadcaster.id }})
            </div>
          </div>
        </div>
        <q-btn icon="close" flat round dense color="white" @click="closeDialog" />
      </q-card-section>

      <!-- Barra di Riepilogo e Azioni Rapide -->
      <div class="bg-teal-50 q-px-md q-py-xs row items-center justify-between text-caption border-bottom shrink-0" style="flex-shrink: 0">
        <div class="row items-center q-gutter-md">
          <span class="text-teal-9 text-weight-bold">
            <q-icon name="info" size="16px" class="q-mr-xs" />
            {{ t('broadcasters.hoursDialog.totalSlots') }}
            <q-badge color="teal-8" class="q-ml-xs">{{ totalConfiguredSlots }}</q-badge>
          </span>
          <span class="text-grey-7">
            {{ t('broadcasters.hoursDialog.daysWithSlots', { active: activeDaysCount, total: hoursList.length }) }}
          </span>
        </div>
        <div class="row q-gutter-xs">
          <q-btn
            flat
            dense
            color="teal-8"
            icon="add_circle"
            :label="t('broadcasters.hoursDialog.addDay')"
            size="sm"
            @click="showAddDayPrompt = true"
          />
        </div>
      </div>

      <!-- Corpo del Dialog -->
      <q-card-section class="col scroll q-pa-md">
        <q-banner v-if="localError" class="bg-negative text-white q-mb-md rounded-borders" dense>
          <template #avatar>
            <q-icon name="error" />
          </template>
          {{ localError }}
        </q-banner>

        <!-- Lista Giorni Settimanali -->
        <div class="q-gutter-y-sm">
          <q-card
            v-for="(dayItem, dayIndex) in hoursList"
            :key="dayItem.day"
            flat
            bordered
            class="day-card"
            :class="{ 'day-card--has-slots': dayItem.hours.length > 0 }"
          >
            <q-card-section class="q-py-xs q-px-sm bg-grey-1 row items-center justify-between">
              <!-- Intestazione Giorno -->
              <div class="row items-center q-gutter-xs">
                <q-icon
                  name="calendar_today"
                  size="18px"
                  :color="dayItem.hours.length > 0 ? 'teal-7' : 'grey-5'"
                />
                <span class="text-subtitle2 text-weight-bold text-grey-9 q-ml-xs">
                  {{ dayItem.day }}
                </span>
                <q-badge
                  :color="dayItem.hours.length > 0 ? 'teal-7' : 'grey-5'"
                  class="q-ml-xs text-weight-medium"
                >
                  {{ dayItem.hours.length }} {{ dayItem.hours.length === 1 ? t('broadcasters.hoursDialog.slotsSingle') : t('broadcasters.hoursDialog.slotsPlural') }}
                </q-badge>
              </div>

              <!-- Azioni Giorno -->
              <div class="row items-center q-gutter-xs">
                <!-- Copia orari in altri giorni -->
                <q-btn
                  v-if="dayItem.hours.length > 0"
                  flat
                  round
                  dense
                  size="sm"
                  color="teal-8"
                  icon="content_copy"
                >
                  <q-tooltip>{{ t('broadcasters.hoursDialog.copyTooltip', { day: dayItem.day }) }}</q-tooltip>
                  <q-menu auto-close>
                    <q-list dense style="min-width: 200px">
                      <q-item clickable @click="copyDayHours(dayIndex, 'weekdays')">
                        <q-item-section avatar><q-icon name="work" size="18px" color="teal" /></q-item-section>
                        <q-item-section>{{ t('broadcasters.hoursDialog.copyWeekdays') }}</q-item-section>
                      </q-item>
                      <q-item clickable @click="copyDayHours(dayIndex, 'weekend')">
                        <q-item-section avatar><q-icon name="weekend" size="18px" color="deep-orange" /></q-item-section>
                        <q-item-section>{{ t('broadcasters.hoursDialog.copyWeekend') }}</q-item-section>
                      </q-item>
                      <q-item clickable @click="copyDayHours(dayIndex, 'all')">
                        <q-item-section avatar><q-icon name="event_repeat" size="18px" color="indigo" /></q-item-section>
                        <q-item-section>{{ t('broadcasters.hoursDialog.copyAll') }}</q-item-section>
                      </q-item>
                    </q-list>
                  </q-menu>
                </q-btn>

                <!-- Svuota orari del giorno -->
                <q-btn
                  v-if="dayItem.hours.length > 0"
                  flat
                  round
                  dense
                  size="sm"
                  color="grey-6"
                  icon="clear_all"
                  @click="clearDayHours(dayIndex)"
                >
                  <q-tooltip>{{ t('broadcasters.hoursDialog.clearTooltip', { day: dayItem.day }) }}</q-tooltip>
                </q-btn>

                <!-- Rimuovi giorno se personalizzato -->
                <q-btn
                  v-if="isCustomDay(dayItem.day)"
                  flat
                  round
                  dense
                  size="sm"
                  color="negative"
                  icon="delete_outline"
                  @click="removeDay(dayIndex)"
                >
                  <q-tooltip>{{ t('broadcasters.hoursDialog.removeDayTooltip') }}</q-tooltip>
                </q-btn>

                <!-- Pulsante Aggiungi Fascia Oraria -->
                <q-btn
                  unelevated
                  dense
                  color="teal-7"
                  icon="add"
                  :label="t('broadcasters.hoursDialog.addSlot')"
                  size="sm"
                  class="q-px-sm"
                  @click="openAddSlot(dayIndex)"
                />
              </div>
            </q-card-section>

            <!-- Lista Fasce Orarie del Giorno -->
            <q-card-section class="q-pa-sm">
              <div v-if="dayItem.hours.length === 0" class="row items-center text-grey-5 q-pa-xs">
                <q-icon name="access_time" size="16px" class="q-mr-xs" />
                <span class="text-caption">{{ t('broadcasters.hoursDialog.noSlots') }}</span>
                <q-btn
                  flat
                  dense
                  no-caps
                  color="teal-8"
                  :label="t('broadcasters.hoursDialog.addSlotQuick')"
                  size="xs"
                  class="q-ml-sm"
                  @click="openAddSlot(dayIndex)"
                />
              </div>

              <div v-else class="row q-gutter-xs items-center">
                <q-chip
                  v-for="(slot, slotIndex) in dayItem.hours"
                  :key="slotIndex"
                  removable
                  clickable
                  outline
                  color="teal-8"
                  text-color="teal-9"
                  icon="schedule"
                  class="slot-chip bg-white text-weight-medium"
                  @remove="removeSlot(dayIndex, slotIndex)"
                  @click="openEditSlot(dayIndex, slotIndex)"
                >
                  <span class="slot-text">{{ slot.from }} - {{ slot.to }}</span>
                  <span class="text-caption text-grey-6 q-ml-xs font-monospace">
                    ({{ getSlotDurationText(slot.from, slot.to) }})
                  </span>
                  <q-icon name="edit" size="13px" class="q-ml-xs text-teal-6" />
                  <q-tooltip>{{ t('broadcasters.hoursDialog.slotTooltip', { from: slot.from, to: slot.to }) }}</q-tooltip>
                </q-chip>
              </div>
            </q-card-section>
          </q-card>
        </div>
      </q-card-section>

      <q-separator style="flex-shrink: 0" />

      <!-- Footer Dialog Principale -->
      <q-card-actions align="right" class="q-pa-md bg-grey-1" style="flex-shrink: 0">
        <q-btn flat :label="t('actions.cancel')" :disable="loading" @click="closeDialog" />
        <q-btn
          color="teal-7"
          icon="save"
          :loading="loading"
          :label="t('broadcasters.hoursDialog.save')"
          unelevated
          @click="onSave"
        />
      </q-card-actions>
    </q-card>
  </q-dialog>

  <!-- ======================================================== -->
  <!-- POPUP PER AGGIUNTA / MODIFICA FASCIA ORARIA (2 x Q-TIME) -->
  <!-- ======================================================== -->
  <q-dialog v-model="slotDialogOpen" persistent>
    <q-card style="width: 720px; max-width: 96vw; max-height: 90vh" class="column no-wrap">
      <!-- Header Popup Fascia -->
      <q-card-section class="row items-center justify-between bg-teal-8 text-white q-py-sm" style="flex-shrink: 0">
        <div class="row items-center q-gutter-sm">
          <q-icon :name="editingSlotIndex !== null ? 'edit' : 'more_time'" size="22px" />
          <div>
            <div class="text-subtitle1 text-weight-bold">
              {{ editingSlotIndex !== null ? t('broadcasters.hoursDialog.editSlotTitle') : t('broadcasters.hoursDialog.newSlotTitle') }}
            </div>
            <div class="text-caption text-teal-1" v-if="activeDay">
              {{ t('broadcasters.hoursDialog.dayPrefix') }} <strong>{{ activeDay.day }}</strong>
            </div>
          </div>
        </div>
        <q-btn icon="close" flat round dense color="white" v-close-popup />
      </q-card-section>

      <!-- Corpo Popup con 2 q-time (scrollabile) -->
      <q-card-section class="col scroll q-pa-md">
        <!-- Indicatore Orario Selezionato in Tempo Reale -->
        <div class="row items-center justify-center q-mb-md">
          <q-banner
            rounded
            dense
            class="full-width text-center"
            :class="slotValidationError ? 'bg-red-1 text-negative border-negative' : 'bg-teal-1 text-teal-9 border-teal'"
          >
            <div class="row items-center justify-center q-gutter-sm text-subtitle1">
              <span class="text-weight-bold">{{ t('broadcasters.hoursDialog.from') }}</span>
              <q-badge color="teal-8" class="text-subtitle2 text-weight-bold q-px-sm">
                {{ slotForm.from || '--:--' }}
              </q-badge>
              <q-icon name="arrow_forward" size="18px" />
              <span class="text-weight-bold">{{ t('broadcasters.hoursDialog.to') }}</span>
              <q-badge color="teal-8" class="text-subtitle2 text-weight-bold q-px-sm">
                {{ slotForm.to || '--:--' }}
              </q-badge>
              <span v-if="!slotValidationError && slotDuration" class="text-caption text-teal-8 text-weight-medium">
                ({{ t('broadcasters.hoursDialog.duration', { duration: slotDuration }) }})
              </span>
            </div>

            <!-- Messaggio di Errore / Sovrapposizione -->
            <div v-if="slotValidationError" class="text-caption text-weight-bold q-mt-xs text-negative row items-center justify-center q-gutter-xs">
              <q-icon name="warning" size="16px" />
              <span>{{ slotValidationError }}</span>
            </div>
          </q-banner>
        </div>

        <!-- Sezione 2 Componenti q-time (Dalle e Alle) -->
        <div class="row q-col-gutter-md justify-center items-start">
          <!-- Componente 1: Dalle (Inizio) -->
          <div class="col-12 col-sm-6 text-center">
            <div class="text-subtitle2 text-weight-bold text-teal-9 q-mb-xs row items-center justify-center q-gutter-xs">
              <q-icon name="flight_takeoff" size="18px" />
              <span>{{ t('broadcasters.hoursDialog.startTime') }}</span>
            </div>
            <div class="q-time-wrapper shadow-1 rounded-borders overflow-hidden bg-white">
              <q-time
                v-model="slotForm.from"
                format24h
                mask="HH:mm"
                color="teal-7"
                flat
              />
            </div>
          </div>

          <!-- Componente 2: Alle (Fine) -->
          <div class="col-12 col-sm-6 text-center">
            <div class="text-subtitle2 text-weight-bold text-teal-9 q-mb-xs row items-center justify-center q-gutter-xs">
              <q-icon name="flight_land" size="18px" />
              <span>{{ t('broadcasters.hoursDialog.endTime') }}</span>
            </div>
            <div class="q-time-wrapper shadow-1 rounded-borders overflow-hidden bg-white">
              <q-time
                v-model="slotForm.to"
                format24h
                mask="HH:mm"
                color="teal-7"
                flat
              />
            </div>
          </div>
        </div>

        <!-- Suggerimenti Fasce Rapide -->
        <div class="q-mt-md">
          <div class="text-caption text-weight-medium text-grey-7 q-mb-xs">
            {{ t('broadcasters.hoursDialog.recommendedSlots') }}
          </div>
          <div class="row q-gutter-xs">
            <q-chip
              v-for="preset in timePresets"
              :key="preset.label"
              clickable
              outline
              size="sm"
              color="teal-8"
              class="bg-grey-1"
              @click="applyPreset(preset.from, preset.to)"
            >
              {{ preset.label }} ({{ preset.from }} - {{ preset.to }})
            </q-chip>
          </div>
        </div>
      </q-card-section>

      <q-separator style="flex-shrink: 0" />

      <!-- Azioni Popup Fascia (Al Piede) -->
      <q-card-actions align="right" class="q-pa-md bg-grey-1" style="flex-shrink: 0">
        <q-btn flat :label="t('actions.cancel')" v-close-popup />
        <q-btn
          color="teal-7"
          :icon="editingSlotIndex !== null ? 'check' : 'add'"
          :label="editingSlotIndex !== null ? t('broadcasters.hoursDialog.updateSlot') : t('broadcasters.hoursDialog.addSlot')"
          unelevated
          :disable="!isSlotValid"
          @click="confirmSlot"
        />
      </q-card-actions>
    </q-card>
  </q-dialog>

  <!-- ======================================================== -->
  <!-- DIALOG AGGIUNTA GIORNO PERSONALIZZATO                    -->
  <!-- ======================================================== -->
  <q-dialog v-model="showAddDayPrompt" persistent>
    <q-card style="width: 380px; max-width: 90vw">
      <q-card-section class="bg-teal-7 text-white q-py-sm">
        <div class="text-subtitle1 text-weight-bold">{{ t('broadcasters.hoursDialog.newDayTitle') }}</div>
      </q-card-section>
      <q-card-section class="q-pa-md">
        <q-select
          v-model="newDayName"
          :options="availableDayOptions"
          :label="t('broadcasters.hoursDialog.dayNameLabel')"
          outlined
          dense
          use-input
          new-value-mode="add-unique"
          :placeholder="t('broadcasters.hoursDialog.dayNamePlaceholder')"
        />
      </q-card-section>
      <q-card-actions align="right" class="q-pa-md bg-grey-1">
        <q-btn flat :label="t('actions.cancel')" v-close-popup />
        <q-btn
          color="teal-7"
          :label="t('broadcasters.hoursDialog.addDay')"
          unelevated
          :disable="!newDayName.trim()"
          @click="addNewDay"
        />
      </q-card-actions>
    </q-card>
  </q-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { Notify } from 'quasar';
import {
  useBroadcastersStore,
  type Broadcaster,
  type BroadcasterHours,
  type BroadcasterTimeSlot,
} from '@/stores/broadcasters-store';

const props = defineProps<{
  modelValue: boolean;
  broadcaster: Broadcaster | null;
}>();

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void;
  (e: 'saved', broadcaster: Broadcaster): void;
}>();

const { t } = useI18n();
const broadcastersStore = useBroadcastersStore();
const loading = ref(false);
const localError = ref<string | null>(null);

// Giorni standard della settimana
const standardDays = [
  'Lunedì',
  'Martedì',
  'Mercoledì',
  'Giovedì',
  'Venerdì',
  'Sabato',
  'Domenica',
];

// Presets per fasce orarie rapide
const timePresets = [
  { label: 'Mattina', from: '09:00', to: '13:00' },
  { label: 'Pomeriggio', from: '14:00', to: '18:00' },
  { label: 'Tardo Pom.', from: '16:00', to: '19:30' },
  { label: 'Serale', from: '20:30', to: '23:30' },
  { label: 'Notturno', from: '21:30', to: '23:59' },
];

// Struttura dati orari modificata: array di { day, hours: [] }
const hoursList = ref<BroadcasterHours[]>([]);

// Stato per popup aggiunta / modifica slot
const slotDialogOpen = ref(false);
const activeDayIndex = ref<number | null>(null);
const editingSlotIndex = ref<number | null>(null);
const slotForm = reactive<BroadcasterTimeSlot>({ from: '09:00', to: '13:00' });

// Stato per dialog aggiunta giorno
const showAddDayPrompt = ref(false);
const newDayName = ref('');

// Computed helper
const activeDay = computed(() => {
  if (activeDayIndex.value === null) return null;
  return hoursList.value[activeDayIndex.value] || null;
});

const totalConfiguredSlots = computed(() => {
  return hoursList.value.reduce((acc, d) => acc + (d.hours?.length || 0), 0);
});

const activeDaysCount = computed(() => {
  return hoursList.value.filter((d) => d.hours && d.hours.length > 0).length;
});

const availableDayOptions = computed(() => {
  const existingNames = new Set(hoursList.value.map((d) => d.day.toLowerCase()));
  return standardDays.filter((d) => !existingNames.has(d.toLowerCase()));
});

// ==========================================
// VALIDAZIONE SOVRAPPOSIZIONE ORARI
// ==========================================
const slotValidationError = computed<string | null>(() => {
  const from = (slotForm.from || '').trim();
  const to = (slotForm.to || '').trim();

  if (!from || !to) {
    return t('broadcasters.hoursDialog.selectBothTimes');
  }

  // Verifica che from < to
  if (from >= to) {
    return t('broadcasters.hoursDialog.startBeforeEnd', { from, to });
  }

  // Se non c'è un giorno attivo, nulla da confrontare
  if (activeDayIndex.value === null || !activeDay.value) {
    return null;
  }

  const existingSlots = activeDay.value.hours || [];
  for (let i = 0; i < existingSlots.length; i++) {
    // Se stiamo modificando uno slot esistente, ignoriamo se stesso
    if (editingSlotIndex.value !== null && i === editingSlotIndex.value) {
      continue;
    }

    const other = existingSlots[i];
    if (!other) continue;
    const otherFrom = (other.from || '').trim();
    const otherTo = (other.to || '').trim();

    if (!otherFrom || !otherTo) continue;

    // Condizione standard di sovrapposizione tra [from, to] e [otherFrom, otherTo]:
    // from < otherTo && otherFrom < to
    if (from < otherTo && otherFrom < to) {
      return t('broadcasters.hoursDialog.overlapDetected', {
        from,
        to,
        otherFrom,
        otherTo,
        day: activeDay.value.day,
      });
    }
  }

  return null;
});

const isSlotValid = computed(() => {
  return slotValidationError.value === null && !!slotForm.from && !!slotForm.to;
});

const slotDuration = computed(() => {
  return getSlotDurationText(slotForm.from, slotForm.to);
});

// ==========================================
// METODI APERTURA E GESTIONE POPUP
// ==========================================
function onShow() {
  localError.value = null;

  if (props.broadcaster?.hours && Array.isArray(props.broadcaster.hours) && props.broadcaster.hours.length > 0) {
    hoursList.value = props.broadcaster.hours.map((h: any) => {
      let slots: BroadcasterTimeSlot[] = [];
      if (Array.isArray(h.hours)) {
        slots = h.hours.map((s: any) => ({
          from: String(s.from || '').trim(),
          to: String(s.to || '').trim(),
        })).filter((s: BroadcasterTimeSlot) => s.from && s.to);
      } else if (typeof h.hours === 'string' && h.hours.includes('-')) {
        // Fallback per vecchi dati stringa tipo "21:00 - 23:00"
        const parts = h.hours.split('-');
        if (parts.length === 2) {
          slots.push({
            from: parts[0].trim(),
            to: parts[1].trim(),
          });
        }
      }
      // Ordina cronologicamente
      slots.sort((a, b) => a.from.localeCompare(b.from));

      return {
        day: h.day,
        hours: slots,
      };
    });
  } else {
    // Inizializza con i 7 giorni standard della settimana vuoti
    hoursList.value = standardDays.map((day) => ({
      day,
      hours: [],
    }));
  }
}

function closeDialog() {
  emit('update:modelValue', false);
}

function openAddSlot(dayIndex: number) {
  activeDayIndex.value = dayIndex;
  editingSlotIndex.value = null;

  // Calcola un orario suggerito che non si sovrappone agli esistenti
  const day = hoursList.value[dayIndex];
  const last = day && day.hours.length > 0 ? day.hours[day.hours.length - 1] : undefined;
  if (last && last.to && last.to < '22:00') {
    slotForm.from = last.to;
    const parts = last.to.split(':');
    const h = Number(parts[0] || '0');
    const m = Number(parts[1] || '0');
    const nextH = Math.min(23, h + 2);
    slotForm.to = `${String(nextH).padStart(2, '0')}:${String(m).padStart(2, '0')}`;
  } else {
    slotForm.from = '09:00';
    slotForm.to = '13:00';
  }

  slotDialogOpen.value = true;
}

function openEditSlot(dayIndex: number, slotIndex: number) {
  activeDayIndex.value = dayIndex;
  editingSlotIndex.value = slotIndex;

  const day = hoursList.value[dayIndex];
  if (day && day.hours[slotIndex]) {
    const s = day.hours[slotIndex];
    slotForm.from = s.from;
    slotForm.to = s.to;
  }

  slotDialogOpen.value = true;
}

function applyPreset(from: string, to: string) {
  slotForm.from = from;
  slotForm.to = to;
}

function confirmSlot() {
  if (!isSlotValid.value || activeDayIndex.value === null) return;
  const day = hoursList.value[activeDayIndex.value];
  if (!day) return;

  const newSlot: BroadcasterTimeSlot = {
    from: slotForm.from.trim(),
    to: slotForm.to.trim(),
  };

  if (editingSlotIndex.value !== null && editingSlotIndex.value >= 0) {
    day.hours[editingSlotIndex.value] = newSlot;
  } else {
    day.hours.push(newSlot);
  }

  // Ordina le fasce del giorno cronologicamente
  day.hours.sort((a, b) => a.from.localeCompare(b.from));

  slotDialogOpen.value = true;
  Notify.create({
    type: 'positive',
    message: t('broadcasters.hoursDialog.slotSavedToast', {
      from: newSlot.from,
      to: newSlot.to,
      day: day.day,
    }),
    timeout: 1500,
  });
  slotDialogOpen.value = false;
}

function removeSlot(dayIndex: number, slotIndex: number) {
  const day = hoursList.value[dayIndex];
  if (day && day.hours) {
    day.hours.splice(slotIndex, 1);
  }
}

function clearDayHours(dayIndex: number) {
  const day = hoursList.value[dayIndex];
  if (day) {
    day.hours = [];
  }
}

function isCustomDay(dayName: string): boolean {
  return !standardDays.includes(dayName);
}

function removeDay(dayIndex: number) {
  hoursList.value.splice(dayIndex, 1);
}

function addNewDay() {
  const d = newDayName.value.trim();
  if (!d) return;

  if (hoursList.value.some((x) => x.day.toLowerCase() === d.toLowerCase())) {
    Notify.create({
      type: 'warning',
      message: t('broadcasters.hoursDialog.dayAlreadyPresent', { day: d }),
    });
    return;
  }

  hoursList.value.push({ day: d, hours: [] });
  newDayName.value = '';
  showAddDayPrompt.value = false;
}

function copyDayHours(sourceDayIndex: number, target: 'weekdays' | 'weekend' | 'all') {
  const sourceDay = hoursList.value[sourceDayIndex];
  if (!sourceDay || sourceDay.hours.length === 0) return;

  const slotsToCopy = sourceDay.hours.map((s) => ({ ...s }));

  const weekdays = ['Lunedì', 'Martedì', 'Mercoledì', 'Giovedì', 'Venerdì'];
  const weekend = ['Sabato', 'Domenica'];

  let count = 0;
  hoursList.value.forEach((d, idx) => {
    if (idx === sourceDayIndex) return;

    let shouldCopy = false;
    if (target === 'weekdays' && weekdays.includes(d.day)) shouldCopy = true;
    if (target === 'weekend' && weekend.includes(d.day)) shouldCopy = true;
    if (target === 'all') shouldCopy = true;

    if (shouldCopy) {
      d.hours = slotsToCopy.map((s) => ({ ...s }));
      count++;
    }
  });

  Notify.create({
    type: 'positive',
    message: t('broadcasters.hoursDialog.copiedToast', {
      day: sourceDay.day,
      count,
    }),
    timeout: 2000,
  });
}

function getSlotDurationText(from: string, to: string): string {
  if (!from || !to || from >= to) return '';
  const parts1 = from.split(':');
  const parts2 = to.split(':');
  const h1 = Number(parts1[0]);
  const m1 = Number(parts1[1]);
  const h2 = Number(parts2[0]);
  const m2 = Number(parts2[1]);
  if (isNaN(h1) || isNaN(m1) || isNaN(h2) || isNaN(m2)) return '';

  const totalMin = h2 * 60 + m2 - (h1 * 60 + m1);
  if (totalMin <= 0) return '';

  const hours = Math.floor(totalMin / 60);
  const mins = totalMin % 60;
  if (hours > 0 && mins > 0) return `${hours}h ${mins}m`;
  if (hours > 0) return `${hours}h`;
  return `${mins}m`;
}

// ==========================================
// SALVATAGGIO FINALE AL BACKEND
// ==========================================
async function onSave() {
  if (!props.broadcaster?.id) return;

  // Validazione finale di integrità e controllo sovrapposizioni
  for (const d of hoursList.value) {
    const slots = d.hours || [];
    for (let i = 0; i < slots.length; i++) {
      const s1 = slots[i];
      if (!s1) continue;
      if (s1.from >= s1.to) {
        localError.value = t('broadcasters.hoursDialog.invalidTimeRange', {
          day: d.day,
          from: s1.from,
          to: s1.to,
        });
        return;
      }
      for (let j = i + 1; j < slots.length; j++) {
        const s2 = slots[j];
        if (!s2) continue;
        if (s1.from < s2.to && s2.from < s1.to) {
          localError.value = t('broadcasters.hoursDialog.overlapOnSave', {
            day: d.day,
            from: s1.from,
            to: s1.to,
            otherFrom: s2.from,
            otherTo: s2.to,
          });
          return;
        }
      }
    }
  }

  loading.value = true;
  localError.value = null;

  try {
    // Invia solo i giorni con fasce orarie configurate
    const cleanedPayload: BroadcasterHours[] = hoursList.value
      .filter((d) => d.day.trim().length > 0 && d.hours && d.hours.length > 0)
      .map((d) => ({
        day: d.day.trim(),
        hours: d.hours.map((h) => ({
          from: h.from.trim(),
          to: h.to.trim(),
        })),
      }));

    const updated = await broadcastersStore.updateHours(props.broadcaster.id, cleanedPayload);
    Notify.create({
      type: 'positive',
      message: t('broadcasters.hoursDialog.saveSuccess'),
    });
    emit('saved', updated);
    closeDialog();
  } catch (err: unknown) {
    localError.value = err instanceof Error ? err.message : String(err);
    Notify.create({
      type: 'negative',
      message: localError.value || t('broadcasters.hoursDialog.saveError'),
    });
  } finally {
    loading.value = false;
  }
}
</script>

<style scoped>
.day-card {
  border-radius: 8px;
  transition: all 0.2s ease;
}

.day-card--has-slots {
  border-color: #26a69a;
  background-color: #fafdfd;
}

.slot-chip {
  transition: transform 0.15s ease, box-shadow 0.15s ease;
}

.slot-chip:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.1);
}

.slot-text {
  font-weight: 600;
  letter-spacing: 0.3px;
}

.font-monospace {
  font-family: monospace;
}

.border-bottom {
  border-bottom: 1px solid #e0f2f1;
}

.border-negative {
  border: 1px solid #f44336;
}

.border-teal {
  border: 1px solid #26a69a;
}

.q-time-wrapper {
  border: 1px solid #e0e0e0;
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 4px;
}
</style>
