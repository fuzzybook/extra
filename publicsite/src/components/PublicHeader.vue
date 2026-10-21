<template>
  <header class="header">
    <div class="row container items-center justify-between no-wrap q-py-sm">
      <!-- Logo -->
      <div class="logo-container q-pa-sm q-ml-md">
        <router-link :to="getMenuRoute('')">
          <img class="logo-img" src="/assets/logo.svg" alt="Logo" />
        </router-link>
      </div>

      <!-- Top Banner -->
      <div class="topbanner col q-pa-sm q-mx-md">
        <router-link :to="getMenuRoute('')">
          <img class="image-banner" src="/assets/banners/banner.svg" width="700" height="80" alt="Top Banner" />
        </router-link>
      </div>

      <!-- Right Menu: Languages & Actions -->
      <div class="column menu_right q-mr-md">
        <div class="row items-center justify-end">
          <div class="nav__item" :class="{ 'nav__item--active': currentLang === 'it' }">
            <router-link :to="getLangRoute('it')" class="nav__link" @click="setLanguage('it')">IT</router-link>
          </div>
          <div class="nav__item" :class="{ 'nav__item--active': currentLang === 'fr' }">
            <router-link :to="getLangRoute('fr')" class="nav__link" @click="setLanguage('fr')">FR</router-link>
          </div>
          <div class="nav__item" :class="{ 'nav__item--active': currentLang === 'de' }">
            <router-link :to="getLangRoute('de')" class="nav__link" @click="setLanguage('de')">DE</router-link>
          </div>
          <div class="nav__item" :class="{ 'nav__item--active': currentLang === 'en' }">
            <router-link :to="getLangRoute('en')" class="nav__link" @click="setLanguage('en')">EN</router-link>
          </div>
        </div>
        <div class="row items-center justify-end adv_buttons q-mt-xs">
          <div class="nav__item nav__advertise" :class="{ 'nav__item--current': isSectionActive('advertise') }">
            <router-link class="nav__link" :to="getMenuRoute('advertise')">{{ t('actions.advertise') }}</router-link>
          </div>
          <div class="nav__item nav__item--button" :class="{ 'nav__item--current': isSectionActive('login') }">
            <router-link class="nav__link" :to="getMenuRoute('login')">{{ t('actions.login') }}</router-link>
          </div>
        </div>
      </div>
    </div>

    <!-- Main Navigation Menu (Bottoni) -->
    <a class="menu-opener" href="#"><i class="menu-opener__icon"></i></a>
    <nav class="menu__container row items-center justify-center q-px-md q-py-sm">
      <div class="menu__item" :class="{ 'menu__item--active': isHomeActive }">
        <router-link class="menu__link" :to="getMenuRoute('')">
          <span class="menu__dot color-home">•</span>
          <span class="menu__text">{{ t('menu.home') }}</span>
        </router-link>
      </div>

      <div class="menu__item" :class="{ 'menu__item--active': isSectionActive('one') }">
        <router-link class="menu__link" :to="getMenuRoute('one')">
          <span class="menu__dot color-elite">•</span>
          <span class="menu__text">{{ t('menu.one') }}</span>
        </router-link>
      </div>

      <div class="menu__item" :class="{ 'menu__item--active': isSectionActive('two') }">
        <router-link class="menu__link" :to="getMenuRoute('two')">
          <span class="menu__dot color-parlors">•</span>
          <span class="menu__text">{{ t('menu.two') }}</span>
        </router-link>
      </div>

      <div class="menu__item" :class="{ 'menu__item--active': isSectionActive('three') }">
        <router-link class="menu__link" :to="getMenuRoute('three')">
          <span class="menu__dot color-private">•</span>
          <span class="menu__text">{{ t('menu.three') }}</span>
        </router-link>
      </div>

      <div class="menu__item" :class="{ 'menu__item--active': isSectionActive('webcam') }">
        <router-link class="menu__link" :to="getMenuRoute('webcam')">
          <span class="menu__dot color-webcam">•</span>
          <span class="menu__text">{{ t('menu.webcam') }}</span>
        </router-link>
      </div>

      <div class="menu__item" :class="{ 'menu__item--active': isSectionActive('contact') }">
        <router-link class="menu__link" :to="getMenuRoute('contact')">
          <span class="menu__dot color-contact">•</span>
          <span class="menu__text">{{ t('menu.contact') }}</span>
        </router-link>
      </div>
    </nav>
  </header>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import type { SupportedLocale } from '@/i18n';

const route = useRoute();
const { t, locale } = useI18n();

const currentLang = computed<SupportedLocale>(() => {
  const param = route.params.locale as string | undefined;
  if (param && ['it', 'fr', 'de', 'en'].includes(param)) {
    return param as SupportedLocale;
  }
  return (locale.value as SupportedLocale) || 'it';
});

// Sincronizza vue-i18n quando il parametro :locale della rotta cambia
watch(
  () => route.params.locale,
  (newLocale) => {
    if (newLocale && ['it', 'fr', 'de', 'en'].includes(newLocale as string)) {
      locale.value = newLocale as SupportedLocale;
    }
  },
  { immediate: true }
);

