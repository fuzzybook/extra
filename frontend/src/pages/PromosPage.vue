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
                    @click="props.value ? openTableImagePreview(props.value) : undefined"
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
                        {{ t('promos.itemContentTooltip', { url: cnt.url || '-', image: cnt.image_url || '-' }) }}
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
                    @click="props.value ? openTableImagePreview(props.value) : undefined"
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
                        {{ t('promos.itemContentTooltip', { url: cnt.url || '-', image: cnt.image_url || '-' }) }}
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
                    @click="props.value ? openTableImagePreview(props.value) : undefined"
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
                        {{ t('promos.itemContentTooltip', { url: cnt.url || '-', image: cnt.image_url || '-' }) }}
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

            <!-- Selezione o Upload Immagine Promo -->
            <div>
              <q-input
                v-model="contentImageUrl"
                :label="t('promos.dialog.imageUrlLabel')"
                outlined
                dense
                placeholder="https://example.com/promo.jpg"
              >
                <template #append>
                  <!-- Bottone View (Occhio) con popup anteprima dimensioni e descrizione -->
                  <q-btn
                    v-if="contentImageUrl"
                    round
                    flat
                    dense
                    color="teal"
                    icon="visibility"
                    @click="previewImageDialogOpen = true"
                  >
                    <q-tooltip>{{ t('promos.dialog.viewImageTooltip') }}</q-tooltip>
                  </q-btn>
                  <q-btn
                    round
                    flat
                    dense
                    color="primary"
                    icon="photo_library"
                    @click="selectImageDialogOpen = true"
                  >
                    <q-tooltip>{{ t('promos.dialog.selectFromGalleryTooltip') }}</q-tooltip>
                  </q-btn>
                  <q-btn
                    round
                    flat
                    dense
                    color="secondary"
                    icon="add_photo_alternate"
                    @click="uploadImageDialogOpen = true"
                  >
                    <q-tooltip>{{ t('promos.dialog.uploadNewImageTooltip') }}</q-tooltip>
                  </q-btn>
                  <q-btn
                    v-if="contentImageUrl"
                    round
                    flat
                    dense
                    color="grey-6"
                    icon="close"
                    @click="clearContentImage"
                  >
                    <q-tooltip>{{ t('promos.dialog.removeImageTooltip') }}</q-tooltip>
                  </q-btn>
                </template>
              </q-input>
            </div>

            <q-input
              v-model="contentUrl"
              :label="t('promos.dialog.destinationUrlLabel')"
              outlined
              dense
              placeholder="https://example.com/landing"
            />

            <!-- Dimensioni Immagine Promo -->
            <div class="q-pa-sm bg-grey-1 rounded-borders">
              <div class="row items-center justify-between q-mb-xs">
                <span class="text-caption text-weight-bold text-grey-8">
                  {{ t('promos.dialog.imageDimensionsTitle') }}
                  <span v-if="!isImageSizeEditable" class="text-caption text-grey-6 text-weight-normal">{{ t('promos.dialog.fixedPromoLeft') }}</span>
                  <span v-else class="text-caption text-primary text-weight-normal">{{ t('promos.dialog.editableForPos', { pos: getPositionMeta(selectedPosition).label }) }}</span>
                </span>
                <q-badge
                  v-if="effectiveDimensions"
                  color="blue-grey-2"
                  text-color="primary"
                  :label="`${effectiveDimensions.width}x${effectiveDimensions.height} px`"
                  rounded
                />
              </div>

              <!-- Se PromoLeft: readonly con lucchetto -->
              <div v-if="!isImageSizeEditable" class="row q-col-gutter-sm">
                <div class="col-4">
                  <q-input
                    :model-value="promoWidth"
                    :label="t('promos.dialog.widthLabel')"
                    outlined
                    dense
                    readonly
                  >
                    <template #append>
                      <q-icon name="lock" size="xs" color="grey-6" />
                    </template>
                  </q-input>
                </div>
                <div class="col-4">
                  <q-input
                    :model-value="promoHeight"
                    :label="t('promos.dialog.heightLabel')"
                    outlined
                    dense
                    readonly
                    :hint="t('promos.dialog.heightHint')"
                  >
                    <template #append>
                      <q-icon name="lock" size="xs" color="grey-6" />
                    </template>
                  </q-input>
                </div>
                <div class="col-4">
                  <q-input
                    :model-value="promoProportion"
                    :label="t('promos.dialog.proportionLabel')"
                    outlined
                    dense
                    readonly
                  >
                    <template #append>
                      <q-icon name="lock" size="xs" color="grey-6" />
                    </template>
                  </q-input>
                </div>
              </div>

              <!-- Se PromoRight1 o PromoRight2: edit abilitato -->
              <div v-else class="row q-col-gutter-sm">
                <div class="col-4">
                  <q-input
                    v-model.number="promoWidth"
                    type="number"
                    :label="t('promos.dialog.widthLabelRequired')"
                    outlined
                    dense
                    min="1"
                    :rules="[(val: number) => val > 0 || t('promos.dialog.invalidWidth')]"
                  />
                </div>
                <div class="col-4">
                  <q-input
                    v-model.number="promoHeight"
                    type="number"
                    :label="t('promos.dialog.heightLabel')"
                    outlined
                    dense
                    min="0"
                    :hint="t('promos.dialog.heightHint')"
                  />
                </div>
                <div class="col-4">
                  <q-input
                    v-model="promoProportion"
                    :label="t('promos.dialog.proportionLabel')"
                    outlined
                    dense
                    :disable="promoHeight !== 0"
                    :hint="promoHeight !== 0 ? t('promos.dialog.proportionIgnored') : t('promos.dialog.proportionHint')"
                  />
                </div>
              </div>
            </div>

            <div class="row items-center q-gutter-md q-pt-xs">
              <q-toggle v-model="contentMobileOnly" :label="t('promos.dialog.mobileOnly')" dense />
              <q-toggle v-model="contentNewWindow" :label="t('promos.dialog.newWindow')" dense />
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

    <!-- Dialog di Selezione Immagine da Galleria -->
    <SelectImageDialog
      v-model="selectImageDialogOpen"
      :target-width="effectiveDimensions.width"
      :target-height="promoHeight"
      :proportion="promoProportion"
      @selected="onImageSelected"
    />

    <!-- Dialog di Caricamento Nuova Immagine -->
    <ImmageDialog
      v-model="uploadImageDialogOpen"
      :initial-category="ImageCategory.Promo"
      :target-width="effectiveDimensions.width"
      :target-height="promoHeight"
      :proportion="promoProportion"
      @saved="onImageUploaded"
    />

    <!-- Dialog Anteprima Immagine nelle Dimensioni Previste con Descrizione -->
    <q-dialog v-model="previewImageDialogOpen">
      <q-card style="max-width: 90vw; max-height: 90vh" class="column">
        <q-card-section class="row items-center justify-between q-pb-none">
          <div class="row items-center q-gutter-xs">
            <q-icon name="visibility" size="sm" color="teal" />
            <span class="text-h6 text-weight-bold">{{ t('promos.previewDialog.title') }}</span>
          </div>
          <q-space />
          <q-btn icon="close" flat round dense v-close-popup />
        </q-card-section>

        <!-- Descrizione dell'immagine e dettagli dimensioni previste -->
        <q-card-section class="q-pt-sm q-pb-xs">
          <div class="row items-center justify-between q-gutter-sm q-mb-xs">
            <div class="text-subtitle1 text-weight-bold text-grey-9">
              {{ currentImageDescription || t('promos.previewDialog.noDescription') }}
            </div>
            <q-badge
              color="primary"
              :label="`${effectiveDimensions.width}x${effectiveDimensions.height} px`"
              class="q-px-sm q-py-xs text-weight-bold"
              rounded
            />
          </div>
          <div v-if="contentImageUrl" class="text-caption text-grey-7 ellipsis">
            {{ contentImageUrl }}
          </div>
        </q-card-section>

        <q-separator />

        <!-- Area Immagine renderizzata nelle dimensioni previste -->
        <q-card-section class="col scroll flex flex-center bg-grey-2 q-pa-md">
          <div
            class="rounded-borders shadow-2 bg-white flex flex-center overflow-hidden"
            :style="{
              width: `${effectiveDimensions.width}px`,
              height: `${effectiveDimensions.height}px`,
              maxWidth: '85vw',
              maxHeight: '65vh'
            }"
          >
            <q-img
              :src="contentImageUrl"
              class="full-width full-height"
              fit="contain"
              spinner-color="primary"
            />
          </div>
        </q-card-section>

        <q-card-actions align="right" class="q-pa-md bg-grey-1">
          <q-btn flat :label="t('promos.previewDialog.close')" v-close-popup />
        </q-card-actions>
      </q-card>
    </q-dialog>

    <!-- Dialog ingrandimento anteprima immagine dalla tabella -->
    <q-dialog v-model="tableImagePreviewOpen">
      <q-card style="max-width: 90vw; max-height: 90vh;" class="bg-grey-1">
        <q-card-section class="row items-center justify-between q-pb-none">
          <div class="text-subtitle1 text-weight-bold">{{ t('promos.previewDialog.tablePreviewTitle') }}</div>
          <q-btn icon="close" flat round dense v-close-popup />
        </q-card-section>
        <q-card-section class="flex flex-center q-pa-md">
          <img
            :src="tableImagePreviewUrl"
            style="max-width: 85vw; max-height: 75vh; object-fit: contain; border-radius: 6px; box-shadow: 0 4px 12px rgba(0,0,0,0.15);"
          />
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
  getPromoImageSize,
  calculatePromoDimensions,
  type PromoSectionOption,
} from '@/types/promo';
import SelectImageDialog from '@/components/SelectImageDialog.vue';
import ImmageDialog from '@/components/ImmageDialog.vue';
import PromoCalendarView from '@/components/PromoCalendarView.vue';
import WeekSelector from '@/components/WeekSelector.vue';
import { ImageCategory } from '@/types/images';
import { useImagesStore, type ImageItem } from '@/stores/images-store';

