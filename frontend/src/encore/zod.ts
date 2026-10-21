// Code synced from src/encore/client.ts by frontend/tools/zod-sync.mjs.
// Existing field validators are preserved when this file is synced again.
import { z } from 'zod';

export const AdminPersonalDataParamsSchema = z.object({
  first_name: z.string().min(2).max(32),
  last_name: z.string().min(2).max(32),
  address: z.string().min(5).max(32),
  cap: z.string(),
  city: z.string().min(2).max(32),
  country: z.string(),
});

export const AdminProfileParamsSchema = z.object({
  display_name: z.string().min(2).max(32),
  avatar_url: z.string(),
});

export const AdminRegisterParamsSchema = z.object({
  email: z.string(),
  display_name: z.string(),
  avatar_url: z.string(),
  password: z.string(),
});

export const AdminUpdateProfileMemberTypeParamsSchema = z.object({
  member_type: z.string(),
});

export const AdminUpdateProfilePasswordParamsSchema = z.object({
  password: z.string(),
});

export const AdminUpdateProfileRoleParamsSchema = z.object({
  role: z.string(),
});

export const AdminUpdateProfileStatusParamsSchema = z.object({
  status: z.number(),
});

export const AdvertisingCreateAdvertisingParamsSchema = z.object({
  nome: z.string(),
  email: z.string(),
  status: z.string(),
});

export const AdvertisingCreatePromoParamsSchema = z.object({
  position: z.number(),
  section: z.string(),
  item: z.array(z.lazy(() => AdvertisingPromoItemSchema)),
  year: z.number(),
  weeks: z.array(z.number()),
  active_from: z.string(),
});

export const AdvertisingPromoBroadcasterSchema = z.object({
  short_id: z.string(),
  display_name: z.string(),
  image_url: z.string(),
});

export const AdvertisingPromoContentSchema = z.object({
  width: z.number(),
  height: z.number(),
  mobile_only: z.boolean(),
  broadcaster_short_id: z.string(),
  broadcaster: z.lazy(() => AdvertisingPromoBroadcasterSchema).optional(),
});

export const AdvertisingPromoItemSchema = z.object({
  title: z.string(),
  content: z.array(z.lazy(() => AdvertisingPromoContentSchema)),
});

export const AdvertisingUpdatePromoParamsSchema = z.object({
  position: z.number(),
  section: z.string(),
  item: z.array(z.lazy(() => AdvertisingPromoItemSchema)),
  year: z.number(),
  weeks: z.array(z.number()),
  active_from: z.string(),
});

export const AuthSetPasswordParamsSchema = z.object({
  current_password: z.string(),
  password: z.string(),
});

export const BroadcastersAddChatLogParamsSchema = z.object({
  type: z.string(),
  data: z.string(),
});

export const BroadcastersBroadcasterAttributeSchema = z.object({
  label: z.string(),
  value: z.string(),
});

export const BroadcastersBroadcasterHoursSchema = z.object({
  day: z.string(),
  hours: z.array(z.lazy(() => BroadcastersBroadcasterTimeSlotSchema)),
});

export const BroadcastersBroadcasterPhotoSchema = z.object({
  image_url: z.string(),
  alt: z.string(),
  box_x: z.number(),
  box_y: z.number(),
  description: z.lazy(() => ImagesImageDescriptionSchema),
});

export const BroadcastersBroadcasterTimeSlotSchema = z.object({
  from: z.string(),
  to: z.string(),
});

export const BroadcastersCreateBroadcasterParamsSchema = z.object({
  user_id: z.string(),
  display_name: z.string(),
  category: z.string(),
  category_color: z.string(),
  bio: z.string(),
  certified: z.boolean(),
  availability: z.string(),
  phone: z.string(),
  services: z.array(z.string()),
  languages: z.array(z.string()),
  attributes: z.array(z.lazy(() => BroadcastersBroadcasterAttributeSchema)),
  hours: z.array(z.lazy(() => BroadcastersBroadcasterHoursSchema)),
  photos: z.array(z.lazy(() => BroadcastersBroadcasterPhotoSchema)),
  enabled: z.boolean(),
  used: z.boolean(),
});

export const BroadcastersUpdateBroadcasterAttributesParamsSchema = z.object({
  attributes: z.array(z.lazy(() => BroadcastersBroadcasterAttributeSchema)),
});

export const BroadcastersUpdateBroadcasterHoursParamsSchema = z.object({
  hours: z.array(z.lazy(() => BroadcastersBroadcasterHoursSchema)),
});

export const BroadcastersUpdateBroadcasterParamsSchema = z.object({
  display_name: z.string(),
  category: z.string(),
  category_color: z.string(),
  bio: z.string(),
  certified: z.boolean(),
  availability: z.string(),
  phone: z.string(),
  services: z.array(z.string()),
  languages: z.array(z.string()),
  attributes: z.array(z.lazy(() => BroadcastersBroadcasterAttributeSchema)),
  hours: z.array(z.lazy(() => BroadcastersBroadcasterHoursSchema)),
  photos: z.array(z.lazy(() => BroadcastersBroadcasterPhotoSchema)),
  enabled: z.boolean(),
  used: z.boolean(),
});

export const BroadcastersUpdateBroadcasterPhotosParamsSchema = z.object({
  photos: z.array(z.lazy(() => BroadcastersBroadcasterPhotoSchema)),
});

export const ImagesImageDescriptionSchema = z.object({
  text: z.string(),
  box_x: z.number(),
  box_y: z.number(),
});

export const ImagesSaveImageParamsSchema = z.object({
  uuid: z.string(),
  description: z.string(),
  data: z.string(),
  content_type: z.string(),
  proportion: z.string(),
  width: z.number(),
  height: z.number(),
  // TODO: no longer present in Encore type images.SaveImageParams
  owner: z.string(),
  // TODO: no longer present in Encore type images.SaveImageParams
  approved: z.boolean(),
});

export const ProfilesLoginParamsSchema = z.object({
  user_email: z.string(),
  password: z.string(),
});

export const ProfilesPersonalDataParamsSchema = z.object({
  first_name: z.string(),
  last_name: z.string(),
  address: z.string(),
  cap: z.string(),
  city: z.string(),
  country: z.string(),
});

export const ProfilesPersonalSetupParamsSchema = z.object({
  can_broadcast: z.boolean(),
  room_name: z.string(),
});

export const ProfilesProfileParamsSchema = z.object({
  display_name: z.string(),
  avatar_url: z.string(),
});

export const RegistrationPasswordResetConfirmParamsSchema = z.object({
  password: z.string(),
});

export const RegistrationPasswordResetRequestParamsSchema = z.object({
  email: z.string(),
});

export const RegistrationRegisterParamsSchema = z.object({
  email: z.string().email(),
  display_name: z.string().min(2).max(32),
  avatar_url: z.string(),
  password: z.string().min(8),
});

export const VideochatCreateRoomParamsSchema = z.object({
  name: z.string(),
  max_participants: z.number(),
  empty_timeout: z.number(),
});

export const VideochatSendCommandParamsSchema = z.object({
  command: z.string(),
  target_id: z.string(),
  payload: z.string(),
});

export const VideochatSendMessageParamsSchema = z.object({
  client_message_id: z.string(),
  text: z.string(),
  kind: z.string(),
});

export const VideochatTokenParamsSchema = z.object({
  room_name: z.string(),
});
