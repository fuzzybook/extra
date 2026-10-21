<template>
  <q-dialog v-model="open" @show="focusFirstField">
    <q-card style="min-width: 350px">
      <q-card-section>
        <div class="text-h6">{{ t('actions.updateMemberType') }}</div>
      </q-card-section>

      <q-card-section v-if="profile" class="q-gutter-md">
        <div class="text-subtitle1">{{ profile.email }}</div>

        <q-select
          ref="memberTypeSelectRef"
          v-model="selectedMemberType"
          :options="MEMBER_TYPE_OPTIONS"
          option-label="label"
          option-value="value"
          emit-value
          map-options
          :label="t('fields.memberType')"
        >
          <template #option="scope">
            <q-item v-bind="scope.itemProps">
              <q-item-section>
                <q-item-label>
                  <q-badge :color="scope.opt.color" class="q-mr-xs">{{ scope.opt.label }}</q-badge>
                </q-item-label>
              </q-item-section>
            </q-item>
          </template>
        </q-select>

        <q-banner v-if="adminStore.error" class="bg-negative text-white" rounded>
          {{ adminStore.error }}
        </q-banner>
      </q-card-section>

      <q-card-actions align="right">
        <q-btn v-close-popup flat :label="t('actions.cancel')" />
        <q-btn color="primary" :label="t('actions.save')" :loading="adminStore.loading" @click="save" />
      </q-card-actions>
    </q-card>
  </q-dialog>
</template>

<script setup lang="ts">
import { nextTick, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { Dialog, Notify } from 'quasar';
import { useAdminStore, type Profile } from '@/stores/admin-store';
import { MEMBER_TYPE_OPTIONS, type MemberType } from '@/stores/profiles-store';

type Focusable = { focus: () => void };

const props = defineProps<{
  profile: Profile | null;
}>();

const open = defineModel<boolean>({ required: true });
const adminStore = useAdminStore();
const { t } = useI18n();
const selectedMemberType = ref<MemberType>('visitor');
const memberTypeSelectRef = ref<Focusable | null>(null);

watch(
  () => [open.value, props.profile] as const,
  ([isOpen, profile]) => {
    if (isOpen && profile) {
      selectedMemberType.value = (profile.member_type as MemberType) || 'visitor';
    }
  },
  { immediate: true },
);

async function performUpdate() {
  if (!props.profile) return;
  try {
    await adminStore.updateProfileMemberType(props.profile.user_id, selectedMemberType.value);
    Notify.create({ type: 'positive', message: t('admin.notifications.memberTypeUpdated') });
    open.value = false;
  } catch {
    // adminStore.error already holds the failure message
  }
}

async function save() {
  if (!props.profile) return;

  const wasBroadcaster = props.profile.member_type === 'broadcaster';
  const isChangingAway = selectedMemberType.value !== 'broadcaster';

  if (wasBroadcaster && isChangingAway) {
    Dialog.create({
      title: 'Conferma modifica tipo membro',
      message:
        'L\'utente è attualmente un Broadcaster. Modificando il tipo membro, il profilo broadcaster associato verrà disabilitato e disattivato (used=false, enabled=false). Vuoi continuare?',
      cancel: {
        label: t('actions.cancel'),
        flat: true,
      },
      ok: {
        label: t('actions.confirm') || 'Conferma',
        color: 'warning',
      },
      persistent: true,
    }).onOk(async () => {
      await performUpdate();
    });
    return;
  }

  await performUpdate();
}

async function focusFirstField() {
  await nextTick();
  memberTypeSelectRef.value?.focus();
}
</script>
