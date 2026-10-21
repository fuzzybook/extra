<template>
  <q-layout view="hHh lpR fFf" class="blank-layout">
    <q-page-container>
      <router-view />
    </q-page-container>
  </q-layout>
</template>

<script setup lang="ts">
// Blank layout senza header né drawer per pagine pubbliche / landing
</script>

<style scoped>
.blank-layout {
  min-height: 100vh;
  background-color: #fafafa;
}
</style>
