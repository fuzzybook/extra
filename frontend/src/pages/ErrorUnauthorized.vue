<template>
  <div class="fullscreen bg-deep-orange text-white text-center q-pa-md flex flex-center">
    <div>
      <div style="font-size: 30vh">
        401
      </div>

      <div class="text-h2" style="opacity:.55">
        {{ $t('pages.unauthorized') }}
      </div>

      <div class="text-subtitle1 q-mt-md" style="opacity:.8">
        {{ $t('pages.unauthorizedHint') }}
      </div>

      <div class="row justify-center q-gutter-sm q-mt-xl">
        <q-btn
          color="white"
          text-color="deep-orange"
          unelevated
          to="/login"
          :label="$t('nav.login')"
          no-caps
        />
        <q-btn
          flat
          color="white"
          to="/"
          :label="$t('pages.goHome')"
          no-caps
        />
      </div>
    </div>
  </div>
</template>
