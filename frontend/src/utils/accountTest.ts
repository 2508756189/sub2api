import type { ClaudeModel } from '@/types'

const ACCOUNT_ROUTING_MODELS = new Set(['composite'])

/** Account tests call one concrete upstream account, never a routing group. */
export function filterAccountConnectivityModels<T extends Pick<ClaudeModel, 'id'>>(models: T[]): T[] {
  return models.filter(model => !ACCOUNT_ROUTING_MODELS.has(model.id.trim().toLowerCase()))
}
