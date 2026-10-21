<template>
  <q-page class="q-pa-md">
    <div class="row items-center justify-between q-mb-md">
      <div>
        <h1 class="text-h4 q-my-none text-weight-bold">Pubblicità</h1>
        <p class="text-subtitle1 text-grey-7 q-mb-none">Gestione dell'elenco pubblicità (Advertising Directory)</p>
      </div>
      <q-btn
        color="primary"
        icon="add"
        label="Nuova Pubblicità"
        unelevated
        @click="openDialog"
      />
    </div>

    <!-- Banner errori globali -->
    <q-banner v-if="errorMessage" class="bg-negative text-white q-mb-md" rounded>
      <template v-slot:avatar>
        <q-icon name="error" color="white" />
      </template>
      {{ errorMessage }}
      <template v-slot:action>
        <q-btn flat color="white" label="Riprova" @click="fetchAdvertising" />
      </template>
    </q-banner>

    <!-- Tabella Quasar nativa -->
    <q-card flat bordered>
      <q-table
        :rows="advertisingList"
        :columns="columns"
        row-key="id"
        :loading="loading"
        :filter="searchFilter"
        no-data-label="Nessuna pubblicità presente"
        no-results-label="Nessun risultato corrispondente alla ricerca"
      >
        <template v-slot:top-right>
          <q-input
            v-model="searchFilter"
            dense
            outlined
            debounce="300"
            placeholder="Cerca pubblicità..."
            clearable
          >
            <template v-slot:append>
              <q-icon name="search" />
            </template>
          </q-input>
        </template>

        <!-- Slot custom per Status -->
        <template v-slot:body-cell-status="props">
          <q-td :props="props">
            <q-badge
              :color="props.value === 'active' ? 'positive' : 'grey-7'"
              :label="props.value"
              rounded
              class="q-px-sm q-py-xs text-capitalize text-weight-medium"
            />
          </q-td>
        </template>

        <!-- Slot custom per Email -->
        <template v-slot:body-cell-email="props">
          <q-td :props="props">
            <a :href="'mailto:' + props.value" class="text-primary text-weight-medium" style="text-decoration: none">
              {{ props.value }}
            </a>
          </q-td>
        </template>
      </q-table>
    </q-card>

    <!-- Sezione Anteprima Promozioni Settimanali (getWeekPromo) -->
    <q-card flat bordered class="q-mt-lg bg-white">
      <q-card-section class="row items-center justify-between q-pb-sm">
        <div>
          <div class="row items-center q-gutter-sm">
            <q-icon name="campaign" size="28px" color="primary" />
            <div class="text-h6 text-weight-bold">Anteprima Promozioni per Settimana (<code>getWeekPromo</code>)</div>
          </div>
          <div class="text-caption text-grey-7 q-mt-xs">
            Visualizza il payload JSON generato dall'endpoint per i tre tipi di promozioni ordinati per sezione
          </div>
        </div>

        <!-- Badges di riepilogo conteggi -->
        <div v-if="weekPromoData" class="row items-center q-gutter-xs">
          <q-badge color="primary" class="q-px-sm q-py-xs">
            Promo Left: {{ weekPromoData.promo_left?.length || 0 }}
          </q-badge>
          <q-badge color="teal" class="q-px-sm q-py-xs">
            Promo Right 1: {{ weekPromoData.promo_right1?.length || 0 }}
          </q-badge>
          <q-badge color="deep-purple" class="q-px-sm q-py-xs">
            Promo Right 2: {{ weekPromoData.promo_right2?.length || 0 }}
          </q-badge>
        </div>
      </q-card-section>

      <q-separator />

      <!-- Barra di controllo e selezione settimana/anno -->
      <q-card-section class="bg-grey-1 q-py-sm">
        <div class="row items-center justify-between q-col-gutter-md">
          <div class="col-12 col-sm-auto row items-center q-gutter-sm">
            <!-- Selettore Anno -->
            <q-input
              v-model.number="selectedYear"
              type="number"
              label="Anno"
              outlined
              dense
              style="width: 100px"
            />

            <!-- Selettore Settimana -->
            <q-select
              v-model="selectedWeek"
              :options="weekOptions"
              emit-value
              map-options
              outlined
              dense
              label="Settimana (1-53)"
              style="min-width: 220px"
            >
              <template #prepend>
                <q-icon name="calendar_today" />
              </template>
            </q-select>

            <!-- Pulsanti navigazione rapida -->
            <div class="row items-center q-gutter-xs">
              <q-btn
                flat
                round
                dense
                icon="chevron_left"
                :disable="selectedWeek <= 1"
                @click="selectedWeek--"
              >
                <q-tooltip>Settimana precedente</q-tooltip>
              </q-btn>
              <q-btn
                outline
                dense
                color="primary"
                label="Oggi"
                class="q-px-sm"
                @click="selectedWeek = currentIsoWeek; selectedYear = currentIsoYear"
              >
                <q-tooltip>Vai alla settimana corrente (W{{ currentIsoWeek }})</q-tooltip>
              </q-btn>
              <q-btn
                flat
                round
                dense
                icon="chevron_right"
                :disable="selectedWeek >= 53"
                @click="selectedWeek++"
              >
                <q-tooltip>Settimana successiva</q-tooltip>
              </q-btn>
            </div>
          </div>

          <div class="col-12 col-sm-auto row items-center q-gutter-sm">
            <q-btn
              color="primary"
              icon="refresh"
              label="Ricarica"
              flat
              dense
              :loading="weekPromoLoading"
              @click="fetchWeekPromo"
            />
            <q-btn
              color="secondary"
              icon="content_copy"
              label="Copia JSON"
              unelevated
              dense
              class="q-px-sm"
              :disable="!weekPromoData || weekPromoLoading"
              @click="copyJsonToClipboard"
            />
          </div>
        </div>
      </q-card-section>

      <!-- Visualizzatore JSON -->
      <q-card-section class="q-pa-none">
        <div v-if="weekPromoLoading" class="row justify-center items-center q-pa-xl bg-grey-2">
          <q-spinner-dots color="primary" size="40px" />
          <span class="q-ml-sm text-grey-8">Caricamento JSON promozioni settimana W{{ selectedWeek }}...</span>
        </div>

        <q-banner v-else-if="weekPromoError" class="bg-negative text-white q-ma-md" rounded>
          <template #avatar>
            <q-icon name="error_outline" color="white" />
          </template>
          {{ weekPromoError }}
          <template #action>
            <q-btn flat color="white" label="Riprova" @click="fetchWeekPromo" />
          </template>
        </q-banner>

        <div v-else-if="weekPromoData" class="json-viewer-container">
          <div class="row items-center justify-between q-px-md q-py-xs bg-dark text-white text-caption">
            <span>GET /promo/week/{{ selectedWeek }}?year={{ selectedYear }}</span>
            <span class="text-grey-4">Totale: {{ (weekPromoData.promo_left?.length || 0) + (weekPromoData.promo_right1?.length || 0) + (weekPromoData.promo_right2?.length || 0) }} promo attive</span>
          </div>
          <pre class="json-code-block q-pa-md q-ma-none scroll">{{ formattedJson }}</pre>
        </div>

        <div v-else class="text-center q-pa-lg text-grey-6 text-italic">
          Nessun dato disponibile. Seleziona una settimana per visualizzare l'anteprima.
        </div>
      </q-card-section>
    </q-card>

    <!-- Dialog modale per inserimento -->
    <q-dialog v-model="dialogOpen" persistent>
      <q-card style="min-width: 420px; max-width: 90vw">
        <q-card-section class="row items-center q-pb-none">
          <div class="text-h6 text-weight-bold">Nuova Pubblicità</div>
          <q-space />
          <q-btn icon="close" flat round dense v-close-popup :disable="submitting" />
        </q-card-section>

        <q-card-section>
          <q-form @submit="onSubmit" class="q-gutter-md">
            <q-input
              v-model="form.nome"
              label="Nome *"
              outlined
              dense
              autofocus
              :rules="[
                (val: string) => (!!val && val.trim().length > 0) || 'Il nome è obbligatorio',
                (val: string) => val.trim().length >= 2 || 'Il nome deve contenere almeno 2 caratteri'
              ]"
            />

            <q-input
              v-model="form.email"
              label="Email *"
              type="email"
              outlined
              dense
              :rules="[
                (val: string) => (!!val && val.trim().length > 0) || 'L\'email è obbligatoria',
                (val: string) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(val.trim()) || 'Inserisci un indirizzo email valido'
              ]"
            />

            <q-select
              v-model="form.status"
              :options="statusOptions"
              label="Stato *"
              emit-value
              map-options
              outlined
              dense
            />

            <div class="row justify-end q-gutter-sm q-pt-sm">
              <q-btn flat label="Annulla" v-close-popup :disable="submitting" />
              <q-btn
                unelevated
                color="primary"
                type="submit"
                label="Salva"
                :loading="submitting"
              />
            </div>
          </q-form>
        </q-card-section>
      </q-card>
    </q-dialog>
  </q-page>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted } from 'vue';
