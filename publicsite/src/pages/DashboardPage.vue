<template>
  <q-page class="q-pa-md dashboard-page">
    <div class="row items-center justify-between q-mb-md">
      <div>
        <h1 class="text-h5 text-weight-bold q-my-none text-grey-9">Dashboard</h1>
        <p class="text-caption text-grey-6 q-mb-none">
          Area riservata SPA (Client-Side Rendering)
        </p>
      </div>

      <q-chip color="primary" text-color="white" icon="verified_user">
        SPA Mode
      </q-chip>
    </div>

    <!-- Stat Cards Overview -->
    <div class="row q-col-gutter-md q-mb-lg">
      <div class="col-12 col-sm-6 col-md-3">
        <q-card flat bordered class="stat-card">
          <q-card-section class="row items-center no-wrap">
            <q-avatar icon="people" color="blue-1" text-color="primary" size="48px" />
            <div class="q-ml-md">
              <div class="text-caption text-grey-6">Utenti Attivi</div>
              <div class="text-h6 text-weight-bold">1,280</div>
            </div>
          </q-card-section>
        </q-card>
      </div>

      <div class="col-12 col-sm-6 col-md-3">
        <q-card flat bordered class="stat-card">
          <q-card-section class="row items-center no-wrap">
            <q-avatar icon="trending_up" color="green-1" text-color="positive" size="48px" />
            <div class="q-ml-md">
              <div class="text-caption text-grey-6">Sessioni Oggi</div>
              <div class="text-h6 text-weight-bold">3,420</div>
            </div>
          </q-card-section>
        </q-card>
      </div>

      <div class="col-12 col-sm-6 col-md-3">
        <q-card flat bordered class="stat-card">
          <q-card-section class="row items-center no-wrap">
            <q-avatar icon="speed" color="amber-1" text-color="warning" size="48px" />
            <div class="q-ml-md">
              <div class="text-caption text-grey-6">Tempo Risposta</div>
              <div class="text-h6 text-weight-bold">42 ms</div>
            </div>
          </q-card-section>
        </q-card>
      </div>

      <div class="col-12 col-sm-6 col-md-3">
        <q-card flat bordered class="stat-card">
          <q-card-section class="row items-center no-wrap">
            <q-avatar icon="cloud_done" color="purple-1" text-color="purple" size="48px" />
            <div class="q-ml-md">
              <div class="text-caption text-grey-6">Stato Servizio</div>
              <div class="text-h6 text-weight-bold text-positive">Online</div>
            </div>
          </q-card-section>
        </q-card>
      </div>
    </div>

    <!-- Quick Actions / Activity Section -->
    <div class="row q-col-gutter-md">
      <div class="col-12 col-md-8">
        <q-card flat bordered class="dashboard-content-card">
          <q-card-section>
            <div class="text-subtitle1 text-weight-bold text-grey-9">Attività Recente</div>
            <div class="text-caption text-grey-6">Panoramica delle azioni eseguite sulla piattaforma</div>
          </q-card-section>
          <q-separator />
          <q-list separator>
            <q-item>
              <q-item-section avatar>
                <q-icon name="login" color="primary" />
              </q-item-section>
              <q-item-section>
                <q-item-label>Accesso completato con successo</q-item-label>
                <q-item-label caption>Sessione autenticata dal client</q-item-label>
              </q-item-section>
              <q-item-section side class="text-caption text-grey-6">
                2 min fa
              </q-item-section>
            </q-item>
            <q-item>
              <q-item-section avatar>
                <q-icon name="sync" color="secondary" />
              </q-item-section>
              <q-item-section>
                <q-item-label>Sincronizzazione dati completata</q-item-label>
                <q-item-label caption>CSR hydration attiva</q-item-label>
              </q-item-section>
              <q-item-section side class="text-caption text-grey-6">
                15 min fa
              </q-item-section>
            </q-item>
          </q-list>
        </q-card>
      </div>

      <div class="col-12 col-md-4">
        <q-card flat bordered class="dashboard-content-card">
          <q-card-section>
            <div class="text-subtitle1 text-weight-bold text-grey-9">Info Sessione SPA</div>
            <div class="text-caption text-grey-6">Riferimento architettura ibrida</div>
          </q-card-section>
          <q-separator />
          <q-card-section class="q-gutter-y-sm text-body2">
            <div>
              <strong>Rotta:</strong> <code class="bg-grey-2 q-pa-xs rounded-borders">/dashboard</code>
            </div>
            <div>
              <strong>Rendering:</strong> Client-Side (CSR / SPA)
            </div>
            <div>
              <strong>SSG Exclude:</strong> picomatch <code class="bg-grey-2 q-pa-xs rounded-borders">/dashboard/**</code>
            </div>
            <q-banner dense rounded class="bg-blue-1 text-primary q-mt-md">
              <template #avatar>
                <q-icon name="info" color="primary" />
              </template>
              Questa sezione è esclusa dal pre-rendering statico SSG.
            </q-banner>
          </q-card-section>
        </q-card>
      </div>
    </div>
  </q-page>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';

// Nelle pagine SPA escluse da SSG, l'accesso alle API del browser è sicuro
const loadedAt = ref<string>('');

onMounted(() => {
  loadedAt.value = new Date().toLocaleTimeString();
});
</script>

<style scoped>
.dashboard-page {
  background-color: #f8fafc;
  min-height: calc(100vh - 50px);
}

.stat-card {
  border-radius: 12px;
  background: #ffffff;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
}

.dashboard-content-card {
  border-radius: 12px;
  background: #ffffff;
}
</style>
