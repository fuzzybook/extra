<template>
  <q-layout view="lHh Lpr lFf">
    <q-header elevated>
      <q-toolbar>
        <q-btn
          flat
          dense
          round
          icon="menu"
          aria-label="Menu"
          @click="toggleLeftDrawer"
        />

        <q-toolbar-title>
          PublicSite Dashboard
        </q-toolbar-title>

        <div>Quasar v{{ $q.version }}</div>
      </q-toolbar>
    </q-header>

    <!-- Colori espliciti: il testo del sito pubblico è bianco (tema scuro) -->
    <q-drawer
      v-model="leftDrawerOpen"
      show-if-above
      bordered
      class="bg-white text-dark"
    >
      <q-list>
        <q-item-label header>
          Navigazione
        </q-item-label>

        <q-item clickable v-ripple to="/dashboard" exact active-class="text-primary bg-blue-1">
          <q-item-section avatar>
            <q-icon name="dashboard" />
          </q-item-section>
          <q-item-section>
            <q-item-label>Dashboard</q-item-label>
            <q-item-label caption>Area Riservata SPA</q-item-label>
          </q-item-section>
        </q-item>

        <q-item clickable v-ripple to="/" exact>
          <q-item-section avatar>
            <q-icon name="home" />
          </q-item-section>
          <q-item-section>
            <q-item-label>Home</q-item-label>
            <q-item-label caption>Pagina Pubblica SSG</q-item-label>
          </q-item-section>
        </q-item>

        <q-separator class="q-my-sm" />

        <!-- Profilo dell'utente collegato: avatar e dati del profilo -->
        <q-item-label header>
          {{ t('userProfile.title') }}
        </q-item-label>

        <q-item v-if="!session.loaded.value">
          <q-item-section avatar>
            <q-spinner size="20px" />
          </q-item-section>
        </q-item>

        <q-item v-else-if="!session.loggedIn.value" clickable v-ripple :to="loginRoute">
          <q-item-section avatar>
            <q-icon name="login" />
          </q-item-section>
          <q-item-section>{{ t('brEditor.menu.loginPrompt') }}</q-item-section>
        </q-item>

        <template v-else>
          <q-item clickable v-ripple @click="avatarDialogOpen = true">
            <q-item-section avatar>
              <q-avatar size="32px">
                <img v-if="session.me.value?.avatar_url" :src="session.me.value.avatar_url" alt="" />
                <q-icon v-else name="account_circle" color="primary" size="32px" />
              </q-avatar>
            </q-item-section>
            <q-item-section>
              <q-item-label>{{ t('userProfile.avatar') }}</q-item-label>
              <q-item-label caption>{{ t('userProfile.avatarCaption') }}</q-item-label>
            </q-item-section>
          </q-item>
          <q-item clickable v-ripple @click="dataDialogOpen = true">
            <q-item-section avatar>
              <q-icon name="badge" color="primary" />
            </q-item-section>
            <q-item-section>
              <q-item-label>{{ t('userProfile.data') }}</q-item-label>
              <q-item-label caption>{{ t('userProfile.dataCaption') }}</q-item-label>
            </q-item-section>
          </q-item>
          <q-item clickable v-ripple @click="personalDialogOpen = true">
            <q-item-section avatar>
              <q-icon name="contact_mail" color="primary" />
            </q-item-section>
            <q-item-section>
              <q-item-label>{{ t('userProfile.personal') }}</q-item-label>
              <q-item-label caption>{{ t('userProfile.personalCaption') }}</q-item-label>
            </q-item-section>
          </q-item>
        </template>

        <!-- Profilo broadcaster dell'utente collegato: le stesse azioni del
             menu di riga della pagina Broadcasters del pannello. Solo per chi
             ha un profilo broadcaster (o se non è stato possibile caricarlo). -->
        <template v-if="session.loggedIn.value && (myBroadcaster || profileError)">
          <q-separator class="q-my-sm" />

          <q-item-label header>
            {{ t('brEditor.menu.title') }}
          </q-item-label>

          <q-item v-if="!myBroadcaster">
            <q-item-section class="text-grey-7">
              {{ t('brEditor.menu.loadError') }}
            </q-item-section>
          </q-item>

          <template v-else>
            <q-item v-for="action in profileActions" :key="action.key" clickable v-ripple @click="action.open">
              <q-item-section avatar>
                <q-icon :name="action.icon" :color="action.color" />
              </q-item-section>
              <q-item-section>
                <q-item-label>{{ t(`brEditor.menu.${action.key}`) }}</q-item-label>
                <q-item-label caption>{{ t(`brEditor.menu.${action.key}Caption`) }}</q-item-label>
              </q-item-section>
            </q-item>

            <q-separator inset class="q-my-xs" />

            <!-- Visibilità del profilo nel sito pubblico (campo enabled) -->
            <q-item tag="label" v-ripple :disable="savingVisibility">
              <q-item-section avatar>
                <q-icon :name="myBroadcaster.enabled ? 'visibility' : 'visibility_off'" :color="myBroadcaster.enabled ? 'positive' : 'grey-6'" />
              </q-item-section>
              <q-item-section>
                <q-item-label>{{ t('brEditor.menu.visibility') }}</q-item-label>
                <q-item-label caption>
                  {{ myBroadcaster.enabled ? t('brEditor.menu.visibleCaption') : t('brEditor.menu.hiddenCaption') }}
                </q-item-label>
              </q-item-section>
              <q-item-section side>
                <q-toggle
                  :model-value="myBroadcaster.enabled"
                  color="positive"
                  :disable="savingVisibility"
                  :aria-label="t('brEditor.menu.visibility')"
                  @update:model-value="setVisibility"
                />
              </q-item-section>
            </q-item>
          </template>
        </template>

        <q-separator class="q-my-sm" />

        <q-item-label header>
          Documentazione
        </q-item-label>

        <EssentialLink
          v-for="link in linksList"
          :key="link.label"
          v-bind="link"
        />
      </q-list>
    </q-drawer>

    <q-page-container>
      <router-view />
    </q-page-container>

    <UserAvatarDialog v-model="avatarDialogOpen" />
    <UserDataDialog v-model="dataDialogOpen" />
    <UserPersonalDialog v-model="personalDialogOpen" />

    <!-- Dialoghi condivisi con il pannello (packages/broadcaster-editor) -->
    <EditBroadcasterDialog v-model="openDialog.profile" :broadcaster="myBroadcaster" @saved="onSaved" />
    <EditAttributesDialog v-model="openDialog.attributes" :broadcaster="myBroadcaster" @saved="onSaved" />
    <EditHoursDialog v-model="openDialog.hours" :broadcaster="myBroadcaster" @saved="onSaved" />
    <EditPhotosDialog v-model="openDialog.photos" :broadcaster="myBroadcaster" @saved="onSaved" />

  </q-layout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { Notify } from 'quasar';
