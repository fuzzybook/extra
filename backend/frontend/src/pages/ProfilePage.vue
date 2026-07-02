<template>
  <q-page class="q-pa-md">
    <div class="text-h5 q-mb-md">{{ t('profile.title') }}</div>

    <q-banner v-if="profilesStore.error" class="bg-negative text-white q-mb-md" rounded>
      {{ profilesStore.error }}
    </q-banner>

    <q-card v-if="profilesStore.profile" flat bordered style="max-width: 480px">
      <q-card-section class="q-gutter-md">
        <AvatarUpload v-model="form.avatar_url" :uploader="profilesStore.uploadAvatar" />

        <q-input v-model="form.display_name" :label="t('fields.displayName')" />

        <div class="text-caption text-grey">
          {{ profilesStore.profile.email }} · {{ profilesStore.profile.role }}
        </div>
      </q-card-section>

      <q-card-actions align="right">
        <q-btn color="primary" :label="t('actions.save')" :loading="profilesStore.loading" @click="save" />
      </q-card-actions>
    </q-card>

    <div v-else-if="!profilesStore.isAuthenticated" class="text-grey">
      {{ t('profile.loginRequired') }}
    </div>
  </q-page>
</template>

<script setup lang="ts">
import { onMounted, reactive, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useProfilesStore, type Profile } from '@/stores/profiles-store';
import AvatarUpload from '@/components/AvatarUpload.vue';

const { t } = useI18n();
const profilesStore = useProfilesStore();

const form = reactive({
  display_name: '',
  avatar_url: '',
});

function syncForm(profile: Profile | null) {
  form.display_name = profile?.display_name ?? '';
  form.avatar_url = profile?.avatar_url ?? '';
}

watch(() => profilesStore.profile, syncForm, { immediate: true });

async function save() {
  try {
    await profilesStore.update({ ...form });
  } catch {
    // profilesStore.error already holds the failure message
  }
}

onMounted(async () => {
  if (profilesStore.isAuthenticated && !profilesStore.profile) {
    try {
      await profilesStore.me();
    } catch {
      // profilesStore.error already holds the failure message
    }
  }
});
</script>
