import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { LocalStorage } from 'quasar';
import Client, { Local, auth as AuthNS, profiles as ProfilesNS } from '@/encore/client';

export type Profile = ProfilesNS.Profile;
export type ProfileParams = ProfilesNS.ProfileParams;
export type RegisterParams = ProfilesNS.RegisterParams;
export type LoginParams = ProfilesNS.LoginParams;
export type LoginResponse = ProfilesNS.LoginResponse;
export type Status = ProfilesNS.Status;
export type PersonalData = ProfilesNS.PersonalData;
export type PersonalDataParams = ProfilesNS.PersonalDataParams;
export type SetPasswordParams = AuthNS.SetPasswordParams;
export type MemberTypeOption = ProfilesNS.MemberTypeOption;

export const MEMBER_TYPES = {
  BROADCASTER: 'broadcaster',
  VISITOR: 'visitor',
  AFFILIATE: 'affiliate',
  ADVERTISER: 'advertiser',
  ADMIN: 'admin',
} as const;

export type MemberType = (typeof MEMBER_TYPES)[keyof typeof MEMBER_TYPES];

export const MEMBER_TYPE_OPTIONS: { label: string; value: MemberType; color: string }[] = [
  { label: 'Broadcaster', value: 'broadcaster', color: 'purple' },
  { label: 'Visitor', value: 'visitor', color: 'grey-7' },
  { label: 'Affiliate', value: 'affiliate', color: 'teal' },
  { label: 'Advertiser', value: 'advertiser', color: 'orange-8' },
  { label: 'Admin', value: 'admin', color: 'primary' },
];

export function memberTypeInfo(type: string): { label: string; color: string } {
  return MEMBER_TYPE_OPTIONS.find((opt) => opt.value === type) ?? { label: type || 'Unknown', color: 'grey' };
}

/** Display info for each Status value, mirroring profiles/status.go. */
export const STATUS_INFO: { label: string; color: string }[] = [
  { label: 'Active', color: 'positive' },
  { label: 'Inactive', color: 'grey' },
  { label: 'Suspended', color: 'warning' },
  { label: 'Deleted', color: 'negative' },
  { label: 'Pending', color: 'info' },
  { label: 'Waiting deletion', color: 'warning' },
  { label: 'Banned', color: 'negative' },
];

export function statusInfo(status: Status): { label: string; color: string } {
  return STATUS_INFO[status] ?? { label: `Unknown (${status})`, color: 'grey' };
}

const TOKEN_STORAGE_KEY = 'profiles.token';

export const useProfilesStore = defineStore('profiles', () => {
  const token = ref<string | null>(LocalStorage.getItem(TOKEN_STORAGE_KEY));
  const profile = ref<Profile | null>(null);
  const personalData = ref<PersonalData | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);

  const isAuthenticated = computed(() => token.value !== null);

  const client = new Client(Local, {
    auth: () => (token.value ? { Authorization: `Bearer ${token.value}` } : undefined),
  });

  async function withLoading<T>(fn: () => Promise<T>): Promise<T> {
    loading.value = true;
    error.value = null;
    try {
      return await fn();
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err);
      throw err;
    } finally {
      loading.value = false;
    }
  }

  function setToken(newToken: string | null) {
    token.value = newToken;
    if (newToken) {
      LocalStorage.set(TOKEN_STORAGE_KEY, newToken);
    } else {
      LocalStorage.remove(TOKEN_STORAGE_KEY);
    }
  }

  /** Logs in and fetches the resulting profile into `profile`. */
  async function login(params: LoginParams): Promise<LoginResponse> {
    return withLoading(async () => {
      const res = await client.profiles.Login(params);
      setToken(res.token);
      await me();
      return res;
    });
  }

  /** Clears the local session. Does not call the auth service's Logout endpoint. */
  function logout() {
    setToken(null);
    profile.value = null;
    personalData.value = null;
  }

  /** Fetches the authenticated caller's profile, or null if not authenticated. */
  async function me(): Promise<Profile | null> {
    return withLoading(async () => {
      const res = await client.profiles.Me();
      profile.value = res;
      return res;
    });
  }

  /** Registers a new profile (and credentials) and returns it. */
  async function register(params: RegisterParams): Promise<Profile> {
    return withLoading(async () => {
      return await client.profiles.Insert(params);
    });
  }

  /** Updates the authenticated caller's own profile. */
  async function update(params: ProfileParams): Promise<Profile> {
    return withLoading(async () => {
      const res = await client.profiles.Update(params);
      profile.value = res;
      return res;
    });
  }

  /** Deletes the authenticated caller's own profile and clears the session. */
  async function remove(): Promise<void> {
    return withLoading(async () => {
      await client.profiles.Delete();
      logout();
    });
  }

  /** Fetches the authenticated caller's own personal data into `personalData`. */
  async function fetchPersonalData(): Promise<PersonalData> {
    return withLoading(async () => {
      const res = await client.profiles.GetPersonalData();
      personalData.value = res;
      return res;
    });
  }

  /** Creates or replaces the authenticated caller's own personal data. */
  async function savePersonalData(params: PersonalDataParams): Promise<PersonalData> {
    return withLoading(async () => {
      const res = await client.profiles.UpsertPersonalData(params);
      personalData.value = res;
      return res;
    });
  }

  /** Changes the authenticated caller's password after verifying the current one. */
  async function setPassword(params: SetPasswordParams): Promise<void> {
    return withLoading(async () => {
      await client.auth.SetPassword(params);
    });
  }

  /** Uploads an avatar image and returns its public URL. */
  async function uploadAvatar(image: Blob): Promise<string> {
    return withLoading(async () => {
      const resp = await client.profiles.UploadAvatar('POST', image);
      if (!resp.ok) {
        throw new Error(`avatar upload failed (${resp.status})`);
      }
      const data = (await resp.json()) as { url: string };
      return data.url;
    });
  }

  return {
    // state
    token,
    profile,
    personalData,
    loading,
    error,
    // getters
    isAuthenticated,
    // actions
    login,
    logout,
    me,
    register,
    update,
    remove,
    fetchPersonalData,
    savePersonalData,
    setPassword,
    uploadAvatar,
  };
});
