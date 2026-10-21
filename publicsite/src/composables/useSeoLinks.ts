import { useMeta } from 'quasar';
import { useRoute } from 'vue-router';

export function useSeoLinks() {
  const route = useRoute();
  const baseUrl = 'https://il-tuo-sito.com';
  const supportedLocales = ['it', 'fr', 'de', 'en'] as const;

  // Pulisce il path corrente rimuovendo il prefisso locale vecchio
  const getCleanPath = () => {
    const pathSegments = route.path.split('/').filter(Boolean);
    const firstSegment = pathSegments[0];
    if (firstSegment && (supportedLocales as readonly string[]).includes(firstSegment)) {
      pathSegments.shift(); // Rimuove il codice lingua corrente
    }
    return pathSegments.join('/');
  };

  const cleanPath = getCleanPath();
  const suffix = cleanPath ? `/${cleanPath}` : '';

  // Costruisci il dizionario di tag alternativi per Quasar useMeta
  const linkLinks: Record<string, Record<string, string>> = {};

  supportedLocales.forEach((lang) => {
    linkLinks[`alternate-${lang}`] = {
      rel: 'alternate',
      hreflang: lang,
      href: `${baseUrl}/${lang}${suffix}`,
    };
  });

  // Aggiungi l'x-default (consigliato da Google per utenti generici)
  linkLinks['alternate-x-default'] = {
    rel: 'alternate',
    hreflang: 'x-default',
    href: `${baseUrl}/it${suffix}`,
  };

  // Applica i metadati all'head della pagina
  useMeta({
    link: linkLinks,
  });
}
