<template>
  <q-page class="q-pa-md">
    <div class="row items-center justify-between q-mb-md">
      <div>
        <h1 class="text-h4 q-my-none text-weight-bold">{{ t('promos.title') }}</h1>
        <p class="text-subtitle1 text-grey-7 q-mb-md">{{ t('promos.subtitle') }}</p>
      </div>
      <div class="row items-center">
        <div class="row items-center q-gutter-md q-px-md q-py-xs bg-transparent" >
          <q-radio
            v-model="viewMode"
            val="table"
            :label="t('promos.viewModeTable')"
            dense
            color="primary"
          >
            <q-tooltip>{{ t('promos.viewModeTableTooltip') }}</q-tooltip>
          </q-radio>
          <q-radio
            v-model="viewMode"
            val="calendar"
            :label="t('promos.viewModeCalendar')"
            dense
            color="primary"
          >
            <q-tooltip>{{ t('promos.viewModeCalendarTooltip') }}</q-tooltip>
          </q-radio>
        </div>
        <q-btn
          color="primary"
          icon="add"
          :label="t('promos.newPromoWithPos', { pos: getPositionMeta(activeTab).label })"
          unelevated
          @click="openDialog"
        />
      </div>
    </div>

    <!-- Banner errori globali -->
    <q-banner v-if="errorMessage" class="bg-negative text-white q-mb-md" rounded>
      <template v-slot:avatar>
        <q-icon name="error" color="white" />
      </template>
      {{ errorMessage }}
      <template v-slot:action>
        <q-btn flat color="white" :label="t('promos.retry')" @click="fetchPromos" />
      </template>
    </q-banner>

    <!-- Vista Calendario Promozioni -->
    <PromoCalendarView
      v-if="viewMode === 'calendar'"
      :promos="promosList"
      class="q-mb-md"
      @edit="openEditDialog"
    />

    <!-- Vista Tabellare -->
    <div v-else>
      <!-- Barra Ricerca Globale e Tab Selezione PromoPosition -->
      <q-card flat bordered class="q-mb-md">
      <div class="row items-center justify-between q-pa-sm q-col-gutter-sm">
        <div class="col-12 col-md-auto">
          <!-- 1. Tabs per le tre PromoPosition -->
          <q-tabs
            v-model="activeTab"
            dense
            active-color="primary"
            indicator-color="primary"
            align="left"
            narrow-indicator
          >
            <q-tab
              :name="PromoPosition.PromoLeft"
              :icon="getPositionMeta(PromoPosition.PromoLeft).icon"
              :label="getPositionMeta(PromoPosition.PromoLeft).label"
            />
            <q-tab
              :name="PromoPosition.PromoRight1"
              :icon="getPositionMeta(PromoPosition.PromoRight1).icon"
              :label="getPositionMeta(PromoPosition.PromoRight1).label"
            />
            <q-tab
              :name="PromoPosition.PromoRight2"
              :icon="getPositionMeta(PromoPosition.PromoRight2).icon"
              :label="getPositionMeta(PromoPosition.PromoRight2).label"
            />
          </q-tabs>
        </div>

        <div class="col-12 col-md-4">
          <q-input
            v-model="searchFilter"
            dense
            outlined
            debounce="300"
            :placeholder="t('promos.searchPlaceholder')"
            clearable
          >
            <template v-slot:append>
              <q-icon name="search" />
            </template>
          </q-input>
        </div>
      </div>
    </q-card>

    <!-- 2. Ogni tab richiama la <q-table> solo per quella PromoPosition -->
    <q-tab-panels v-model="activeTab" animated keep-alive class="bg-transparent">
      <!-- Pannello Tab: PromoLeft -->
      <q-tab-panel :name="PromoPosition.PromoLeft" class="q-pa-none">
        <q-card flat bordered>
          <q-table
            :rows="promosLeft"
            :columns="columns"
            row-key="id"
            :loading="loading"
            :filter="searchFilter"
            :no-data-label="t('promos.noDataPromoLeft')"
            :no-results-label="t('promos.noResults')"
          >

            <template v-slot:body-cell-preview="props">
              <q-td :props="props" align="center" style="width: 100px;">
                <div class="promo-thumb-wrapper flex flex-center">
                  <img
                    :src="props.value || noImagePlaceholder"
                    :alt="props.row.item?.[0]?.title || t('promos.columns.preview')"
                    class="promo-table-thumb"
                    :class="{ 'cursor-pointer': !!props.value }"
                    @click="props.value ? openTableImagePreview(props.row) : undefined"
                    @error="onImageError"
                  />
                  <q-tooltip v-if="props.value">
                    {{ t('promos.clickToEnlarge') }}
                  </q-tooltip>
                  <q-tooltip v-else>
                    {{ t('promos.noImageAssociated') }}
                  </q-tooltip>
                </div>
              </q-td>
            </template>

            <template v-slot:body-cell-section="props">
              <q-td :props="props">
                <q-badge
                  :color="getSectionMeta(props.row.position, props.value).color"
                  :label="getSectionMeta(props.row.position, props.value).label"
                  rounded
                  class="q-px-sm q-py-xs text-weight-medium"
                />
              </q-td>
            </template>

            <template v-slot:body-cell-items="props">
              <q-td :props="props">
                <div v-if="props.row.item && props.row.item.length > 0">
                  <q-chip
                    v-for="(it, idx) in props.row.item"
                    :key="idx"
                    dense
                    color="blue-grey-1"
                    text-color="primary"
                    icon="campaign"
                    class="q-mr-xs"
                  >
                    {{ it.title || `Item #${Number(idx) + 1}` }}
                    <q-tooltip v-if="it.content && it.content.length > 0">
                      <div v-for="(cnt, cIdx) in it.content" :key="cIdx">
                        {{ t('promos.itemContentTooltip', { name: cnt.broadcaster?.display_name || '-', id: cnt.broadcaster?.short_id || '-' }) }}
                      </div>
                    </q-tooltip>
                  </q-chip>
                </div>
                <span v-else class="text-grey-6 text-italic">{{ t('promos.noItems') }}</span>
              </q-td>
            </template>

            <template v-slot:body-cell-weeks="props">
              <q-td :props="props">
                <div v-if="props.row.weeks && props.row.weeks.length > 0" class="row items-center q-gutter-xs">
                  <q-badge
                    v-for="w in props.row.weeks.slice(0, 4)"
                    :key="w"
                    color="primary"
                    :label="`W${w}`"
                    rounded
                  />
                  <q-badge
                    v-if="props.row.weeks.length > 4"
                    color="grey-7"
                    :label="`+${props.row.weeks.length - 4}`"
                    rounded
                  >
                    <q-tooltip>{{ props.row.weeks.map((w: number) => `W${w}`).join(', ') }}</q-tooltip>
                  </q-badge>
                </div>
                <span v-else class="text-grey-6 text-italic text-caption">{{ t('promos.allYear') }}</span>
              </q-td>
            </template>

            <template v-slot:body-cell-actions="props">
              <q-td :props="props" align="center">
                <q-btn
                  flat
                  round
                  dense
                  color="primary"
                  icon="edit"
                  @click="openEditDialog(props.row)"
                >
                  <q-tooltip>{{ t('promos.editPromoTooltip') }}</q-tooltip>
                </q-btn>
              </q-td>
            </template>
          </q-table>
        </q-card>
      </q-tab-panel>

      <!-- Pannello Tab: PromoRight1 -->
      <q-tab-panel :name="PromoPosition.PromoRight1" class="q-pa-none">
        <q-card flat bordered>
          <q-table
            :rows="promosRight1"
            :columns="columns"
            row-key="id"
            :loading="loading"
            :filter="searchFilter"
            :no-data-label="t('promos.noDataPromoRight1')"
            :no-results-label="t('promos.noResults')"
          >

            <template v-slot:body-cell-preview="props">
              <q-td :props="props" align="center" style="width: 100px;">
                <div class="promo-thumb-wrapper flex flex-center">
                  <img
                    :src="props.value || noImagePlaceholder"
                    :alt="props.row.item?.[0]?.title || t('promos.columns.preview')"
                    class="promo-table-thumb"
                    :class="{ 'cursor-pointer': !!props.value }"
                    @click="props.value ? openTableImagePreview(props.row) : undefined"
                    @error="onImageError"
                  />
                  <q-tooltip v-if="props.value">
                    {{ t('promos.clickToEnlarge') }}
                  </q-tooltip>
                  <q-tooltip v-else>
                    {{ t('promos.noImageAssociated') }}
                  </q-tooltip>
                </div>
              </q-td>
            </template>

            <template v-slot:body-cell-section="props">
              <q-td :props="props">
                <q-badge
                  :color="getSectionMeta(props.row.position, props.value).color"
                  :label="getSectionMeta(props.row.position, props.value).label"
                  rounded
                  class="q-px-sm q-py-xs text-weight-medium"
                />
              </q-td>
            </template>

            <template v-slot:body-cell-items="props">
              <q-td :props="props">
                <div v-if="props.row.item && props.row.item.length > 0">
                  <q-chip
                    v-for="(it, idx) in props.row.item"
                    :key="idx"
                    dense
                    color="blue-grey-1"
                    text-color="primary"
                    icon="campaign"
                    class="q-mr-xs"
                  >
                    {{ it.title || `Item #${Number(idx) + 1}` }}
                    <q-tooltip v-if="it.content && it.content.length > 0">
                      <div v-for="(cnt, cIdx) in it.content" :key="cIdx">
                        {{ t('promos.itemContentTooltip', { name: cnt.broadcaster?.display_name || '-', id: cnt.broadcaster?.short_id || '-' }) }}
                      </div>
                    </q-tooltip>
                  </q-chip>
                </div>
                <span v-else class="text-grey-6 text-italic">{{ t('promos.noItems') }}</span>
              </q-td>
            </template>

            <template v-slot:body-cell-weeks="props">
              <q-td :props="props">
                <div v-if="props.row.weeks && props.row.weeks.length > 0" class="row items-center q-gutter-xs">
                  <q-badge
                    v-for="w in props.row.weeks.slice(0, 4)"
                    :key="w"
                    color="primary"
                    :label="`W${w}`"
                    rounded
                  />
                  <q-badge
                    v-if="props.row.weeks.length > 4"
                    color="grey-7"
                    :label="`+${props.row.weeks.length - 4}`"
                    rounded
                  >
                    <q-tooltip>{{ props.row.weeks.map((w: number) => `W${w}`).join(', ') }}</q-tooltip>
                  </q-badge>
                </div>
                <span v-else class="text-grey-6 text-italic text-caption">{{ t('promos.allYear') }}</span>
              </q-td>
            </template>

            <template v-slot:body-cell-actions="props">
              <q-td :props="props" align="center">
                <q-btn
                  flat
                  round
                  dense
                  color="primary"
                  icon="edit"
                  @click="openEditDialog(props.row)"
                >
                  <q-tooltip>{{ t('promos.editPromoTooltip') }}</q-tooltip>
                </q-btn>
              </q-td>
            </template>
          </q-table>
        </q-card>
      </q-tab-panel>

      <!-- Pannello Tab: PromoRight2 -->
      <q-tab-panel :name="PromoPosition.PromoRight2" class="q-pa-none">
        <q-card flat bordered>
          <q-table
            :rows="promosRight2"
            :columns="columns"
            row-key="id"
            :loading="loading"
            :filter="searchFilter"
            :no-data-label="t('promos.noDataPromoRight2')"
            :no-results-label="t('promos.noResults')"
          >

            <template v-slot:body-cell-preview="props">
              <q-td :props="props" align="center" style="width: 100px;">
                <div class="promo-thumb-wrapper flex flex-center">
                  <img
                    :src="props.value || noImagePlaceholder"
                    :alt="props.row.item?.[0]?.title || t('promos.columns.preview')"
                    class="promo-table-thumb"
                    :class="{ 'cursor-pointer': !!props.value }"
                    @click="props.value ? openTableImagePreview(props.row) : undefined"
                    @error="onImageError"
                  />
                  <q-tooltip v-if="props.value">
                    {{ t('promos.clickToEnlarge') }}
                  </q-tooltip>
                  <q-tooltip v-else>
                    {{ t('promos.noImageAssociated') }}
                  </q-tooltip>
                </div>
              </q-td>
            </template>

            <template v-slot:body-cell-section="props">
              <q-td :props="props">
                <q-badge
                  :color="getSectionMeta(props.row.position, props.value).color"
                  :label="getSectionMeta(props.row.position, props.value).label"
                  rounded
                  class="q-px-sm q-py-xs text-weight-medium"
                />
              </q-td>
            </template>

            <template v-slot:body-cell-items="props">
              <q-td :props="props">
                <div v-if="props.row.item && props.row.item.length > 0">
                  <q-chip
                    v-for="(it, idx) in props.row.item"
                    :key="idx"
                    dense
                    color="blue-grey-1"
                    text-color="primary"
                    icon="campaign"
                    class="q-mr-xs"
                  >
                    {{ it.title || `Item #${Number(idx) + 1}` }}
                    <q-tooltip v-if="it.content && it.content.length > 0">
                      <div v-for="(cnt, cIdx) in it.content" :key="cIdx">
                        {{ t('promos.itemContentTooltip', { name: cnt.broadcaster?.display_name || '-', id: cnt.broadcaster?.short_id || '-' }) }}
                      </div>
                    </q-tooltip>
                  </q-chip>
                </div>
                <span v-else class="text-grey-6 text-italic">{{ t('promos.noItems') }}</span>
              </q-td>
            </template>

            <template v-slot:body-cell-weeks="props">
              <q-td :props="props">
                <div v-if="props.row.weeks && props.row.weeks.length > 0" class="row items-center q-gutter-xs">
                  <q-badge
                    v-for="w in props.row.weeks.slice(0, 4)"
                    :key="w"
                    color="primary"
                    :label="`W${w}`"
                    rounded
                  />
                  <q-badge
                    v-if="props.row.weeks.length > 4"
                    color="grey-7"
                    :label="`+${props.row.weeks.length - 4}`"
                    rounded
                  >
                    <q-tooltip>{{ props.row.weeks.map((w: number) => `W${w}`).join(', ') }}</q-tooltip>
                  </q-badge>
                </div>
                <span v-else class="text-grey-6 text-italic text-caption">{{ t('promos.allYear') }}</span>
              </q-td>
            </template>

            <template v-slot:body-cell-actions="props">
              <q-td :props="props" align="center">
                <q-btn
                  flat
                  round
                  dense
                  color="primary"
                  icon="edit"
                  @click="openEditDialog(props.row)"
                >
                  <q-tooltip>{{ t('promos.editPromoTooltip') }}</q-tooltip>
                </q-btn>
              </q-td>
            </template>
          </q-table>
        </q-card>
      </q-tab-panel>
    </q-tab-panels>
  </div>

    <!-- Dialog modale per inserimento Promo -->
    <q-dialog v-model="dialogOpen" persistent>
      <q-card style="min-width: 520px; max-width: 90vw">
        <q-card-section class="row items-center q-pb-none">
          <div class="text-h6 text-weight-bold">{{ editingPromo ? t('promos.dialog.editTitle', { id: editingPromo.id }) : t('promos.dialog.newTitle') }}</div>
          <q-space />
          <q-btn icon="close" flat round dense v-close-popup :disable="submitting" />
        </q-card-section>

        <q-card-section>
          <q-form @submit="onSubmit" class="q-gutter-md">
            <!-- 3. Indicazione Posizione Promo assegnata automaticamente dal tab attivo -->
            <div class="row items-center justify-between q-pa-sm bg-grey-1 rounded-borders">
              <span class="text-subtitle2 text-grey-8">{{ t('promos.dialog.positionLabel') }}</span>
              <q-badge
                :color="getPositionMeta(selectedPosition).color"
                :label="getPositionMeta(selectedPosition).label"
                rounded
                class="q-px-sm q-py-xs text-weight-medium"
              />
            </div>

            <!-- 4. Selezione Sezione con <q-select> basato sulle costanti/enum di promo.go -->
            <q-select
              v-model="selectedSection"
              :options="sectionOptions"
              :label="t('promos.dialog.sectionLabel')"
              emit-value
              map-options
              outlined
              dense
              :rules="[(val: string) => (!!val && val.trim().length > 0) || t('promos.dialog.sectionRequired')]"
            >
              <template v-slot:option="scope">
                <q-item v-bind="scope.itemProps">
                  <q-item-section>
                    <q-item-label>{{ scope.opt.label }}</q-item-label>
                  </q-item-section>
                  <q-item-section side>
                    <q-badge :color="scope.opt.color ?? 'grey-7'" :label="scope.opt.value" rounded />
                  </q-item-section>
                </q-item>
              </template>
            </q-select>

            <!-- Periodo di validità: Selezione Settimane Lavorative (1-52) -->
            <div>
              <div class="text-caption text-weight-medium text-grey-8 q-mb-xs">
                {{ t('promos.dialog.validityPeriod') }}
              </div>
              <WeekSelector
                v-model="selectedWeeks"
                v-model:year="selectedYear"
                class="q-mb-sm"
              />
            </div>

            <q-separator class="q-my-sm" />

            <!-- Sezione Dettagli Item Promo -->
            <div class="text-subtitle2 text-weight-bold text-grey-8">{{ t('promos.dialog.bannerDetails') }}</div>

            <q-input
              v-model="itemTitle"
              :label="t('promos.dialog.itemTitleLabel')"
              outlined
              dense
              :rules="[(val: string) => (!!val && val.trim().length > 0) || t('promos.dialog.itemTitleRequired')]"
            />

            <!-- Broadcaster mostrato dalla promo (immagine: la sua prima foto) -->
            <q-select
              v-model="contentBroadcasterShortId"
              :options="filteredBroadcasterOptions"
              :label="t('promos.dialog.broadcasterLabel')"
              emit-value
              map-options
              use-input
              input-debounce="0"
              outlined
              dense
              :loading="broadcastersStore.loading"
              :rules="[(val: string | null) => !!val || t('promos.dialog.broadcasterRequired')]"
              @filter="filterBroadcasters"
            >
              <template #option="scope">
                <q-item v-bind="scope.itemProps">
                  <q-item-section avatar>
                    <q-avatar rounded size="40px" color="grey-3" text-color="grey-7">
                      <img v-if="scope.opt.image" :src="scope.opt.image" :alt="scope.opt.label" style="object-fit: cover" />
                      <q-icon v-else name="person" />
                    </q-avatar>
                  </q-item-section>
                  <q-item-section>
                    <q-item-label>{{ scope.opt.label }}</q-item-label>
                    <q-item-label caption>{{ scope.opt.value }}</q-item-label>
                  </q-item-section>
                  <q-item-section v-if="!scope.opt.enabled" side>
                    <q-badge color="grey-6" :label="t('promos.dialog.broadcasterDisabled')" />
                  </q-item-section>
                </q-item>
              </template>
              <template #selected-item="scope">
                <div class="row items-center no-wrap q-gutter-sm">
                  <q-avatar rounded size="24px" color="grey-3" text-color="grey-7">
                    <img v-if="scope.opt.image" :src="scope.opt.image" :alt="scope.opt.label" style="object-fit: cover" />
                    <q-icon v-else name="person" size="16px" />
                  </q-avatar>
                  <span>{{ scope.opt.label }}</span>
                  <span class="text-caption text-grey-6">{{ scope.opt.value }}</span>
                </div>
              </template>
              <template #no-option>
                <q-item>
                  <q-item-section class="text-grey-6">{{ t('promos.dialog.noBroadcasters') }}</q-item-section>
                </q-item>
              </template>
            </q-select>

            <div class="row items-center q-gutter-md q-pt-xs">
              <q-toggle v-model="contentMobileOnly" :label="t('promos.dialog.mobileOnly')" dense />
            </div>

            <div class="row justify-end q-gutter-sm q-pt-md">
              <q-btn flat :label="t('actions.cancel')" v-close-popup :disable="submitting" />
              <q-btn
                unelevated
                color="primary"
                type="submit"
                :label="editingPromo ? t('promos.dialog.updatePromo') : t('promos.dialog.savePromo')"
                :loading="submitting"
              />
            </div>
          </q-form>
        </q-card-section>
      </q-card>
    </q-dialog>

    <!-- Dialog ingrandimento anteprima immagine dalla tabella -->
    <q-dialog v-model="tableImagePreviewOpen">
      <q-card style="max-width: 90vw; max-height: 90vh;" class="bg-grey-1">
        <q-card-section class="row items-center justify-between q-pb-none">
          <div>
            <div class="text-subtitle1 text-weight-bold">{{ t('promos.previewDialog.tablePreviewTitle') }}</div>
            <div v-if="tableImagePreviewBroadcaster?.display_name" class="text-caption text-grey-7">
              {{ tableImagePreviewBroadcaster.display_name }}
            </div>
          </div>
          <q-btn icon="close" flat round dense v-close-popup />
        </q-card-section>
        <q-card-section class="flex flex-center q-pa-md">
          <!-- Il riquadro prende le dimensioni dell'immagine caricata; la didascalia (HTML)
               viene posizionata in box_x / box_y solo dopo il rendering dell'immagine. -->
          <div class="promo-preview-frame" :style="tableImagePreviewFrameStyle">
            <img
              :key="tableImagePreviewUrl"
              :src="tableImagePreviewUrl"
              :alt="tableImagePreviewBroadcaster?.display_name || t('promos.previewDialog.tablePreviewTitle')"
              class="promo-preview-img"
              :class="{ 'promo-preview-img--sized': !!tableImagePreviewSize }"
              @load="onPreviewImageLoad"
              @error="tableImagePreviewSize = null"
            />
            <!-- Livello grande quanto la foto: la didascalia scala con essa (vedi app.css) -->
            <div v-if="tableImagePreviewSize" class="promo-preview-caption-layer photo-caption-frame">
              <div
                v-if="tableImagePreviewDescription"
                class="promo-preview-caption photo-caption-text"
                :style="tableImagePreviewCaptionStyle"
                v-html="tableImagePreviewDescription"
              />
            </div>
          </div>
        </q-card-section>
      </q-card>
    </q-dialog>
  </q-page>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import type { QTableColumn } from 'quasar';