import type { QTableColumn } from 'quasar';
import { Notify } from 'quasar';
import Client, { Local, advertising as AdvertisingNS } from '@/encore/client';

// Tipizzazione forte basata sulle definizioni generate da Encore
type Advertising = AdvertisingNS.Advertising;
type CreateAdvertisingParams = AdvertisingNS.CreateAdvertisingParams;
type Status = AdvertisingNS.Status;
type WeekPromoResponse = AdvertisingNS.WeekPromoResponse;

const client = new Client(Local);

const advertisingList = ref<Advertising[]>([]);
const loading = ref(false);
const submitting = ref(false);
const errorMessage = ref<string | null>(null);
const searchFilter = ref('');
const dialogOpen = ref(false);

const form = reactive<CreateAdvertisingParams>({
  nome: '',
  email: '',
  status: 'active',
});

const statusOptions = [
  { label: 'Attivo (active)', value: 'active' },
  { label: 'Inattivo (inactive)', value: 'inactive' },
];

const columns: QTableColumn<Advertising>[] = [
  {
    name: 'nome',
    label: 'Nome',
    field: (row: Advertising) => row.nome,
    align: 'left',
    sortable: true,
  },
  {
    name: 'email',
    label: 'Email',
    field: (row: Advertising) => row.email,
    align: 'left',
    sortable: true,
  },
  {
    name: 'status',
    label: 'Stato',
    field: (row: Advertising) => row.status,
    align: 'center',
    sortable: true,
  },
  {
    name: 'id',
    label: 'ID (UUID)',
    field: (row: Advertising) => row.id,
    align: 'left',
    sortable: true,
    style: 'font-family: monospace; font-size: 0.85em; color: #666;',
  },
];

