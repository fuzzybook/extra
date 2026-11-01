// Collegamento del pacchetto a un'app Quasar (vedi README): da usare in
// quasar.config.ts con la cartella dell'app.
import path from 'node:path';

/** Dipendenze peer, risolte sempre dalla cartella dell'app. */
const peers = ['vue', 'quasar', 'vue-i18n', 'livekit-client'];

interface ViteConf {
  resolve?: { dedupe?: string[] | undefined } | undefined;
  server?: { fs?: { allow?: string[] | undefined } | undefined } | undefined;
}

interface TsConf {
  compilerOptions?: { paths?: Record<string, string[]> | undefined } | undefined;
}

/**
 * Host LiveKit da permettere in connect-src della CSP: server locale, LiveKit
 * Cloud e quello di QCLI_LIVEKIT_URL (produzione).
 */
function livekitSources(): string[] {
  const sources = [
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
    sources.push(`${secure ? 'https' : 'http'}://${url.host}`, `${secure ? 'wss' : 'ws'}://${url.host}`);
  }
  return sources;
}

export function webchatSetup(appDir: string) {
  const packageDir = path.resolve(appDir, '../packages/webchat');
  return {
    /** Valore per htmlVariables, da usare in connect-src di index.html. */
    livekitConnectSources: livekitSources().join(' '),

    alias: { '@extra/webchat': path.join(packageDir, 'src') },

    extendViteConf(viteConf: ViteConf) {
      viteConf.resolve = { ...viteConf.resolve, dedupe: [...(viteConf.resolve?.dedupe ?? []), ...peers] };
      const fs = viteConf.server?.fs;
      viteConf.server = {
        ...viteConf.server,
        fs: { ...fs, allow: [...(fs?.allow ?? [appDir]), packageDir] },
      };
    },

    // I file del pacchetto non hanno node_modules: TypeScript cerca le
    // dipendenze peer in quella dell'app (percorsi relativi a .quasar/).
    extendTsConfig(tsConfig: TsConf) {
      const paths = tsConfig.compilerOptions?.paths ?? {};
      for (const name of peers) {
        paths[name] = [`./../node_modules/${name}`];
      }
      tsConfig.compilerOptions = { ...tsConfig.compilerOptions, paths };
    },
  };
}