import { Notify } from 'quasar';
import { advertising as AdvertisingNS } from '@/encore/client';
import { api } from '@/encore/api';
import {
  PromoPosition,
  PROMO_SECTION_OPTIONS,
  getPositionMeta,
  getSectionMeta,
  type PromoSectionOption,
} from '@/types/promo';
import PromoCalendarView from '@/components/PromoCalendarView.vue';
import WeekSelector from '@/components/WeekSelector.vue';
import { useBroadcastersStore } from '@/stores/broadcasters-store';
import { sanitizeInlineHtml } from '@/utils/sanitize';

// 5. Tipi fortemente tipizzati generati da Encore
type Promo = AdvertisingNS.Promo;
type PromoItem = AdvertisingNS.PromoItem;
type PromoContent = AdvertisingNS.PromoContent;
type PromoBroadcaster = AdvertisingNS.PromoBroadcaster;
type CreatePromoParams = AdvertisingNS.CreatePromoParams;
type UpdatePromoParams = AdvertisingNS.UpdatePromoParams;

const { t } = useI18n();
const client = api;
const broadcastersStore = useBroadcastersStore();

const promosList = ref<Promo[]>([]);
const loading = ref(false);
const submitting = ref(false);
const errorMessage = ref<string | null>(null);
const searchFilter = ref('');
const dialogOpen = ref(false);
const editingPromo = ref<Promo | null>(null);
const viewMode = ref<'table' | 'calendar'>('table');

