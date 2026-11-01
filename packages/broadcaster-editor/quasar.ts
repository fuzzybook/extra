// Collegamento del pacchetto a un'app Quasar (vedi README): da usare in
// quasar.config.ts con la cartella dell'app, come packages/webchat.
import path from 'node:path';

/** Dipendenze peer, risolte sempre dalla cartella dell'app. */
const peers = ['vue', 'quasar', 'vue-i18n'];

interface ViteConf {
  resolve?: { dedupe?: string[] | undefined } | undefined;
  server?: { fs?: { allow?: string[] | undefined } | undefined } | undefined;
}

interface TsConf {
  compilerOptions?: { paths?: Record<string, string[]> | undefined } | undefined;
}

export function broadcasterEditorSetup(appDir: string) {
  const packageDir = path.resolve(appDir, '../packages/broadcaster-editor');
  return {
    alias: { '@extra/broadcaster-editor': path.join(packageDir, 'src') },

    extendViteConf(viteConf: ViteConf) {
      viteConf.resolve = { ...viteConf.resolve, dedupe: [...new Set([...(viteConf.resolve?.dedupe ?? []), ...peers])] };
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
