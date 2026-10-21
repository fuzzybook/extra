<template>
  <q-layout view="hHh lpR fFf" class="public-layout">
    <PublicHeader />

    <q-page-container>
      <router-view />
    </q-page-container>
  </q-layout>
</template>

<script setup lang="ts">
import PublicHeader from '@/components/PublicHeader.vue';
import { useSeoLinks } from '@/composables/useSeoLinks';

useSeoLinks();
</script>

<style scoped>
.public-layout {
  min-height: 100vh;
  background-color: #0b0b0e;
}
</style>
