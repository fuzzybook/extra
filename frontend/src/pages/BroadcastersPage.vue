<template>
  <q-page class="q-pa-md bg-grey-1">
    <!-- Header Pagina -->
    <div class="row items-center justify-between q-mb-md">
      <div>
        <h1 class="text-h4 q-my-none text-weight-bold text-grey-9">{{ t('broadcasters.title') }}</h1>
        <p class="text-subtitle1 text-grey-7 q-mb-none">
          {{ t('broadcasters.subtitle') }}
        </p>
      </div>

      <div class="row items-center q-gutter-sm">
        <q-btn-toggle
          v-model="viewMode"
          toggle-color="primary"
          flat
          dense
          :options="[
            { icon: 'grid_view', value: 'cards' },
            { icon: 'view_list', value: 'table' },
          ]"
        >
          <q-tooltip>{{ t('broadcasters.switchView') }}</q-tooltip>
        </q-btn-toggle>

        <q-btn
          flat
          round
          dense
          icon="refresh"
          color="grey-8"
          :loading="broadcastersStore.loading"
          @click="loadData"
        >
          <q-tooltip>{{ t('broadcasters.reloadList') }}</q-tooltip>
        </q-btn>
      </div>
    </div>

    <!-- Banner errori globali -->
    <q-banner v-if="broadcastersStore.error" class="bg-negative text-white q-mb-md" rounded>
      <template #avatar>
        <q-icon name="error" color="white" />
      </template>
      {{ broadcastersStore.error }}
      <template #action>
        <q-btn flat color="white" :label="t('broadcasters.retry')" @click="loadData" />
      </template>
    </q-banner>

    <!-- Barra Ricerca e Filtri -->
    <q-card flat bordered class="q-mb-md bg-white">
      <q-card-section class="q-py-sm">
        <div class="row q-col-gutter-sm items-center">
          <div class="col-12 col-sm-4">
            <q-input
              v-model="searchQuery"
              :placeholder="t('broadcasters.searchPlaceholder')"
              outlined
              dense
              clearable
            >
              <template #prepend>
                <q-icon name="search" />
              </template>
            </q-input>
          </div>

          <div class="col-12 col-sm-3">
            <q-select
              v-model="categoryFilter"
              :options="categoryFilterOptions"
              :label="t('broadcasters.filterCategory')"
              outlined
              dense
              clearable
            >
             
              <template #option="scope">
                <q-item v-bind="scope.itemProps">
                  <q-item-section avatar side class="q-pr-sm">
                    <div
                      class="filter-color-dot"
                      :style="{ backgroundColor: getCategoryColor(scope.opt) }"
                    />
                  </q-item-section>
                  <q-item-section>
                    <q-item-label>{{ scope.opt }}</q-item-label>
                  </q-item-section>
                </q-item>
              </template>
            </q-select>
          </div>

          <div class="col-12 col-sm-3">
            <q-select
              v-model="availabilityFilter"
              :options="availabilityFilterOptions"
              emit-value
              map-options
              :label="t('broadcasters.filterAvailability')"
              outlined
              dense
              clearable
            >
             
            </q-select>
          </div>

          <div class="col-12 col-sm-2">
            <q-checkbox
              v-model="certifiedOnly"
              :label="t('broadcasters.certifiedOnly')"
              dense
              color="positive"
            />
          </div>
        </div>
      </q-card-section>
    </q-card>

    <!-- Stato di caricamento -->
    <div v-if="broadcastersStore.loading && filteredBroadcasters.length === 0" class="row justify-center q-pa-xl">
      <q-spinner color="primary" size="48px" />
    </div>

    <!-- Stato vuoto -->
    <div
      v-else-if="filteredBroadcasters.length === 0"
      class="text-center q-pa-xl bg-white rounded-borders bordered text-grey-6"
    >
      <q-icon name="podcasts" size="56px" class="q-mb-sm text-grey-4" />
      <div class="text-h6 text-grey-8">{{ t('broadcasters.noBroadcastersFound') }}</div>
      <div class="text-body2 text-grey-6">
        {{ t('broadcasters.noBroadcastersMatch') }}
      </div>
    </div>

    <!-- Vista a Schede (Cards) -->
    <div v-else-if="viewMode === 'cards'" class="row q-col-gutter-md">
      <div
        v-for="b in filteredBroadcasters"
        :key="b.id"
        class="col-12 col-sm-6 col-md-4 col-lg-3"
      >
        <q-card flat bordered class="broadcaster-card column full-height bg-white">
          <!-- Area Immagine Copertina con Aspect Ratio 2:3 proporzionato -->
          <div class="card-cover-wrapper">
            <img
              v-if="hasPhoto(b)"
              :src="getPrimaryPhoto(b)"
              :alt="b.display_name"
              class="card-cover-img"
              loading="lazy"
              @error="onImgError($event, b)"
            />
            <div v-else class="card-cover-placeholder flex flex-center">
              <q-icon name="photo_camera" size="56px" class="camera-icon-dark" />
            </div>
            <div class="card-cover-overlay">
              <!-- Badge Categoria -->
              <q-badge
                :style="{ backgroundColor: getCategoryColor(b.category) }"
                class="text-weight-bold shadow-1"
              >
                {{ b.category || t('broadcasters.generalCategory') }}
              </q-badge>

              <!-- Badge Disponibilità -->
              <q-badge :color="getAvailabilityBadge(b.availability).color">
                {{ getAvailabilityBadge(b.availability).label }}
              </q-badge>
            </div>

            <!-- Pulsante Menu Azioni in alto a destra -->
            <div class="card-menu-btn">
              <q-btn
                round
                dense
                color="white"
                text-color="dark"
                icon="more_vert"
                class="shadow-2"
              >
                <q-menu cover anchor="bottom right" self="top right">
                  <q-list style="min-width: 220px">
                    <q-item-label header class="text-weight-bold text-caption text-grey-7">
                      {{ t('broadcasters.manageBroadcaster') }}
                    </q-item-label>

                    <!-- Edit Profilo Principale -->
                    <q-item clickable v-close-popup @click="openEditBroadcaster(b)">
                      <q-item-section avatar>
                        <q-icon name="edit" color="primary" />
                      </q-item-section>
                      <q-item-section>
                        <q-item-label>{{ t('broadcasters.editProfile') }}</q-item-label>
                        <q-item-label caption>{{ t('broadcasters.editProfileCaption') }}</q-item-label>
                      </q-item-section>
                    </q-item>

                    <!-- Edit Attributi -->
                    <q-item clickable v-close-popup @click="openEditAttributes(b)">
                      <q-item-section avatar>
                        <q-icon name="tune" color="indigo" />
                      </q-item-section>
                      <q-item-section>
                        <q-item-label>{{ t('broadcasters.editAttributes') }}</q-item-label>
                        <q-item-label caption>{{ t('broadcasters.editAttributesCaption', { count: b.attributes?.length || 0 }) }}</q-item-label>
                      </q-item-section>
                    </q-item>

                    <!-- Edit Orari -->
                    <q-item clickable v-close-popup @click="openEditHours(b)">
                      <q-item-section avatar>
                        <q-icon name="schedule" color="teal" />
                      </q-item-section>
                      <q-item-section>
                        <q-item-label>{{ t('broadcasters.editHours') }}</q-item-label>
                        <q-item-label caption>{{ t('broadcasters.editHoursCaption', { count: getTotalTimeSlots(b) }) }}</q-item-label>
                      </q-item-section>
                    </q-item>

                    <!-- Edit Foto -->
                    <q-item clickable v-close-popup @click="openEditPhotos(b)">
                      <q-item-section avatar>
                        <q-icon name="photo_library" color="deep-orange" />
                      </q-item-section>
                      <q-item-section>
                        <q-item-label>{{ t('broadcasters.editPhotos') }}</q-item-label>
                        <q-item-label caption>{{ t('broadcasters.editPhotosCaption', { count: b.photos?.length || 0 }) }}</q-item-label>
                      </q-item-section>
                    </q-item>

                    <q-separator />

                    <!-- Elimina -->
                    <q-item clickable v-close-popup class="text-negative" @click="confirmDelete(b)">
                      <q-item-section avatar>
                        <q-icon name="delete" color="negative" />
                      </q-item-section>
                      <q-item-section>
                        <q-item-label>{{ t('broadcasters.delete') }}</q-item-label>
                      </q-item-section>
                    </q-item>
                  </q-list>
                </q-menu>
              </q-btn>
            </div>
          </div>

          <!-- Contenuto Card -->
          <q-card-section class="col q-pa-sm">
            <div class="row items-center no-wrap justify-between q-mb-xs">
              <div class="row items-center no-wrap text-subtitle1 text-weight-bold text-grey-9 ellipsis">
                <!-- Icone di stato: Certificato, Abilitato, In uso -->
                <div class="row items-center no-wrap q-gutter-xs q-mr-sm" style="flex-shrink: 0;">
                  <!-- Certificato -->
                  <q-icon
                    name="verified"
                    :color="b.certified ? 'positive' : 'grey-4'"
                    size="18px"
                  >
                    <q-tooltip>{{ b.certified ? t('broadcasters.certified') : t('broadcasters.notCertified') }}</q-tooltip>
                  </q-icon>

                  <!-- Abilitato -->
                  <q-icon
                    :name="b.enabled ? 'check_circle' : 'cancel'"
                    :color="b.enabled ? 'positive' : 'grey-4'"
                    size="18px"
                  >
                    <q-tooltip>{{ b.enabled ? t('broadcasters.enabled') : t('broadcasters.notEnabled') }}</q-tooltip>
                  </q-icon>

                  <!-- In uso -->
                  <q-icon
                    :name="b.used ? 'how_to_reg' : 'person_outline'"
                    :color="b.used ? 'primary' : 'grey-4'"
                    size="18px"
                  >
                    <q-tooltip>{{ b.used ? t('broadcasters.used') : t('broadcasters.notUsed') }}</q-tooltip>
                  </q-icon>
                </div>

                <span class="ellipsis">{{ b.display_name }}</span>
              </div>
            </div>

            <!-- Bio -->
            <p v-if="b.bio" class="text-caption text-grey-7 line-clamp-2 q-mb-xs">
              {{ b.bio }}
            </p>
            <p v-else class="text-caption text-grey-5 italic q-mb-xs">
              {{ t('broadcasters.noBioSpecified') }}
            </p>

            <!-- Servizi -->
            <div v-if="b.services && b.services.length > 0" class="row q-gutter-xs q-mb-xs">
              <q-badge
                v-for="serv in b.services.slice(0, 3)"
                :key="serv"
                outline
                color="primary"
                class="text-caption"
              >
                {{ serv }}
              </q-badge>
              <q-badge
                v-if="b.services.length > 3"
                outline
                color="grey-7"
                class="text-caption"
              >
                +{{ b.services.length - 3 }}
              </q-badge>
            </div>
          </q-card-section>

          <!-- Footer Card con Indicatori e Quick Buttons -->
          <q-separator />
          <q-card-actions align="between" class="q-pa-xs bg-grey-1">
            <div class="row items-center q-gutter-xs text-caption text-grey-7">
              <span class="row items-center q-mr-xs" :title="t('broadcasters.photosCount', { count: b.photos?.length || 0 })">
                <q-icon name="photo" size="14px" class="q-mr-xs text-deep-orange" />
                {{ b.photos?.length || 0 }}
              </span>
              <span class="row items-center q-mr-xs" :title="t('broadcasters.attributesCount', { count: b.attributes?.length || 0 })">
                <q-icon name="tune" size="14px" class="q-mr-xs text-indigo" />
                {{ b.attributes?.length || 0 }}
              </span>
              <span class="row items-center" :title="t('broadcasters.timeSlotsCount', { count: getTotalTimeSlots(b) })">
                <q-icon name="schedule" size="14px" class="q-mr-xs text-teal" />
                {{ getTotalTimeSlots(b) }}
              </span>
            </div>

            <div class="row q-gutter-none">
              <!-- Menu rapido button -->
              <q-btn
                flat
                dense
                round
                size="sm"
                icon="edit"
                color="primary"
                @click="openEditBroadcaster(b)"
              >
                <q-tooltip>{{ t('broadcasters.editProfile') }}</q-tooltip>
              </q-btn>
              <q-btn
                flat
                dense
                round
                size="sm"
                icon="photo_library"
                color="deep-orange"
                @click="openEditPhotos(b)"
              >
                <q-tooltip>{{ t('broadcasters.managePhotos') }}</q-tooltip>
              </q-btn>
            </div>
          </q-card-actions>
        </q-card>
      </div>
    </div>

    <!-- Vista Tabellare -->
    <div v-else class="bg-white rounded-borders shadow-1">
      <q-table
        :rows="filteredBroadcasters"
        :columns="tableColumns"
        row-key="id"
        flat
        bordered
        :pagination="{ rowsPerPage: 15 }"
      >
        <!-- Colonna Foto -->
        <template #body-cell-cover="props">
          <q-td :props="props" auto-width>
            <div class="table-cover-box">
              <img
                v-if="hasPhoto(props.row)"
                :src="getPrimaryPhoto(props.row)"
                :alt="props.row.display_name"
                class="table-cover-img"
                @error="onImgError($event, props.row)"
              />
              <div v-else class="table-cover-placeholder flex flex-center">
                <q-icon name="photo_camera" size="24px" class="camera-icon-dark" />
              </div>
            </div>
          </q-td>
        </template>

        <!-- Colonna Nome e ID -->
        <template #body-cell-display_name="props">
          <q-td :props="props">
            <div class="row items-center no-wrap text-weight-bold text-grey-9">
              <!-- Icone di stato: Certificato, Abilitato, In uso -->
              <div class="row items-center no-wrap q-gutter-xs q-mr-sm" style="flex-shrink: 0;">
                <!-- Certificato -->
                <q-icon
                  name="verified"
                  :color="props.row.certified ? 'positive' : 'grey-4'"
                  size="16px"
                >
                  <q-tooltip>{{ props.row.certified ? t('broadcasters.certified') : t('broadcasters.notCertified') }}</q-tooltip>
                </q-icon>

                <!-- Abilitato -->
                <q-icon
                  :name="props.row.enabled ? 'check_circle' : 'cancel'"
                  :color="props.row.enabled ? 'positive' : 'grey-4'"
                  size="16px"
                >
                  <q-tooltip>{{ props.row.enabled ? t('broadcasters.enabled') : t('broadcasters.notEnabled') }}</q-tooltip>
                </q-icon>

                <!-- In uso -->
                <q-icon
                  :name="props.row.used ? 'how_to_reg' : 'person_outline'"
                  :color="props.row.used ? 'primary' : 'grey-4'"
                  size="16px"
                >
                  <q-tooltip>{{ props.row.used ? t('broadcasters.used') : t('broadcasters.notUsed') }}</q-tooltip>
                </q-icon>
              </div>

              <span>{{ props.row.display_name }}</span>
            </div>
            <div class="text-caption text-grey-6">ID #{{ props.row.id }}</div>
          </q-td>
        </template>

        <!-- Colonna Categoria -->
        <template #body-cell-category="props">
          <q-td :props="props">
            <q-badge
              :style="{ backgroundColor: getCategoryColor(props.row.category) }"
              class="text-weight-bold"
            >
              {{ props.row.category }}
            </q-badge>
          </q-td>
        </template>

        <!-- Colonna Disponibilità -->
        <template #body-cell-availability="props">
          <q-td :props="props">
            <q-badge :color="getAvailabilityBadge(props.row.availability).color">
              {{ getAvailabilityBadge(props.row.availability).label }}
            </q-badge>
          </q-td>
        </template>

        <!-- Colonna Attributi & Orari -->
        <template #body-cell-meta="props">
          <q-td :props="props">
            <div class="text-caption text-grey-8">
              <span>{{ t('broadcasters.attributesCount', { count: props.row.attributes?.length || 0 }) }}</span> •
              <span>{{ t('broadcasters.timeSlotsCount', { count: getTotalTimeSlots(props.row) }) }}</span> •
              <span>{{ t('broadcasters.photosCount', { count: props.row.photos?.length || 0 }) }}</span>
            </div>
          </q-td>
        </template>

        <!-- Colonna Azioni con Menu -->
        <template #body-cell-actions="props">
          <q-td :props="props" auto-width>
            <q-btn flat round dense icon="more_vert" color="grey-8">
              <q-menu cover anchor="bottom right" self="top right">
                <q-list style="min-width: 200px">
                  <q-item clickable v-close-popup @click="openEditBroadcaster(props.row)">
                    <q-item-section avatar>
                      <q-icon name="edit" color="primary" />
                    </q-item-section>
                    <q-item-section>{{ t('broadcasters.editProfile') }}</q-item-section>
                  </q-item>

                  <q-item clickable v-close-popup @click="openEditAttributes(props.row)">
                    <q-item-section avatar>
                      <q-icon name="tune" color="indigo" />
                    </q-item-section>
                    <q-item-section>{{ t('broadcasters.editAttributes') }}</q-item-section>
                  </q-item>

                  <q-item clickable v-close-popup @click="openEditHours(props.row)">
                    <q-item-section avatar>
                      <q-icon name="schedule" color="teal" />
                    </q-item-section>
                    <q-item-section>{{ t('broadcasters.editHours') }}</q-item-section>
                  </q-item>

                  <q-item clickable v-close-popup @click="openEditPhotos(props.row)">
                    <q-item-section avatar>
                      <q-icon name="photo_library" color="deep-orange" />
                    </q-item-section>
                    <q-item-section>{{ t('broadcasters.editPhotos') }}</q-item-section>
                  </q-item>

                  <q-separator />

                  <q-item clickable v-close-popup class="text-negative" @click="confirmDelete(props.row)">
                    <q-item-section avatar>
                      <q-icon name="delete" color="negative" />
                    </q-item-section>
                    <q-item-section>{{ t('broadcasters.delete') }}</q-item-section>
                  </q-item>
                </q-list>
              </q-menu>
            </q-btn>
          </q-td>
        </template>
      </q-table>
    </div>

    <!-- Dialogs di Gestione Separatamente -->
    <EditBroadcasterDialog
      v-model="editBroadcasterOpen"
      :broadcaster="selectedBroadcaster"
      @saved="onSaved"
    />

    <EditAttributesDialog
      v-model="editAttributesOpen"
      :broadcaster="selectedBroadcaster"
      @saved="onSaved"
    />

    <EditHoursDialog
      v-model="editHoursOpen"
      :broadcaster="selectedBroadcaster"
      @saved="onSaved"
    />

    <EditPhotosDialog
      v-model="editPhotosOpen"
      :broadcaster="selectedBroadcaster"
      @saved="onSaved"
    />

    <!-- Dialog Conferma Cancellazione -->
    <q-dialog v-model="deleteConfirmOpen" persistent>
      <q-card style="width: 400px; max-width: 90vw">
        <q-card-section class="row items-center q-pb-none">
          <q-avatar icon="warning" color="negative" text-color="white" />
          <span class="q-ml-sm text-h6 text-weight-bold">{{ t('broadcasters.deleteConfirmTitle') }}</span>
        </q-card-section>

        <q-card-section class="text-body2 text-grey-8">
          {{ t('broadcasters.deleteConfirmMessage', { name: selectedBroadcaster?.display_name || '', id: selectedBroadcaster?.id || '' }) }}
        </q-card-section>

        <q-card-actions align="right">
          <q-btn flat :label="t('actions.cancel')" v-close-popup />
          <q-btn
            color="negative"
            :label="t('broadcasters.delete')"
            unelevated
            :loading="broadcastersStore.loading"
            @click="executeDelete"
          />
        </q-card-actions>
      </q-card>
    </q-dialog>
  </q-page>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { type QTableColumn, Notify } from 'quasar';
