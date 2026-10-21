/**
 * Trasformazione delle costanti Go ImageCategory in Enum TypeScript.
 */
export enum ImageCategory {
  General = 'general',
  Avatar = 'avatar',
  Promo = 'promo',
  Banner = 'banner',
}

export interface ImageCategoryOption {
  label: string;
  value: ImageCategory;
  icon: string;
  color: string;
}

export const IMAGE_CATEGORY_OPTIONS: ImageCategoryOption[] = [
  { label: 'Generale', value: ImageCategory.General, icon: 'image', color: 'primary' },
  { label: 'Avatar', value: ImageCategory.Avatar, icon: 'account_circle', color: 'secondary' },
  { label: 'Promozione', value: ImageCategory.Promo, icon: 'campaign', color: 'deep-orange' },
  { label: 'Banner', value: ImageCategory.Banner, icon: 'view_carousel', color: 'teal' },
];

export function getCategoryMeta(category: string): { label: string; icon: string; color: string } {
  const found = IMAGE_CATEGORY_OPTIONS.find((opt) => opt.value === category);
  return found ?? { label: category || 'Generale', icon: 'image', color: 'grey-7' };
}