// 5. Tipi fortemente tipizzati generati da Encore
type Promo = AdvertisingNS.Promo;
type PromoItem = AdvertisingNS.PromoItem;
type PromoContent = AdvertisingNS.PromoContent;
type CreatePromoParams = AdvertisingNS.CreatePromoParams;
type UpdatePromoParams = AdvertisingNS.UpdatePromoParams;

const { t } = useI18n();
const client = api;
const imagesStore = useImagesStore();

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

const promoWidth = ref<number>(230);
const promoHeight = ref<number>(0);
const promoProportion = ref<string>('2:3');

const isImageSizeEditable = computed(() => {
  return selectedPosition.value === PromoPosition.PromoRight1 || selectedPosition.value === PromoPosition.PromoRight2;
});

const effectiveDimensions = computed(() => {
  return calculatePromoDimensions(
    promoWidth.value,
    promoHeight.value,
    promoProportion.value
  );
});

const itemTitle = ref<string>('');
const contentImageUrl = ref<string>('');
const contentImageDescription = ref<string>('');
const previewImageDialogOpen = ref<boolean>(false);
const selectImageDialogOpen = ref<boolean>(false);
const uploadImageDialogOpen = ref<boolean>(false);

function imageDescriptionText(desc: any): string {
  if (!desc) return '';
  if (typeof desc === 'string') return desc;
  return desc.text ?? '';
}