import { useBroadcastersStore, type Broadcaster } from '@/stores/broadcasters-store';
import EditBroadcasterDialog from '@/components/broadcasters/EditBroadcasterDialog.vue';
import EditAttributesDialog from '@/components/broadcasters/EditAttributesDialog.vue';
import EditHoursDialog from '@/components/broadcasters/EditHoursDialog.vue';
import EditPhotosDialog from '@/components/broadcasters/EditPhotosDialog.vue';

const { t } = useI18n();
const broadcastersStore = useBroadcastersStore();

const viewMode = ref<'cards' | 'table'>('cards');
const searchQuery = ref('');
const categoryFilter = ref<string | null>(null);
const availabilityFilter = ref<string | null>(null);
const certifiedOnly = ref(false);

const selectedBroadcaster = ref<Broadcaster | null>(null);
const editBroadcasterOpen = ref(false);
const editAttributesOpen = ref(false);
const editHoursOpen = ref(false);
const editPhotosOpen = ref(false);
const deleteConfirmOpen = ref(false);

const categoryFilterOptions = computed(() => {
  if (broadcastersStore.categories && broadcastersStore.categories.length > 0) {
    return broadcastersStore.categories.map((c) => c.name);
  }
  return ['General', 'Music', 'Beauty', 'test'];
});