function setLanguage(lang: SupportedLocale) {
  locale.value = lang;
  if (typeof window !== 'undefined') {
    try {
      localStorage.setItem('user.locale', lang);
    } catch {
      // Safe fallback
    }
  }
}

// Estrae il segmento pulito senza codice lingua
function getSubPath(): string {
  const segments = route.path.split('/').filter(Boolean);
  if (segments.length > 0 && ['it', 'fr', 'de', 'en'].includes(segments[0]!)) {
    segments.shift();
  }
  return segments.join('/');
}

// Genera la rotta quando si clicca un pulsante lingua (mantenendo la sottopagina corrente)
function getLangRoute(targetLang: SupportedLocale): string {
  const sub = getSubPath();
  return `/${targetLang}${sub ? `/${sub}` : ''}`;
}

// Genera la rotta per un elemento del menu conservando la lingua corrente
function getMenuRoute(section: string): string {
  const lang = currentLang.value;
  if (!section) {
    return `/${lang}`;
  }
  return `/${lang}/${section}`;
}

const isHomeActive = computed(() => {
  const sub = getSubPath();
  return !sub;
});

function isSectionActive(section: string): boolean {
  return getSubPath() === section;
}
</script>

<style scoped>
.header {
  width: 100%;
  background-color: #000000;
  border-bottom: 2px solid #222226;
}

.container {
  width: 100%;
  max-width: 1400px;
  margin-right: auto;
  margin-left: auto;
}

.logo-container {
  width: 280px;
  height: 80px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px dotted rgba(255, 255, 255, 0.2);
  border-radius: 6px;
}

.logo-img {
  max-height: 65px;
  max-width: 100%;
  height: auto;
  width: auto;
  object-fit: contain;
  display: block;
}

.topbanner {
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px dotted rgba(255, 255, 255, 0.2);
  border-radius: 6px;
  min-height: 80px;
}

.image-banner {
  max-width: 100%;
  height: auto;
  max-height: 80px;
  display: block;
}

.menu_right {
  display: flex;
  flex-direction: column;
  justify-content: center;
  min-width: 200px;
}

.nav__item {
  color: white;
  border: 1px solid rgba(255, 255, 255, 0.6);
  padding: 3px 8px;
  margin-left: 6px;
  font-weight: 600;
  font-size: 12px;
  text-decoration: none;
  text-align: center;
  border-radius: 4px;
  transition: all 0.2s ease-in-out;
  cursor: pointer;
}

.nav__item:hover,
.nav__item--active {
  background-color: #ffffff;
}

.nav__item:hover .nav__link,
.nav__item--active .nav__link {
  color: #000000 !important;
}

.nav__link {
  text-decoration: none;
  color: #ffffff;
  display: block;
}

.adv_buttons {
  margin-top: 8px;
}

.nav__advertise {
  border-color: #ffd700;
}

.nav__advertise .nav__link {
  color: #ffd700;
}

.nav__advertise:hover {
  background-color: #ffd700;
}

.nav__advertise:hover .nav__link {
  color: #000000 !important;
}

.nav__item--button {
  border-color: #00d2ff;
}

.nav__item--button .nav__link {
  color: #00d2ff;
}

.nav__item--button:hover {
  background-color: #00d2ff;
}

.nav__item--button:hover .nav__link {
  color: #000000 !important;
}

/* ========================================================================= */
/* MENU CONTAINER: Ancore visualizzate come bottoni                           */
/* ========================================================================= */
.menu-opener {
  display: none;
}

.menu__container {
  width: 100%;
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  background: #111116;
  border-top: 1px solid #1f1f27;
  border-bottom: 2px solid #2a2a36;
}

.menu__item {
  display: inline-flex;
}

.menu__link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 8px 20px;
  border-radius: 6px;
  background-color: #1a1a24;
  color: #f0f0f5;
  border: 1px solid #2f2f40;
  text-decoration: none;
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0.6px;
  text-transform: uppercase;
  cursor: pointer;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.4);
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  user-select: none;
}

.menu__link:hover {
  background-color: #272738;
  border-color: #ffffff;
  color: #ffffff;
  transform: translateY(-2px);
  box-shadow: 0 6px 14px rgba(0, 0, 0, 0.6), 0 0 8px rgba(255, 255, 255, 0.15);
}

.menu__link:active {
  transform: translateY(0);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.4);
}

.menu__item--active .menu__link {
  background: linear-gradient(135deg, #ffffff 0%, #e0e0e6 100%);
  color: #000000;
  border-color: #ffffff;
  font-weight: 800;
  box-shadow: 0 0 16px rgba(255, 255, 255, 0.4);
}

.menu__item--active .menu__link:hover {
  background: #ffffff;
  color: #000000;
  border-color: #ffffff;
}

.menu__dot {
  font-size: 20px;
  line-height: 1;
  display: inline-block;
  vertical-align: middle;
}

.color-home {
  color: #ff3b30;
}

.color-elite {
  color: #ff9500;
}

.color-parlors {
  color: #bf5af2;
}

.color-private {
  color: #30d158;
}

.color-webcam {
  color: #0a84ff;
}

.color-contact {
  color: #ff2d55;
}

@media (max-width: 1024px) {
  .container {
    flex-direction: column;
    gap: 12px;
  }
  .menu_right {
    align-items: center;
  }
}
</style>