// Selezione settimane e anno per il dialog promo
const selectedWeeks = ref<number[]>([]);
const selectedYear = ref<number>(new Date().getFullYear());

// Calcola la data ISO del lunedì per una determinata settimana e anno
function getDateOfISOWeek(w: number, y: number): Date {
  const simple = new Date(Date.UTC(y, 0, 1 + (w - 1) * 7));
  const dow = simple.getUTCDay();
  const ISOweekStart = simple;
  if (dow <= 4) {
    ISOweekStart.setUTCDate(simple.getUTCDate() - simple.getUTCDay() + 1);
  } else {
    ISOweekStart.setUTCDate(simple.getUTCDate() + 8 - simple.getUTCDay());
  }
  return ISOweekStart;
}

// 1. Tab attivo per PromoPosition (PromoLeft come default)
const activeTab = ref<PromoPosition>(PromoPosition.PromoLeft);

// 2. Filtro delle promozioni per ciascuna PromoPosition
const promosLeft = computed<Promo[]>(() =>
  promosList.value.filter((p) => p.position === PromoPosition.PromoLeft)
);
const promosRight1 = computed<Promo[]>(() =>
  promosList.value.filter((p) => p.position === PromoPosition.PromoRight1)
);
const promosRight2 = computed<Promo[]>(() =>
  promosList.value.filter((p) => p.position === PromoPosition.PromoRight2)
);