function getCategoryColor(catName: string): string {
  const found = broadcastersStore.categories.find(
    (c) => c.name.toLowerCase() === (catName || '').toLowerCase()
  );
  return found?.color || '#6366f1';
}

const availabilityFilterOptions = computed(() => {
  const options = (broadcastersStore.availabilityOptions && broadcastersStore.availabilityOptions.length > 0)
    ? broadcastersStore.availabilityOptions
    : ['available', 'live', 'busy', 'offline'];

  return options.map((opt) => {
    const key = `broadcasters.availability.${opt.toLowerCase()}`;
    return {
      label: t(key, opt),
      value: opt,
    };
  });
});

const tableColumns = computed<QTableColumn[]>(() => [
  { name: 'cover', label: t('broadcasters.table.cover'), field: 'cover', align: 'center' },
  { name: 'display_name', label: t('broadcasters.table.displayName'), field: 'display_name', align: 'left', sortable: true },
  { name: 'category', label: t('broadcasters.table.category'), field: 'category', align: 'left', sortable: true },
  { name: 'availability', label: t('broadcasters.table.availability'), field: 'availability', align: 'left', sortable: true },
  { name: 'meta', label: t('broadcasters.table.details'), field: 'meta', align: 'left' },
  { name: 'actions', label: t('broadcasters.table.actions'), field: 'actions', align: 'right' },
]);

