import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
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
        path: 'one',
        name: 'one',
        component: () => import('@/pages/OnePage.vue'),
      },
      {
        path: 'two',
        name: 'two',
        component: () => import('@/pages/TwoPage.vue'),
      },
      {
        path: 'three',
        name: 'three',
        component: () => import('@/pages/ThreePage.vue'),
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