// 3. Posizione form (predeterminata dal tab selezionato)
const selectedPosition = ref<PromoPosition>(PromoPosition.PromoLeft);

// 4. Sezione selezionata per il form
const selectedSection = ref<string>('');

// Opzioni sezioni reattive in base alla posizione selezionata
const sectionOptions = computed<PromoSectionOption[]>(() => {
  return PROMO_SECTION_OPTIONS[selectedPosition.value] ?? [];
});

const itemTitle = ref<string>('');
// Broadcaster selezionato (short_id): la promo mostra la sua prima foto.
// null = nessuna selezione (per q-select una stringa vuota è un valore scelto).
const contentBroadcasterShortId = ref<string | null>(null);
const broadcasterSearch = ref('');

interface BroadcasterOption {
  label: string;
  value: string;
  image: string;
  enabled: boolean;
}

const broadcasterOptions = computed<BroadcasterOption[]>(() =>
  broadcastersStore.broadcasters
    .filter((b) => !!b.short_id)
    .map((b) => ({
      label: b.display_name,
      value: b.short_id,
      image: b.photos?.find((p) => p.image_url)?.image_url || '',
      enabled: b.enabled,
    }))
    .sort((a, b) => a.label.localeCompare(b.label)),
);

const filteredBroadcasterOptions = computed<BroadcasterOption[]>(() => {
  const q = broadcasterSearch.value.trim().toLowerCase();
  if (!q) return broadcasterOptions.value;
  return broadcasterOptions.value.filter((o) => o.label.toLowerCase().includes(q) || o.value.includes(q));
});