const currentImageDescription = computed(() => {
  if (contentImageDescription.value) {
    return contentImageDescription.value;
  }
  if (contentImageUrl.value) {
    const found = imagesStore.images.find((img) => img.url === contentImageUrl.value);
    if (found?.description) return imageDescriptionText(found.description);
  }
  return '';
});

function onImageSelected(image: ImageItem) {
  contentImageUrl.value = image.url;
  contentImageDescription.value = imageDescriptionText(image.description);
}

function onImageUploaded(image: ImageItem) {
  contentImageUrl.value = image.url;
  contentImageDescription.value = imageDescriptionText(image.description);
}

function clearContentImage() {
  contentImageUrl.value = '';
  contentImageDescription.value = '';
}
const contentUrl = ref<string>('');
const contentMobileOnly = ref<boolean>(false);
const contentNewWindow = ref<boolean>(false);

const noImagePlaceholder = '/no-image.svg';
const tableImagePreviewOpen = ref<boolean>(false);
const tableImagePreviewUrl = ref<string>('');

function getPromoImageUrl(promo: Promo): string {
  if (promo.item && Array.isArray(promo.item)) {
    for (const it of promo.item) {
      if (it.content && Array.isArray(it.content)) {
        for (const cnt of it.content) {
          if (cnt.image_url && cnt.image_url.trim().length > 0) {
            return cnt.image_url.trim();
          }
        }
      }
    }
  }
  return '';
}

