// Stati europei accettati alla registrazione (ISO 3166-1 alpha-2, più XK per
// il Kosovo): stessa lista del backend (profiles.signupCountries). I nomi
// vengono tradotti con Intl.DisplayNames nella lingua dell'interfaccia.
const EUROPEAN_COUNTRY_CODES = (
  'AD AL AT BA BE BG BY CH CY CZ DE DK EE ES FI FR GB GR HR HU IE IS IT LI LT LU ' +
  'LV MC MD ME MK MT NL NO PL PT RO RS RU SE SI SK SM UA VA XK'
).split(' ');

/** Stati europei come opzioni { label, value }, ordinati per nome. */
export function europeanCountryOptions(locale: string): { label: string; value: string }[] {
  let names: Intl.DisplayNames | null = null;
  try {
    names = new Intl.DisplayNames([locale], { type: 'region' });
  } catch {
    names = null;
  }
  return EUROPEAN_COUNTRY_CODES.map((value) => ({ value, label: names?.of(value) ?? value })).sort((a, b) =>
    a.label.localeCompare(b.label, locale),
  );
}
