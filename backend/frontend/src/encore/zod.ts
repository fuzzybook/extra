// Code synced from src/encore/client.ts by frontend/tools/zod-sync.mjs.
// Existing field validators are preserved when this file is synced again.
import { z } from 'zod';

export const AdminPersonalDataParamsSchema = z.object({
  first_name: z.string().min(2).max(32),
  last_name: z.string().min(2).max(32),
  address: z.string().min(5).max(32),
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

export const AdminUpdateProfileArtistParamsSchema = z.object({
  is_artist: z.boolean(),
});

export const AdminUpdateProfileRoleParamsSchema = z.object({
  role: z.string(),
});

export const AdminUpdateProfileStatusParamsSchema = z.object({
  status: z.number(),
});

export const AuthSetPasswordParamsSchema = z.object({
  password: z.string(),
});

export const ProfilesLoginParamsSchema = z.object({
  user_email: z.string(),
  password: z.string(),
});

export const ProfilesPersonalDataParamsSchema = z.object({
  first_name: z.string(),
  last_name: z.string(),
  address: z.string(),
  city: z.string(),
  country: z.string(),
});

export const ProfilesProfileParamsSchema = z.object({
  display_name: z.string(),
  avatar_url: z.string(),
});

export const ProfilesRegisterParamsSchema = z.object({
  email: z.string(),
  display_name: z.string(),
  avatar_url: z.string(),
  password: z.string(),
});

export const RegistrationRegisterParamsSchema = z.object({
  email: z.string().email(),
  display_name: z.string().min(2).max(32),
  avatar_url: z.string(),
  password: z.string().min(8),
});
