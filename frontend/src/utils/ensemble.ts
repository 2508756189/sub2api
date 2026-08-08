import type { CreateGroupRequest, GroupPlatform } from '@/types'

export interface EnsembleSourceGroup {
  id: number
  name: string
  platform: string
  status?: string
}

export interface EnsembleChannelModelPricing {
  platform?: string
	models?: string[]
}

export interface EnsembleModelOption {
  model: string
  platform: string
}

export interface EnsembleChannel {
  id?: number
  name?: string
  status?: string
  group_ids?: number[]
  model_pricing?: EnsembleChannelModelPricing[]
}

export interface EnsembleDraftValidation {
  name: string
  sourceGroupIds: number[]
  proposers: string[]
  minProposers: number
}

export interface EnsembleGroupDraft {
  name: string
  description: string
  sourceGroupIds: number[]
  rateMultiplier: number
}

export interface EnsembleMemberRow {
  id: number
  role: 'proposer' | 'aggregator'
  model: string
  platform?: string
  priority: number
  enabled: boolean
}

export interface EnsembleMemberDesired {
  role: 'proposer' | 'aggregator'
  model: string
  platform: string
  priority: number
}

export interface EnsembleMemberReconciliationPlan {
  updates: Array<{ id: number; member: EnsembleMemberDesired }>
  creates: EnsembleMemberDesired[]
  deletes: number[]
}

/** Reuse unmatched rows before creating new ones at the hard proposer limit. */
export function planEnsembleMemberReconciliation(
  existing: EnsembleMemberRow[],
  desired: EnsembleMemberDesired[]
): EnsembleMemberReconciliationPlan {
  const used = new Set<number>()
  const updates: EnsembleMemberReconciliationPlan['updates'] = []
  const creates: EnsembleMemberDesired[] = []

  for (const item of desired) {
    const exact = existing.find(member =>
      !used.has(member.id) &&
      member.role === item.role &&
      member.model === item.model &&
      (member.platform ?? '') === item.platform
    )
    if (exact) {
      used.add(exact.id)
      updates.push({ id: exact.id, member: item })
      continue
    }

    const reusable = existing.find(member => !used.has(member.id) && member.role === item.role)
    if (reusable) {
      used.add(reusable.id)
      updates.push({ id: reusable.id, member: item })
    } else {
      creates.push(item)
    }
  }

  return {
    updates,
    creates,
    deletes: existing.filter(member => !used.has(member.id)).map(member => member.id)
  }
}

/** Return only groups whose accounts can be copied into an ensemble group. */
export function getEnsembleSourceGroups<T extends EnsembleSourceGroup>(groups: T[]): T[] {
  return groups.filter(group =>
    group.status !== 'inactive' &&
    group.platform !== 'ensemble' &&
    group.platform !== 'composite'
  )
}

/**
 * Resolve models from channel pricing, scoped to the selected source groups.
 * Platform default model lists are intentionally excluded: they are not proof
 * that a model is configured or billable for the selected accounts.
 */
export function deriveEnsembleModels(
  selectedGroupIds: number[],
  channels: EnsembleChannel[],
  sourceGroups: EnsembleSourceGroup[] = []
): string[] {
  return deriveEnsembleModelOptions(selectedGroupIds, channels, sourceGroups).map(option => option.model)
}

/** Resolve each public model to the concrete platform that prices it. */
export function deriveEnsembleModelOptions(
  selectedGroupIds: number[],
  channels: EnsembleChannel[],
  sourceGroups: EnsembleSourceGroup[] = []
): EnsembleModelOption[] {
  const selected = new Set(selectedGroupIds)
  const sourcePlatforms = new Set(
    sourceGroups
      .filter(group => selected.has(group.id))
      .map(group => group.platform)
      .filter(Boolean)
  )
  const models = new Map<string, string>()

  for (const channel of channels) {
    if (channel.status === 'inactive' || !(channel.group_ids ?? []).some(id => selected.has(id))) {
      continue
    }
    for (const pricing of channel.model_pricing ?? []) {
      const platform = pricing.platform?.trim().toLowerCase() ?? ''
      if (sourcePlatforms.size > 0 && platform && !sourcePlatforms.has(platform)) continue
      for (const model of pricing.models ?? []) {
        const normalized = model.trim()
        if (normalized && !models.has(normalized)) models.set(normalized, platform)
      }
    }
  }

  return [...models.entries()]
    .map(([model, platform]) => ({ model, platform }))
    .sort((a, b) => a.model.localeCompare(b.model))
}

/**
 * A native ensemble group has one channel because channel pricing is scoped by
 * group and the gateway must have one unambiguous billing policy. Source
 * groups from different channels are therefore rejected instead of merged.
 */
export function findSharedEnsembleChannel<T extends EnsembleChannel>(
  selectedGroupIds: number[],
  channels: T[]
): T | null {
  if (selectedGroupIds.length === 0) return null
  return channels.find(channel =>
    channel.status !== 'inactive' &&
    selectedGroupIds.every(id => (channel.group_ids ?? []).includes(id))
  ) ?? null
}

export function validateEnsembleDraft(draft: EnsembleDraftValidation): string | null {
  if (!draft.name.trim()) return 'name-required'
  if (draft.sourceGroupIds.length === 0) return 'source-group-required'
  if (draft.proposers.length < 2) return 'at-least-two-proposers'
  if (draft.minProposers < 1 || draft.minProposers > draft.proposers.length) {
    return 'invalid-min-proposers'
  }
  return null
}

export function buildEnsembleGroupPayload(draft: EnsembleGroupDraft): CreateGroupRequest {
  return {
    name: draft.name.trim(),
    description: draft.description.trim() || null,
    platform: 'ensemble' as GroupPlatform,
    rate_multiplier: Number.isFinite(draft.rateMultiplier) && draft.rateMultiplier > 0
      ? draft.rateMultiplier
      : 1,
    copy_accounts_from_group_ids: [...draft.sourceGroupIds],
    models_list_config: {
      enabled: true,
      models: ['ensemble']
    }
  }
}