const failedImages = ref<Set<string>>(new Set());

onMounted(async () => {
  await loadData();
});

async function loadData() {
  failedImages.value = new Set();
  await Promise.all([
    broadcastersStore.fetchBroadcasters(),
    broadcastersStore.fetchImageConfig(),
    broadcastersStore.fetchCategories(),
    broadcastersStore.fetchAttributePresets(),
    broadcastersStore.fetchAvailabilityOptions(),
  ]);
}

const filteredBroadcasters = computed(() => {
  return broadcastersStore.broadcasters.filter((b) => {
    if (certifiedOnly.value && !b.certified) return false;
    if (categoryFilter.value && b.category !== categoryFilter.value) return false;
    if (availabilityFilter.value && b.availability !== availabilityFilter.value) return false;
    if (searchQuery.value) {
      const q = searchQuery.value.toLowerCase().trim();
      const nameMatch = b.display_name?.toLowerCase().includes(q);
      const bioMatch = b.bio?.toLowerCase().includes(q);
      if (!nameMatch && !bioMatch) return false;
    }
    return true;
  });
});

function hasPhoto(b: Broadcaster): boolean {
  if (!b || !b.id) return false;
  if (failedImages.value.has(b.id)) return false;
  return !!(b.photos && b.photos.length > 0 && b.photos[0]?.image_url);
}

