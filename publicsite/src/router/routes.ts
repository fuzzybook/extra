import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  // Studio in finestra separata ("Apri in finestra"): solo webchat, senza header del sito
  {
    path: '/:locale(it|fr|de|en)?/studio/window',
    component: () => import('@/layouts/BlankLayout.vue'),
    children: [
      {
        path: '',
        name: 'studio-window',
        component: () => import('@/pages/BroadcasterStudioPage.vue'),
      },
    ],
  },

  // Rotte pubbliche SSG con parametro lingua opzionale (it|fr|de|en)
  {
    path: '/:locale(it|fr|de|en)?',
    component: () => import('@/layouts/PublicLayout.vue'),
    children: [
      {
        path: '',
        name: 'home',
        component: () => import('@/pages/IndexPage.vue'),
      },
      {
        path: 'elite',
        name: 'elite',
        component: () => import('@/pages/ElitePage.vue'),
      },
      {
        path: 'clubs',
        name: 'clubs',
        component: () => import('@/pages/ClubsPage.vue'),
      },
      {
        path: 'soft',
        name: 'soft',
        component: () => import('@/pages/SoftPage.vue'),
      },
      {
        path: 'webcam',
        name: 'webcam',
        component: () => import('@/pages/WebcamPage.vue'),
      },
      {
        path: 'contact',
        name: 'contact',
        component: () => import('@/pages/ContactPage.vue'),
      },
      {
        path: 'advertise',
        name: 'advertise',
        component: () => import('@/pages/AdvertisePage.vue'),
      },
      {
        path: 'login',
        name: 'login',
        component: () => import('@/pages/LoginPage.vue'),
      },
      {
        path: 'terms',
        name: 'terms',
        component: () => import('@/pages/TermsPage.vue'),
      },
      {
        // Pagina pubblica del broadcaster, per short_id (6 caratteri)
        path: 'br/:shortId',
        name: 'broadcaster',
        component: () => import('@/pages/BroadcasterPage.vue'),
      },
      {
        // Webcam del broadcaster collegato: apre (o crea) la sua stanza
        path: 'studio',
        name: 'studio',
        component: () => import('@/pages/BroadcasterStudioPage.vue'),
      },
      {
        // Diretta del broadcaster: video e chat (packages/webchat)
        path: 'br/:shortId/live',
        name: 'broadcaster-live',
        component: () => import('@/pages/BroadcasterLivePage.vue'),
      },
    ],
  },

  // Rotta privata SPA: layout standard Quasar (header, drawer, navigazione)
  {
    path: '/dashboard',
    component: () => import('@/layouts/MainLayout.vue'),
    children: [
      {
        path: '',
        name: 'dashboard',
        component: () => import('@/pages/DashboardPage.vue'),
      },
    ],
  },

  // 404 Catch-All
  {
    path: '/:catchAll(.*)*',
    component: () => import('@/pages/ErrorNotFound.vue'),
  },
];

export default routes;
