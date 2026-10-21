import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import Client, { Local, admin as AdminNS, auth as AuthNS, profiles as ProfilesNS } from '@/encore/client';
import { useProfilesStore } from '@/stores/profiles-store';

export type Profile = AdminNS.Profile;
export type ProfileParams = AdminNS.ProfileParams;
export type RegisterParams = AdminNS.RegisterParams;
export type PersonalData = AdminNS.PersonalData;
export type PersonalDataParams = AdminNS.PersonalDataParams;
export type PersonalSetup = ProfilesNS.PersonalSetup;
export type PersonalSetupParams = ProfilesNS.PersonalSetupParams;
export type UpdateProfilePasswordParams = AdminNS.UpdateProfilePasswordParams;
export type RoleOption = AuthNS.RoleOption;
export type StatusOption = ProfilesNS.StatusOption;
export type MemberTypeOption = ProfilesNS.MemberTypeOption;

export const useAdminStore = defineStore('admin', () => {
  const profiles = ref<Profile[]>([]);
  const roles = ref<RoleOption[]>([]);
  const statuses = ref<StatusOption[]>([]);
  const memberTypes = ref<MemberTypeOption[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);

  const profilesStore = useProfilesStore();

  const client = new Client(Local, {
    auth: () => (profilesStore.token ? { Authorization: `Bearer ${profilesStore.token}` } : undefined),
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

  /** Fetches all profiles into `profiles`. Requires the caller to be logged in as an admin. */
  async function listProfiles(): Promise<Profile[]> {
    return withLoading(async () => {
      const res = await client.admin.ListProfiles();
      profiles.value = res.profiles;
      return res.profiles;
    });
  }

  /** Creates a new user profile with credentials. Requires the caller to be logged in as an admin. */
  async function insertProfile(params: RegisterParams): Promise<Profile> {
    return withLoading(async () => {
      const res = await client.admin.InsertProfile(params);
      profiles.value.push(res);
      return res;
    });
  }

  /** Fetches the valid roles and profile statuses into `roles` and `statuses`. */
  async function fetchSystemOptions(): Promise<void> {
    return withLoading(async () => {
      const res = await client.admin.GetSystemOptions();
      roles.value = res.roles;
      statuses.value = res.statuses;
      memberTypes.value = res.member_types || [];
    });
  }

  /** Statuses that an admin can manually set on a profile. */
  const updatableStatuses = computed(() => statuses.value.filter((s) => s.updatable));

  function replaceProfile(updated: Profile) {
    const idx = profiles.value.findIndex((p) => p.user_id === updated.user_id);
    if (idx !== -1) {
      profiles.value[idx] = updated;
    }
  }

  /** Updates any user's profile. Requires the caller to be logged in as an admin. */
  async function updateProfile(userID: string, params: ProfileParams): Promise<Profile> {
    return withLoading(async () => {
      const res = await client.admin.UpdateProfile(userID, params);
      replaceProfile(res);
      return res;
    });
  }

  /** Updates the role of any user's profile. */
  async function updateProfileRole(userID: string, role: string): Promise<Profile> {
    return withLoading(async () => {
      const res = await client.admin.UpdateProfileRole(userID, { role });
      replaceProfile(res);
      return res;
    });
  }

  /** Updates the status of any user's profile. */
  async function updateProfileStatus(userID: string, status: ProfilesNS.Status): Promise<Profile> {
    return withLoading(async () => {
      const res = await client.admin.UpdateProfileStatus(userID, { status });
      replaceProfile(res);
      return res;
    });
  }

  /** Updates the member type of any user's profile. */
  async function updateProfileMemberType(userID: string, memberType: string): Promise<Profile> {
    return withLoading(async () => {
      const res = await client.admin.UpdateProfileMemberType(userID, { member_type: memberType });
      replaceProfile(res);
      return res;
    });
  }

  /** Updates the password of any user's profile. */
  async function updateProfilePassword(userID: string, params: UpdateProfilePasswordParams): Promise<void> {
    return withLoading(async () => {
      await client.admin.UpdateProfilePassword(userID, params);
    });
  }

  /** Fetches the personal data of any user. Requires the caller to be logged in as an admin. */
  async function getPersonalData(userID: string): Promise<PersonalData> {
    return withLoading(async () => {
      return await client.admin.GetPersonalData(userID);
    });
  }

  /** Creates or replaces the personal data of any user. */
  async function upsertPersonalData(userID: string, params: PersonalDataParams): Promise<PersonalData> {
    return withLoading(async () => {
      return await client.admin.UpsertPersonalData(userID, params);
    });
  }

  /** Reads the selected user’s settings; the dialog handles loading errors locally. */
  async function getPersonalSetup(userID: string): Promise<PersonalSetup> {
    return client.profiles.GetUserPersonalSetup(userID);
  }

  async function upsertPersonalSetup(userID: string, params: PersonalSetupParams): Promise<PersonalSetup> {
    return withLoading(() => client.profiles.UpsertPersonalSetup(userID, params));
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
    profiles,
    roles,
    statuses,
    memberTypes,
    loading,
    error,
    // getters
    updatableStatuses,
    // actions
    listProfiles,
    fetchSystemOptions,
    insertProfile,
    updateProfile,
    updateProfileRole,
    updateProfileStatus,
    updateProfileMemberType,
    updateProfilePassword,
    getPersonalData,
    upsertPersonalData,
    getPersonalSetup,
    upsertPersonalSetup,
    uploadAvatar,
  };
});
