<template>
  <q-layout view="lHh Lpr lFf">
    <q-header elevated>
      <q-toolbar>
        <q-btn
          flat
          dense
          round
          icon="menu"
          :aria-label="t('nav.menu')"
          @click="toggleLeftDrawer"
        />

        <q-toolbar-title>
          {{ t('app.title') }}
        </q-toolbar-title>

        <q-select
          v-model="selectedLocale"
          :options="localeOptions"
          option-label="label"
          option-value="value"
          emit-value
          map-options
          dense
          borderless
          class="q-ml-md language-select"
          :aria-label="t('nav.language')"
        />




      </q-toolbar>
    </q-header>

    <q-drawer
      v-model="leftDrawerOpen"
      show-if-above
      bordered
    >
      <q-list>
        <DrawerUserCard />
        <q-separator />

        <q-item v-if="isAdmin" clickable to="/admin/dashboard">
          <q-item-section avatar>
            <q-icon name="dashboard" />
          </q-item-section>
          <q-item-section>
            <q-item-label>{{ t('admin.dashboard') }}</q-item-label>
          </q-item-section>
        </q-item>

        <q-item-label
          header
        >
          Essential Links
        </q-item-label>

        <q-item clickable to="/wideochat">
          <q-item-section avatar>
            <q-icon name="videocam" />
          </q-item-section>
          <q-item-section>
            <q-item-label>{{ t('nav.videochat') }}</q-item-label>
          </q-item-section>
        </q-item>

        <q-item clickable to="/advertising">
          <q-item-section avatar>
            <q-icon name="campaign" />
          </q-item-section>
          <q-item-section>
            <q-item-label>Pubblicità</q-item-label>
          </q-item-section>
        </q-item>

        <q-item clickable to="/promos">
          <q-item-section avatar>
            <q-icon name="local_offer" />
          </q-item-section>
          <q-item-section>
            <q-item-label>{{ t('nav.promos') }}</q-item-label>
          </q-item-section>
        </q-item>

        <q-item clickable to="/broadcasters">
          <q-item-section avatar>
            <q-icon name="podcasts" />
          </q-item-section>
          <q-item-section>
            <q-item-label>{{ t('nav.broadcasters') }}</q-item-label>
          </q-item-section>
        </q-item>
      </q-list>
    </q-drawer>

    <q-page-container>
      <router-view />
    </q-page-container>
  </q-layout>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import DrawerUserCard from '@/components/DrawerUserCard.vue';
import { useProfilesStore } from '@/stores/profiles-store';
import { localeOptions, setLocale, type SupportedLocale } from '@/i18n';

const profilesStore = useProfilesStore();
const router = useRouter();
const { locale, t } = useI18n();
const selectedLocale = computed({
  get: () => locale.value as SupportedLocale,
  set: (value: SupportedLocale) => setLocale(value),
});
const isAdmin = computed(() => profilesStore.profile?.role === 'admin');

onMounted(async () => {
  if (profilesStore.isAuthenticated && !profilesStore.profile) {
    try {
      await profilesStore.me();
    } catch {
      profilesStore.logout();
    }
  }
});



const leftDrawerOpen = ref(false);

function toggleLeftDrawer () {
  leftDrawerOpen.value = !leftDrawerOpen.value;
}

async function logout() {
  profilesStore.logout();
  await router.push('/');
}
</script>

<style scoped>
.language-select {
  width: 90px;
  max-width: 90px;
  color: #fff;
  font-weight: 700;
}

.language-select :deep(.q-field__native),
.language-select :deep(.q-field__append),
.language-select :deep(.q-icon) {
  color: #fff;
  font-weight: 700;
}
</style>