function getPrimaryPhoto(b: Broadcaster): string {
  if (hasPhoto(b)) {
    return b.photos?.[0]?.image_url || '';
  }
  return '';
}

function getAvailabilityBadge(status?: string): { label: string; color: string; icon: string } {
  switch (status?.toLowerCase()) {
    case 'live':
    case 'in diretta':
      return { label: t('broadcasters.availability.live'), color: 'deep-orange', icon: 'sensors' };
    case 'busy':
    case 'occupato':
      return { label: t('broadcasters.availability.busy'), color: 'warning', icon: 'do_not_disturb_on' };
    case 'offline':
    case 'non disponibile':
      return { label: t('broadcasters.availability.offline'), color: 'grey-7', icon: 'power_settings_new' };
    case 'available':
    case 'disponibile':
      return { label: t('broadcasters.availability.available'), color: 'positive', icon: 'check_circle' };
    default:
      return { label: status || t('broadcasters.notSet'), color: 'blue-grey', icon: 'help_outline' };
  }
}

function onImgError(e: Event, b?: Broadcaster) {
  if (b?.id) {
    const updated = new Set(failedImages.value);
    updated.add(b.id);
    failedImages.value = updated;
  }
}

// Handlers apertura dialog separati
function openEditBroadcaster(b: Broadcaster) {
  selectedBroadcaster.value = b;
  editBroadcasterOpen.value = true;
}