function filterBroadcasters(val: string, update: (fn: () => void) => void) {
  update(() => {
    broadcasterSearch.value = val;
  });
}

const contentMobileOnly = ref<boolean>(false);

const noImagePlaceholder = '/no-image.svg';
const tableImagePreviewOpen = ref<boolean>(false);
const tableImagePreviewUrl = ref<string>('');
const tableImagePreviewBroadcaster = ref<PromoBroadcaster | null>(null);
// Dimensioni a cui viene mostrata l'immagine, calcolate dopo il suo caricamento
// (null finché l'immagine non è stata renderizzata).
const tableImagePreviewSize = ref<{ width: number; height: number } | null>(null);

const tableImagePreviewFrameStyle = computed(() =>
  tableImagePreviewSize.value
    ? { width: `${tableImagePreviewSize.value.width}px`, height: `${tableImagePreviewSize.value.height}px` }
    : {},
);

// Usa le dimensioni reali dell'immagine (naturalWidth/naturalHeight), adattate
// alla finestra mantenendo le proporzioni e senza ingrandirla oltre il naturale.
function onPreviewImageLoad(event: Event) {
  const img = event.target as HTMLImageElement;
  const naturalW = img.naturalWidth || tableImagePreviewBroadcaster.value?.width || 0;
  const naturalH = img.naturalHeight || tableImagePreviewBroadcaster.value?.height || 0;
  if (!naturalW || !naturalH) {
    tableImagePreviewSize.value = null;
    return;
  }
  const scale = Math.min(1, (window.innerWidth * 0.85) / naturalW, (window.innerHeight * 0.75) / naturalH);
  tableImagePreviewSize.value = {
    width: Math.round(naturalW * scale),
    height: Math.round(naturalH * scale),
  };
}
const tableImagePreviewDescription = computed(() =>
  sanitizeInlineHtml(tableImagePreviewBroadcaster.value?.description),
);