async function fetchAdvertising() {
  loading.value = true;
  errorMessage.value = null;
  try {
    const res = await client.advertising.List();
    advertisingList.value = res.advertising ?? [];
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err);
    errorMessage.value = `Impossibile caricare le pubblicità: ${msg}`;
  } finally {
    loading.value = false;
  }
}

function openDialog() {
  form.nome = '';
  form.email = '';
  form.status = 'active';
  dialogOpen.value = true;
}

async function onSubmit() {
  submitting.value = true;
  try {
    const payload: CreateAdvertisingParams = {
      nome: form.nome.trim(),
      email: form.email.trim(),
      status: form.status,
    };
    const created = await client.advertising.Create(payload);
    advertisingList.value.unshift(created);
    dialogOpen.value = false;
    Notify.create({
      type: 'positive',
      message: `Pubblicità "${created.nome}" creata con successo!`,
      position: 'top-right',
    });
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err);
    Notify.create({
      type: 'negative',
      message: `Errore durante la creazione: ${msg}`,
      position: 'top-right',
    });
  } finally {
    submitting.value = false;
  }
}

function getCurrentISOWeek(): number {
  const date = new Date();
  const target = new Date(date.valueOf());
  const dayNr = (date.getDay() + 6) % 7;
  target.setDate(target.getDate() - dayNr + 3);
  const firstThursday = target.valueOf();
  target.setMonth(0, 1);
  if (target.getDay() !== 4) {
    target.setMonth(0, 1 + ((4 - target.getDay()) + 7) % 7);
  }
  return 1 + Math.ceil((firstThursday - target.valueOf()) / 604800000);
}

const currentIsoWeek = getCurrentISOWeek();
const currentIsoYear = new Date().getFullYear();

const selectedWeek = ref<number>(currentIsoWeek);
const selectedYear = ref<number>(currentIsoYear);
const weekPromoData = ref<WeekPromoResponse | null>(null);
const weekPromoLoading = ref<boolean>(false);
const weekPromoError = ref<string | null>(null);

const weekOptions = computed(() => {
  const opts = [];
  for (let w = 1; w <= 53; w++) {
    opts.push({
      label: `Settimana ${w}${w === currentIsoWeek && selectedYear.value === currentIsoYear ? ' (Corrente)' : ''}`,
      value: w,
    });
  }
  return opts;
});

const formattedJson = computed(() => {
  if (!weekPromoData.value) return '';
  return JSON.stringify(weekPromoData.value, null, 2);
});

async function fetchWeekPromo() {
  if (selectedWeek.value < 1 || selectedWeek.value > 53) return;
  weekPromoLoading.value = true;
  weekPromoError.value = null;
  try {
    const res = await client.advertising.GetWeekPromo(selectedWeek.value, {
      year: selectedYear.value,
    });
    weekPromoData.value = res;
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err);
    weekPromoError.value = `Errore durante il recupero delle promozioni per la settimana ${selectedWeek.value}: ${msg}`;
    weekPromoData.value = null;
  } finally {
    weekPromoLoading.value = false;
  }
}

async function copyJsonToClipboard() {
  if (!weekPromoData.value) return;
  try {
    await navigator.clipboard.writeText(JSON.stringify(weekPromoData.value, null, 2));
    Notify.create({
      type: 'positive',
      message: 'JSON copiato negli appunti!',
      position: 'top-right',
    });
  } catch {
    Notify.create({
      type: 'warning',
      message: 'Impossibile copiare negli appunti.',
      position: 'top-right',
    });
  }
}

watch([selectedWeek, selectedYear], () => {
  void fetchWeekPromo();
});

onMounted(() => {
  void fetchAdvertising();
  void fetchWeekPromo();
});
</script>

<style scoped>
.json-viewer-container {
  border-top: 1px solid #e2e8f0;
}

.json-code-block {
  background-color: #0f172a;
  color: #38bdf8;
  font-family: 'JetBrains Mono', 'Fira Code', 'Roboto Mono', Menlo, Monaco, Consolas, monospace;
  font-size: 13px;
  line-height: 1.5;
  max-height: 500px;
  overflow-y: auto;
  border-bottom-left-radius: 4px;
  border-bottom-right-radius: 4px;
}
</style>