function openTableImagePreview(url: string) {
  if (url) {
    tableImagePreviewUrl.value = url;
    tableImagePreviewOpen.value = true;
  }
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

  const defaultSize = getPromoImageSize(selectedPosition.value);
  promoWidth.value = defaultSize.width;
  promoHeight.value = defaultSize.height;
  promoProportion.value = defaultSize.proportion;

  itemTitle.value = '';
  contentImageUrl.value = '';
  contentImageDescription.value = '';
  contentUrl.value = '';
  contentMobileOnly.value = false;
  contentNewWindow.value = false;
  dialogOpen.value = true;
}

// Apertura del dialogo per la modifica di una promozione esistente
function openEditDialog(promo: Promo) {
  editingPromo.value = promo;
  selectedPosition.value = promo.position;
  selectedSection.value = promo.section;
  selectedYear.value = promo.year || (promo.active_from ? new Date(promo.active_from).getFullYear() : new Date().getFullYear());
  selectedWeeks.value = promo.weeks ? [...promo.weeks] : [];

  const defaultSize = getPromoImageSize(promo.position);

  const firstItem = promo.item?.[0];
  if (firstItem) {
    itemTitle.value = firstItem.title || '';
    const firstContent = firstItem.content?.[0];
    if (firstContent) {
      contentImageUrl.value = firstContent.image_url || '';
      const found = imagesStore.images.find((img) => img.url === contentImageUrl.value);
      contentImageDescription.value = found ? imageDescriptionText(found.description) : '';
      contentUrl.value = firstContent.url || '';
      contentMobileOnly.value = !!firstContent.mobile_only;
      contentNewWindow.value = !!firstContent.new_window;
      promoWidth.value = firstContent.width > 0 ? firstContent.width : defaultSize.width;
      promoHeight.value = firstContent.height !== undefined ? firstContent.height : defaultSize.height;
      promoProportion.value = promoHeight.value === 0 ? (defaultSize.proportion || '2:3') : '';
    } else {
      contentImageUrl.value = '';
      contentImageDescription.value = '';
      contentUrl.value = '';
      contentMobileOnly.value = false;
      contentNewWindow.value = false;
      promoWidth.value = defaultSize.width;
      promoHeight.value = defaultSize.height;
      promoProportion.value = defaultSize.proportion;
    }
  } else {
    itemTitle.value = '';
    contentImageUrl.value = '';
    contentImageDescription.value = '';
    contentUrl.value = '';
    contentMobileOnly.value = false;
    contentNewWindow.value = false;
    promoWidth.value = defaultSize.width;
    promoHeight.value = defaultSize.height;
    promoProportion.value = defaultSize.proportion;
  }

  dialogOpen.value = true;
}

// 5. Creazione fortemente tipizzata tramite client generato Encore
async function onSubmit() {
  submitting.value = true;
  try {
    const promoContent: PromoContent = {
      image_url: contentImageUrl.value.trim(),
      url: contentUrl.value.trim(),
      width: promoWidth.value,
      height: promoHeight.value,
      mobile_only: contentMobileOnly.value,
      mobile_image_url: '',
      new_window: contentNewWindow.value,
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
  void imagesStore.fetchImages();
});
</script>

<style scoped>
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