// box_x / box_y sono l'offset in percentuale della foto (0-100), quindi la
// didascalia resta nello stesso punto a qualunque dimensione sia mostrata.
function captionPercent(value: number | undefined): number {
  return Math.min(Math.max(value ?? 0, 0), 100);
}

const tableImagePreviewCaptionStyle = computed(() => {
  const b = tableImagePreviewBroadcaster.value;
  return {
    left: `${captionPercent(b?.box_x)}%`,
    top: `${captionPercent(b?.box_y)}%`,
  };
});

// Primo broadcaster della promo con una foto: è quello mostrato in tabella.
function getPromoBroadcaster(promo: Promo): PromoBroadcaster | null {
  for (const it of promo.item ?? []) {
    for (const cnt of it.content ?? []) {
      if (cnt.broadcaster?.image_url?.trim()) {
        return cnt.broadcaster;
      }
    }
  }
  return null;
}

function getPromoImageUrl(promo: Promo): string {
  return getPromoBroadcaster(promo)?.image_url?.trim() ?? '';
}

// Popup con la foto, il nome del broadcaster come sottotitolo e la didascalia.
function openTableImagePreview(promo: Promo) {
  const broadcaster = getPromoBroadcaster(promo);
  if (!broadcaster?.image_url) return;
  tableImagePreviewSize.value = null;
  tableImagePreviewUrl.value = broadcaster.image_url;
  tableImagePreviewBroadcaster.value = broadcaster;
  tableImagePreviewOpen.value = true;
}

