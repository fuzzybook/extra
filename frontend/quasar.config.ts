// Configuration for your app
// https://v2.quasar.dev/quasar-cli-vite/quasar-config-file

import { defineConfig } from '#q-app';

export default defineConfig((/* ctx */) => {
  const livekitSources = [
    'http://localhost:7880',
    'ws://localhost:7880',
    'http://127.0.0.1:7880',
    'ws://127.0.0.1:7880',
    'https://*.livekit.cloud',
    'wss://*.livekit.cloud',
  ];
  if (process.env.QCLI_LIVEKIT_URL) {
    const url = new URL(process.env.QCLI_LIVEKIT_URL);
    if (!['http:', 'https:', 'ws:', 'wss:'].includes(url.protocol) || !/^[a-z0-9.:[\]-]+$/i.test(url.host) || url.username || url.password) {
      throw new Error('QCLI_LIVEKIT_URL must be a LiveKit HTTP(S) or WS(S) URL without credentials');
    }
    const secure = url.protocol === 'https:' || url.protocol === 'wss:';
    livekitSources.push(`${secure ? 'https' : 'http'}://${url.host}`, `${secure ? 'wss' : 'ws'}://${url.host}`);
  }
  // The app calls the API on its own origin, so the HttpOnly session cookie is
  // first-party. In development the dev server proxies the API routes to the
  // Encore backend (override with QCLI_API_BASE_URL=http://host:port); in
  // production the reverse proxy (Caddy) must route the same paths.
  const apiBaseUrl = new URL(process.env.QCLI_API_BASE_URL || 'http://127.0.0.1:4002');
  if (!['http:', 'https:'].includes(apiBaseUrl.protocol) || apiBaseUrl.username || apiBaseUrl.password) {
    throw new Error('QCLI_API_BASE_URL must be an HTTP(S) URL without credentials');
  }
  const apiOrigin = apiBaseUrl.origin;
  // Top-level paths served by the Encore backend (see src/encore/client.ts).
  const apiPaths = [
    'admin', 'advertising', 'auth', 'broadcasters', 'extrahome', 'image', 'images',
    'profiles', 'promo', 'promotions', 'registration', 'videochat',
  ];
  return {
    htmlVariables: { livekitConnectSources: livekitSources.join(' ') },
    // https://v2.quasar.dev/quasar-cli-vite/prefetch-feature
    // preFetch: true,

    // app boot file (/src/boot)
    // --> boot files are part of "main.js"
    // https://v2.quasar.dev/quasar-cli-vite/boot-files
    boot: [
      'i18n',
    ],

    // https://v2.quasar.dev/quasar-cli-vite/quasar-config-file#css
    css: [
      'app.css'
    ],

    // https://github.com/quasarframework/quasar/tree/dev/extras
    extras: [
      // 'ionicons-v4',
      // 'mdi-v7',
      // 'fontawesome-v7',
      // 'eva-icons',
      // 'themify',
      // 'line-awesome',
      // 'roboto-font-latin-ext', // this or either 'roboto-font', NEVER both!

      'roboto-font', // optional, you are not bound to it
      'material-icons', // optional, you are not bound to it
    ],

    // https://v2.quasar.dev/quasar-cli-vite/quasar-config-file#build
    build: {
      target: {
        // browser: 'baseline-widely-available',
        // node: 'node22'
      },

      typescript: {
        strict: true,
        vueShim: true
        // extendTsConfig (tsConfig) {}
      },

      // https://v2.quasar.dev/quasar-cli-vite/page-routing-with-vue-router#filename-based-routing
      // filenameBasedRouting: true,

      vueRouterMode: 'hash', // available values: 'hash', 'history'
      // vueRouterBase,
      // vueDevtools,

      // publicPath: '/',
      // define: {},
      // defineEnv: {}
      // ignorePublicFolder: true,
      // minify: false,
      // distDir

      extendViteConf(viteConf) {
        viteConf.optimizeDeps = {
          ...viteConf.optimizeDeps,
          include: [
            ...(viteConf.optimizeDeps?.include ?? []),
            'zod',
            'vue-i18n',
          ],
        };
      },
      viteVuePluginOptions: {
        template: {
          compilerOptions: {
            isCustomElement: (tag: string) => tag.startsWith('extra-'),
          },
        },
      },

      // vitePlugins: [
      //   [ 'package-name', { ..pluginOptions.. }, { server: true, client: true } ]
      // ]
    },

    // https://v2.quasar.dev/quasar-cli-vite/quasar-config-file#devserver
    devServer: {
      // https: true,
      open: true, // opens browser window automatically
      proxy: {
        [`^/(${apiPaths.join('|')})(/|\\?|$)`]: {
          target: apiOrigin,
          changeOrigin: true,
        },
      },
    },

    // https://v2.quasar.dev/quasar-cli-vite/quasar-config-file#framework
    framework: {
      config: {},

      // iconSet: 'material-icons', // Quasar icon set
      // lang: 'en-US', // Quasar language pack

      // For special cases outside of where the auto-import strategy can have an impact
      // (like functional components as one of the examples),
      // you can manually specify Quasar components/directives to be available everywhere:
      //
      // components: [],
      // directives: [],

      // Quasar plugins
      plugins: ['Notify', 'Dialog']
    },

    // animations: 'all', // --- includes all animations
    // https://v2.quasar.dev/options/animations
    animations: [],

    // https://v2.quasar.dev/quasar-cli-vite/quasar-config-file#sourcefiles
    sourceFiles: {
      // rootComponent: 'src/App.vue',
      // router: 'src/router/index',
      store: 'src/stores/index',
      // pwaRegisterServiceWorker: 'src-pwa/register-sw',
      // pwaServiceWorker: 'src-pwa/sw/custom-sw',
      // pwaManifestFile: 'src-pwa/manifest.json',
      // electronMain: 'src-electron/electron-main',
      // electronPreload: 'src-electron/electron-preload'
      // bexManifestFile: 'src-bex/manifest.json'
    },

    // https://v2.quasar.dev/quasar-cli-vite/developing-ssr/configuring-ssr
    ssr: {
      prodPort: 3000, // The default port that the production server should use
      // (gets superseded if process.env.PORT is specified at runtime)

      middlewares: [
        'render' // keep this as last one
      ],

      // extendSSRPackageJson (pkgJson) {},
      // extendSSRWebserverConf (rolldownConf) {},

      // manualStoreSerialization: true,
      // manualStoreSsrContextInjection: true,
      // manualStoreHydration: true,
      // manualPostHydrationTrigger: true,

      pwa: false
      // pwaOfflineHtmlFilename: 'offline.html', // do NOT use index.html as name!

      // extendSSRGenerateSWOptions (cfg) {},
      // extendSSRInjectManifestOptions (cfg) {}
    },

    // https://v2.quasar.dev/quasar-cli-vite/developing-pwa/configuring-pwa
    pwa: {
      workboxMode: 'GenerateSW' // 'GenerateSW' or 'InjectManifest'
      // swFilename: 'sw.js',
      // manifestFilename: 'manifest.json',
      // extendPWAManifestJson (json) {},
      // useCredentialsForManifestTag: true,
      // injectPWAMetaTags: false,
      // extendPWACustomSWConf (rolldownConf) {},
      // extendPWAGenerateSWOptions (cfg) {},
      // extendPWAInjectManifestOptions (cfg) {},
      // extendPWASwTsConfig (tsConfig) {}
    },

    // https://v2.quasar.dev/quasar-cli-vite/developing-cordova-apps/configuring-cordova
    cordova: {},

    // https://v2.quasar.dev/quasar-cli-vite/developing-capacitor-apps/configuring-capacitor
    capacitor: {
      hideSplashscreen: true
    },

    // https://v2.quasar.dev/quasar-cli-vite/developing-electron-apps/configuring-electron
    electron: {
      // extendElectronMainConf (rolldownConf) {},
      // extendElectronPreloadConf (rolldownConf) {},
      // extendElectronPackageJson (pkgJson) {},

      // Electron preload scripts (if any) from /src-electron, WITHOUT file extension
      preloadScripts: ['electron-preload'],

      // specify the debugging port to use for the Electron app when running in development mode
      inspectPort: 5858,

      bundler: 'packager', // 'packager' or 'builder'

      packager: {
        // https://github.com/electron-userland/electron-packager/blob/master/docs/api.md#options

        // OS X / Mac App Store
        // appBundleId: '',
        // appCategoryType: '',
        // osxSign: '',
        // protocol: 'myapp://path',

        // Windows only
        // win32metadata: { ... }
      },

      builder: {
        // https://www.electron.build/configuration

        appId: 'frontend'
      }
    },

    // https://v2.quasar.dev/quasar-cli-vite/developing-browser-extensions/configuring-bex
    bex: {
      // extendBexScriptsConf (rolldownConf) {},
      // extendBexManifestJson (json) {},

      /**
       * The list of extra scripts (js/ts) not in your bex manifest that you want to
       * compile and use in your browser extension. Maybe dynamic use them?
       *
       * Each entry in the list should be a relative filename to /src-bex/
       *
       * @example [ 'my-script.ts', 'sub-folder/my-other-script.js' ]
       */
      extraScripts: []
    }
  }
});