function openEditAttributes(b: Broadcaster) {
  selectedBroadcaster.value = b;
  editAttributesOpen.value = true;
}

function openEditHours(b: Broadcaster) {
  selectedBroadcaster.value = b;
  editHoursOpen.value = true;
}

function openEditPhotos(b: Broadcaster) {
  selectedBroadcaster.value = b;
  editPhotosOpen.value = true;
}

function confirmDelete(b: Broadcaster) {
  selectedBroadcaster.value = b;
  deleteConfirmOpen.value = true;
}

async function executeDelete() {
  if (!selectedBroadcaster.value?.id) return;
  try {
    await broadcastersStore.deleteBroadcaster(selectedBroadcaster.value.id);
    Notify.create({
      type: 'positive',
      message: t('broadcasters.deletedSuccess'),
    });
    deleteConfirmOpen.value = false;
    selectedBroadcaster.value = null;
  } catch (err: unknown) {
    Notify.create({
      type: 'negative',
      message: err instanceof Error ? err.message : t('broadcasters.deleteError'),
    });
  }
}

function getTotalTimeSlots(b: Broadcaster): number {
  if (!b?.hours) return 0;
  return b.hours.reduce((acc, d) => acc + (d.hours?.length || 0), 0);
}

function onSaved(updated: Broadcaster) {
  selectedBroadcaster.value = updated;
}
</script>

<style scoped>
.broadcaster-card {
  border-radius: 10px;
  overflow: hidden;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.broadcaster-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.08);
}

.card-cover-wrapper {
  position: relative;
  width: 100%;
  aspect-ratio: 2 / 3;
  max-height: 280px;
  overflow: hidden;
  background-color: #e0e0e0;
}

.card-cover-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.card-cover-placeholder {
  width: 100%;
  height: 100%;
  background-color: #e0e0e0;
}

.card-cover-overlay {
  position: absolute;
  top: 8px;
  left: 8px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  z-index: 1;
}

.card-menu-btn {
  position: absolute;
  top: 8px;
  right: 8px;
  z-index: 2;
}

.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.table-cover-box {
  width: 44px;
  height: 66px;
  aspect-ratio: 2 / 3;
  border-radius: 4px;
  overflow: hidden;
  background-color: #e0e0e0;
}

.table-cover-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.table-cover-placeholder {
  width: 100%;
  height: 100%;
  background-color: #e0e0e0;
}

.camera-icon-dark {
  color: #424242;
}

.filter-color-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  border: 1px solid rgba(0, 0, 0, 0.15);
}
</style>