function onImageError(e: Event) {
  const target = e.target as HTMLImageElement | null;
  if (target && !target.src.endsWith(noImagePlaceholder)) {
    target.src = noImagePlaceholder;
  }
}

const columns = computed<QTableColumn<Promo>[]>(() => [
  {
    name: 'preview',
    label: t('promos.columns.preview'),
    field: (row: Promo) => getPromoImageUrl(row),
    align: 'center',
    sortable: false,
  },
  {
    name: 'section',
    label: t('promos.columns.section'),
    field: (row: Promo) => row.section,
    align: 'left',
    sortable: true,
  },
  {
    name: 'items',
    label: t('promos.columns.items'),
    field: (row: Promo) => (row.item ? row.item.map((i) => i.title).join(', ') : ''),
    align: 'left',
    sortable: false,
  },
  {
    name: 'weeks',
    label: t('promos.columns.weeks'),
    field: (row: Promo) => (row.weeks && row.weeks.length > 0 ? row.weeks.map((w: number) => `W${w}`).join(', ') : t('promos.allYear')),
    align: 'left',
    sortable: false,
  },
  {
    name: 'created_at',
    label: t('promos.columns.createdAt'),
    field: (row: Promo) => new Date(row.created_at).toLocaleString(),
    align: 'left',
    sortable: true,
  },
  {
    name: 'actions',
    label: t('promos.columns.actions'),
    field: () => '',
    align: 'center',
    sortable: false,
  },
]);

async function fetchPromos() {
  loading.value = true;
  errorMessage.value = null;
  try {
    const res = await client.advertising.ListPromotions();
    promosList.value = res.promotions ?? [];
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err);
    errorMessage.value = t('promos.toasts.fetchError', { msg });
  } finally {
    loading.value = false;
  }
}

// 3. Il bottone per aggiungere una promo utilizza la PromoPosition selezionata nel tab
function openDialog() {
  editingPromo.value = null;
  selectedPosition.value = activeTab.value;

  // Inizializza la sezione con la prima opzione disponibile per la posizione attiva
  const opts = PROMO_SECTION_OPTIONS[activeTab.value];
  selectedSection.value = opts && opts.length > 0 && opts[0] ? opts[0].value : '';

  selectedWeeks.value = [];
  selectedYear.value = new Date().getFullYear();

  itemTitle.value = '';
  contentBroadcasterShortId.value = null;
  contentMobileOnly.value = false;
  dialogOpen.value = true;
}

