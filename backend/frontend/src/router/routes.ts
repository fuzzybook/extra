import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/layouts/MainLayout.vue'),
    children: [
      { path: '', component: () => import('@/pages/IndexPage.vue') },
      { path: 'second', component: () => import('@/pages/SecondPage.vue') },
      { path: 'login', component: () => import('@/pages/LoginPage.vue') },
      { path: 'password-reset', component: () => import('@/pages/PasswordResetPage.vue') },
      { path: 'register', component: () => import('@/pages/RegisterPage.vue') },
      { path: 'welcome', component: () => import('@/pages/WelcomePage.vue') },
      { path: 'profile', component: () => import('@/pages/ProfilePage.vue') },
      { path: '401', component: () => import('@/pages/ErrorUnauthorized.vue') },
     ],
  },
 {
    path: '/admin',
    component: () => import('@/layouts/AdminLayout.vue'),
    children: [
       { path: 'dashboard', component: () => import('@/pages/admin/DashboardPage.vue') },
       { path: 'profiles', component: () => import('@/pages/admin/ProfilesPage.vue') }
    ],
  },
  // Always leave this as last one,
  // but you can also remove it
  {
    path: '/:catchAll(.*)*',
    component: () => import('@/pages/ErrorNotFound.vue'),
  },
];

export default routes;
