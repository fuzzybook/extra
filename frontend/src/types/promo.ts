/**
 * Trasformazione delle const Go PromoLeft, PromoRight1, PromoRight2 in Enum TypeScript
 */
export enum PromoPosition {
  PromoLeft = 0,
  PromoRight1 = 1,
  PromoRight2 = 2,
}

/**
 * Trasformazione delle sezioni Promo da Go const a TypeScript enum
 */
export enum PromoLeftSection {
  Week = 'week',
  Raccomended = 'raccomended',
  Discover = 'discover',
}

export enum PromoRight1Section {
  Show = 'show',
  Service = 'service',
}

export enum PromoRight2Section {
  Top = 'top',
  Selected = 'selected',
}

export type PromoSection = PromoLeftSection | PromoRight1Section | PromoRight2Section;

// Mapping etichette e colori per l'enum PromoPosition
export const PROMO_POSITION_MAP: Record<PromoPosition, { label: string; color: string; icon: string }> = {
  [PromoPosition.PromoLeft]: { label: 'Promo Left', color: 'primary', icon: 'view_sidebar' },
  [PromoPosition.PromoRight1]: { label: 'Promo Right 1', color: 'teal', icon: 'view_agenda' },
  [PromoPosition.PromoRight2]: { label: 'Promo Right 2', color: 'deep-purple', icon: 'view_column' },
};

export interface PromoSectionOption {
  label: string;
  value: string;
  color?: string;
}

// Opzioni di sezione per ciascuna PromoPosition (usate nel componente <q-select>)
export const PROMO_SECTION_OPTIONS: Record<PromoPosition, PromoSectionOption[]> = {
  [PromoPosition.PromoLeft]: [
    { label: 'Settimana (week)', value: PromoLeftSection.Week, color: 'blue' },
    { label: 'Raccomandati (raccomended)', value: PromoLeftSection.Raccomended, color: 'cyan' },
    { label: 'Scopri (discover)', value: PromoLeftSection.Discover, color: 'indigo' },
  ],
  [PromoPosition.PromoRight1]: [
    { label: 'Show (show)', value: PromoRight1Section.Show, color: 'teal' },
    { label: 'Servizi (service)', value: PromoRight1Section.Service, color: 'green' },
  ],
  [PromoPosition.PromoRight2]: [
    { label: 'Top (top)', value: PromoRight2Section.Top, color: 'purple' },
    { label: 'Selezionati (selected)', value: PromoRight2Section.Selected, color: 'deep-purple' },
  ],
};

export function getPositionMeta(position: number): { label: string; color: string; icon: string } {
  return PROMO_POSITION_MAP[position as PromoPosition] ?? { label: `Posizione ${position}`, color: 'grey-7', icon: 'campaign' };
}

export function getSectionMeta(position: number, section: string): { label: string; color: string } {
  const options = PROMO_SECTION_OPTIONS[position as PromoPosition];
  if (options) {
    const found = options.find((opt) => opt.value === section);
    if (found) {
      return { label: found.label, color: found.color ?? 'grey-7' };
    }
  }
  return { label: section || 'N/A', color: 'grey-6' };
}

export interface PromoImageSize {
  width: number;
  height: number;
  proportion: string;
}

// Dimensioni richieste per le immagini promo (PromoLeft, PromoRight1, PromoRight2 definite in promo.go)
// Se height == 0 si usa la proporzione, se diverso da 0 si usa l'altezza
export const PROMO_IMAGE_SIZES: Record<PromoPosition, PromoImageSize> = {
  [PromoPosition.PromoLeft]: { width: 230, height: 0, proportion: '2:3' },
  [PromoPosition.PromoRight1]: { width: 230, height: 0, proportion: '2:3' },
  [PromoPosition.PromoRight2]: { width: 230, height: 345, proportion: '' },
};

export function getPromoImageSize(position: number): PromoImageSize {
  return PROMO_IMAGE_SIZES[position as PromoPosition] ?? { width: 230, height: 0, proportion: '2:3' };
}

export function parseProportion(proportion: string): { widthRatio: number; heightRatio: number } {
  const parts = proportion.split(':').map((s) => parseFloat(s.trim()));
  const w = parts[0];
  const h = parts[1];
  if (parts.length === 2 && w !== undefined && h !== undefined && !isNaN(w) && !isNaN(h) && w > 0 && h > 0) {
    return { widthRatio: w, heightRatio: h };
  }
  return { widthRatio: 2, heightRatio: 3 };
}

/**
 * Se height != 0 si utilizza l'altezza specificata.
 * Se height == 0 si calcola l'altezza dalla proporzione.
 */
export function calculatePromoDimensions(
  width: number,
  height: number = 0,
  proportion: string = '2:3'
): { width: number; height: number } {
  const w = width && width > 0 ? width : 230;
  if (height !== undefined && height !== 0) {
    return { width: w, height };
  }
  if (proportion && proportion.trim().length > 0) {
    const { widthRatio, heightRatio } = parseProportion(proportion);
    const calculatedHeight = Math.round((w / widthRatio) * heightRatio);
    return { width: w, height: calculatedHeight };
  }
  return { width: w, height: Math.round((w * 3) / 2) };
}