// Apertura del dialogo per la modifica di una promozione esistente
function openEditDialog(promo: Promo) {
  editingPromo.value = promo;
  selectedPosition.value = promo.position;
  selectedSection.value = promo.section;
  selectedYear.value = promo.year || (promo.active_from ? new Date(promo.active_from).getFullYear() : new Date().getFullYear());
  selectedWeeks.value = promo.weeks ? [...promo.weeks] : [];

  const firstItem = promo.item?.[0];
  if (firstItem) {
    itemTitle.value = firstItem.title || '';
    const firstContent = firstItem.content?.[0];
    if (firstContent) {
      contentBroadcasterShortId.value = firstContent.broadcaster?.short_id || null;
      contentMobileOnly.value = !!firstContent.mobile_only;
    } else {
      contentBroadcasterShortId.value = null;
      contentMobileOnly.value = false;
    }
  } else {
    itemTitle.value = '';
    contentBroadcasterShortId.value = null;
    contentMobileOnly.value = false;
  }

  dialogOpen.value = true;
}

// 5. Creazione fortemente tipizzata tramite client generato Encore
async function onSubmit() {
  submitting.value = true;
  try {
    const promoContent: PromoContent = {
      broadcaster: { short_id: contentBroadcasterShortId.value ?? '' },
      mobile_only: contentMobileOnly.value,
    };

    const promoItem: PromoItem = {
      title: itemTitle.value.trim(),
      content: [promoContent],
    };

    let activeFromIso: string | undefined;
    if (selectedWeeks.value.length > 0) {
      const sortedWeeks = [...selectedWeeks.value].sort((a, b) => a - b);
      const minWeek = sortedWeeks[0] ?? 1;
      activeFromIso = getDateOfISOWeek(minWeek, selectedYear.value).toISOString();
    } else if (editingPromo.value?.active_from) {
      activeFromIso = new Date(editingPromo.value.active_from).toISOString();
    } else {
      activeFromIso = new Date().toISOString();
    }

    if (editingPromo.value) {
      const updatePayload: UpdatePromoParams = {
        position: selectedPosition.value,
        section: selectedSection.value,
        item: [promoItem],
        year: selectedYear.value,
        weeks: selectedWeeks.value,
        active_from: activeFromIso,
      };

      const updated = await client.advertising.UpdatePromo(editingPromo.value.id, updatePayload);
      const index = promosList.value.findIndex((p) => p.id === updated.id);
      if (index !== -1) {
        promosList.value[index] = updated;
      }
      dialogOpen.value = false;
      Notify.create({
        type: 'positive',
        message: t('promos.toasts.updateSuccess', { id: updated.id }),
        position: 'top-right',
      });
    } else {
      const createPayload: CreatePromoParams = {
        position: selectedPosition.value,
        section: selectedSection.value,
        item: [promoItem],
        year: selectedYear.value,
        weeks: selectedWeeks.value,
        active_from: activeFromIso,
      };

      const created = await client.advertising.CreatePromo(createPayload);
      promosList.value.unshift(created);
      dialogOpen.value = false;
      Notify.create({
        type: 'positive',
        message: t('promos.toasts.createSuccess', { id: created.id, pos: getPositionMeta(created.position).label }),
        position: 'top-right',
      });
    }
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err);
    Notify.create({
      type: 'negative',
      message: t('promos.toasts.saveError', { msg }),
      position: 'top-right',
    });
  } finally {
    submitting.value = false;
  }
}

onMounted(() => {
  void fetchPromos();
  void broadcastersStore.fetchBroadcasters();
});
</script>

<style scoped>
.promo-preview-frame {
  position: relative;
  display: inline-block;
  line-height: 0;
  border-radius: 6px;
  overflow: hidden;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.promo-preview-img {
  display: block;
  max-width: 85vw;
  max-height: 75vh;
  object-fit: contain;
}

/* Dopo il caricamento l'immagine riempie esattamente il riquadro dimensionato */
.promo-preview-img--sized {
  width: 100%;
  height: 100%;
  max-width: none;
  max-height: none;
}

.promo-preview-caption-layer {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

/* Stessa geometria della didascalia nell'editor delle foto (padding in em,
   bordo da 1px), senza i controlli di trascinamento. */
.promo-preview-caption {
  position: absolute;
  max-width: 85%;
  padding: 0.35em 0.7em;
  border: 1px solid transparent;
  color: #ffffff;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.85);
  word-break: break-word;
}

.promo-thumb-wrapper {
  min-height: 60px;
  height: 60px;
}

.promo-table-thumb {
  max-height: 60px;
  max-width: 90px;
  width: auto;
  height: auto;
  object-fit: contain;
  border-radius: 4px;
  border: 1px solid #e2e8f0;
  background-color: #f8fafc;
  transition: transform 0.15s ease, box-shadow 0.15s ease;
}

.promo-table-thumb.cursor-pointer:hover {
  transform: scale(1.08);
  box-shadow: 0 4px 10px rgba(0, 0, 0, 0.15);
}
</style>