import { useI18n } from 'vue-i18n';
import {
  EditAttributesDialog,
  EditBroadcasterDialog,
  EditHoursDialog,
  EditPhotosDialog,
  type Broadcaster,
} from '@extra/broadcaster-editor';
import { APIError, ErrCode } from '@/encore/client';
import { api } from '@/encore/api';
import EssentialLink, { type EssentialLinkProps } from '@/components/EssentialLink.vue';
import UserAvatarDialog from '@/components/UserAvatarDialog.vue';
import UserDataDialog from '@/components/UserDataDialog.vue';
import UserPersonalDialog from '@/components/UserPersonalDialog.vue';
import { useSession } from '@/composables/useSession';

const { t } = useI18n();
const route = useRoute();
const session = useSession();
const loginRoute = computed(() => `/login?redirect=${encodeURIComponent(route.fullPath)}`);

// Profilo broadcaster di chi è collegato (il backend lo mostra completo al
// proprietario e gli permette di modificarlo).
const myBroadcaster = ref<Broadcaster | null>(null);
const profileError = ref(false);

async function loadProfile() {
  myBroadcaster.value = null;
  profileError.value = false;
  const me = session.me.value;
  if (!me) return;
  try {
    myBroadcaster.value = await api.broadcasters.GetBroadcasterByUserID(me.user_id);
  } catch (err) {
    // NotFound: l'utente non ha un profilo broadcaster.
    profileError.value = !(err instanceof APIError && err.code === ErrCode.NotFound);
  }
}

onMounted(async () => {
  await session.refresh();
  await loadProfile();
});
watch(() => session.me.value?.user_id, () => void loadProfile());

const avatarDialogOpen = ref(false);
const dataDialogOpen = ref(false);
const personalDialogOpen = ref(false);

const openDialog = reactive({ profile: false, attributes: false, hours: false, photos: false });
const profileActions = [
  { key: 'editProfile', icon: 'edit', color: 'primary', open: () => (openDialog.profile = true) },
  { key: 'editAttributes', icon: 'tune', color: 'indigo', open: () => (openDialog.attributes = true) },
  { key: 'editHours', icon: 'schedule', color: 'teal', open: () => (openDialog.hours = true) },
  { key: 'editPhotos', icon: 'photo_library', color: 'deep-orange', open: () => (openDialog.photos = true) },
] as const;

function onSaved(updated: Broadcaster) {
  myBroadcaster.value = updated;
}

// Mostra o nasconde il profilo nel sito pubblico (il proprietario può farlo).
const savingVisibility = ref(false);

async function setVisibility(enabled: boolean) {
  const id = myBroadcaster.value?.id;
  if (!id || savingVisibility.value) return;
  savingVisibility.value = true;
  try {
    myBroadcaster.value = await api.broadcasters.UpdateBroadcasterVisibility(id, { enabled });
    Notify.create({
      type: 'positive',
      message: enabled ? t('brEditor.menu.shownSuccess') : t('brEditor.menu.hiddenSuccess'),
    });
  } catch {
    Notify.create({ type: 'negative', message: t('brEditor.menu.visibilityError') });
  } finally {
    savingVisibility.value = false;
  }
}

const linksList: EssentialLinkProps[] = [
  {
    label: 'Docs',
    caption: 'quasar.dev',
    icon: 'school',
    link: 'https://quasar.dev'
  },
  {
    label: 'GitHub',
    caption: 'github.com/quasarframework',
    icon: 'code',
    link: 'https://github.com/quasarframework'
  },
  {
    label: 'Discord Chat Channel',
    caption: 'chat.quasar.dev',
    icon: 'chat',
    link: 'https://chat.quasar.dev'
  },
  {
    label: 'Forum',
    caption: 'forum.quasar.dev',
    icon: 'record_voice_over',
    link: 'https://forum.quasar.dev'
  },
  {
    label: 'Twitter',
    caption: '@quasarframework',
    icon: 'rss_feed',
    link: 'https://twitter.quasar.dev'
  },
  {
    label: 'Facebook',
    caption: '@QuasarFramework',
    icon: 'public',
    link: 'https://facebook.quasar.dev'
  },
  {
    label: 'Quasar Awesome',
    caption: 'Community Quasar projects',
    icon: 'favorite',
    link: 'https://awesome.quasar.dev'
  }
];

const leftDrawerOpen = ref(false);

function toggleLeftDrawer () {
  leftDrawerOpen.value = !leftDrawerOpen.value;
}
</script>
