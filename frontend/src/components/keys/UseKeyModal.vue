<template>
  <ClientAccessCenterDialog v-bind="$props" @close="emit('close')" />
</template>

<script setup lang="ts">
import type { GroupPlatform } from '@/types'
import ClientAccessCenterDialog from '@/tokenport/access-center/ClientAccessCenterDialog.vue'

defineProps<{
  show: boolean
  apiKey: string
  baseUrl: string
  platform: GroupPlatform | null
  groupId?: number | null
  allowMessagesDispatch?: boolean
  initialMode?: 'direct' | 'ccs'
  keyName?: string
}>()

const emit = defineEmits<{ (event: 'close'): void }>()
</script>
