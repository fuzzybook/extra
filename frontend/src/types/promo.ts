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

