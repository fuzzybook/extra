// Tipi del backend usati dai dialoghi, uguali a quelli del client Encore
// generato (broadcasters, config, images). Il pacchetto non importa il client
// delle app: l'adattatore passato da ogni app viene controllato su questi tipi.

export interface ImageDescription {
  text: string;
  box_x: number;
  box_y: number;
}

export interface BroadcasterAttribute {
  label: string;
  value: string;
}

export interface BroadcasterTimeSlot {
  from: string;
  to: string;
}

export interface BroadcasterHours {
  day: string;
  hours: BroadcasterTimeSlot[];
}

export interface BroadcasterPhoto {
  image_url: string;
  alt: string;
  box_x: number;
  box_y: number;
  description: ImageDescription;
}

export interface BroadcasterImage {
  width: number;
  height: number;
  proportion: string;
}

export interface Broadcaster {
  id: string;
  short_id: string;
  user_id: string;
  display_name: string;
  category: string;
  category_color: string;
  bio: string;
  certified: boolean;
  availability: string;
  phone: string;
  services: string[];
  languages: string[];
  attributes: BroadcasterAttribute[];
  hours: BroadcasterHours[];
  photos: BroadcasterPhoto[];
  enabled: boolean;
  used: boolean;
  created_at: string;
  updated_at: string;
}

export type UpdateBroadcasterParams = Omit<Broadcaster, 'id' | 'short_id' | 'user_id' | 'created_at' | 'updated_at'>;
export type CreateBroadcasterParams = UpdateBroadcasterParams & { user_id: string };

export interface Category {
  name: string;
  color: string;
}

export interface Image {
  uuid: string;
  owner: string;
  description: ImageDescription;
  approved: boolean;
  created_at: string;
  url: string;
  proportion: string;
  width: number;
  height: number;
}

export interface SaveImageParams {
  uuid: string;
  description: ImageDescription;
  data: string;
  content_type: string;
  proportion: string;
  width: number;
  height: number;
}
