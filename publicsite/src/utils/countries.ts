// Stati europei accettati alla registrazione (ISO 3166-1 alpha-2, più XK per
// il Kosovo): stessa lista del backend (profiles.signupCountries). I nomi
// vengono tradotti con Intl.DisplayNames nella lingua del sito.
const COUNTRY_CODES = (
  'AD AL AT BA BE BG BY CH CY CZ DE DK EE ES FI FR GB GR HR HU IE IS IT LI LT LU ' +
  'LV MC MD ME MK MT NL NO PL PT RO RS RU SE SI SK SM UA VA XK'
).split(' ');

export interface CountryOption {
  code: string;
  name: string;
}

/** Paesi ordinati per nome nella lingua indicata. */
export function countryOptions(locale: string): CountryOption[] {
  let names: Intl.DisplayNames | null = null;
  try {
    names = new Intl.DisplayNames([locale], { type: 'region' });
  } catch {
    names = null;
  }
  return COUNTRY_CODES.map((code) => ({ code, name: names?.of(code) ?? code })).sort((a, b) =>
    a.name.localeCompare(b.name, locale),
  );
}
