<template>
  <q-item
    clickable
    :to="profilesStore.isAuthenticated ? '/profile' : '/login'"
    class="drawer-user-card"
  >
    <q-item-section avatar>
      <q-avatar size="48px">
        <img v-if="profile?.avatar_url" :src="profile.avatar_url" />
        <q-icon v-else name="person" size="32px" />
      </q-avatar>
    </q-item-section>

    <q-item-section>
      <q-item-label class="text-weight-medium">
        {{ profileName }}
      </q-item-label>
      <q-item-label v-if="profile?.email" caption>
        {{ profile.email }}
      </q-item-label>
    </q-item-section>

    <q-item-section v-if="profilesStore.isAuthenticated" side>
      <div class="column items-center">
        <q-btn
          flat
          round
          dense
          icon="edit"
          :aria-label="t('actions.editProfile')"
          @click.stop.prevent="editProfile"
        >
          <q-tooltip>{{ t('actions.editProfile') }}</q-tooltip>
        </q-btn>
        <q-btn
          flat
          round
          dense
          icon="logout"
          :aria-label="t('nav.logoutAction')"
          @click.stop.prevent="logout"
        >
          <q-tooltip>{{ t('nav.logoutAction') }}</q-tooltip>
        </q-btn>
      </div>
    </q-item-section>
  </q-item>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { useProfilesStore } from '@/stores/profiles-store';

const profilesStore = useProfilesStore();
const router = useRouter();
const { t } = useI18n();
const profile = computed(() => profilesStore.profile);
const profileName = computed(() => {
  if (!profilesStore.isAuthenticated) return t('nav.login');
  return profile.value?.display_name || profile.value?.email || '';
});

async function logout() {
  await profilesStore.logout();
  await router.push('/');
}

async function editProfile() {
  await router.push({ path: '/profile', query: { mode: 'edit' } });
}
</script>

<style scoped>
.drawer-user-card {
  min-height: 72px;
}
</style>
