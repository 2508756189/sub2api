import { describe, expect, it } from 'vitest'
import {
  buildEnsembleGroupPayload,
  deriveEnsembleModelOptions,
  findSharedEnsembleChannel,
  getEnsembleSourceGroups,
  planEnsembleMemberReconciliation,
  validateEnsembleDraft
} from '@/utils/ensemble'

describe('Ensemble configuration helpers', () => {
  it('only exposes active non-ensemble groups as account sources', () => {
    const groups = [
      { id: 1, name: 'OpenAI 主组', platform: 'openai', status: 'active' },
      { id: 2, name: '已停用组', platform: 'anthropic', status: 'inactive' },
      { id: 3, name: '旧 Ensemble', platform: 'ensemble', status: 'active' },
      { id: 4, name: 'Composite', platform: 'composite', status: 'active' }
    ]

    expect(getEnsembleSourceGroups(groups)).toEqual([groups[0]])
  })

  it('derives models only from active channel pricing attached to selected groups', () => {
    const channels = [
      {
        status: 'active',
        group_ids: [1],
        model_pricing: [{ models: ['gpt-5.6', 'claude-sonnet-4.5'] }]
      },
      {
        status: 'inactive',
        group_ids: [1],
        model_pricing: [{ models: ['should-not-appear'] }]
      },
      {
        status: 'active',
        group_ids: [2],
        model_pricing: [{ models: ['other-group-model'] }]
      }
    ]

    expect(deriveEnsembleModelOptions([1], channels).map(option => option.model))
      .toEqual(['claude-sonnet-4.5', 'gpt-5.6'])
  })

  it('keeps the concrete pricing platform for opaque model names', () => {
    const groups = [{ id: 1, name: 'OpenAI', platform: 'openai', status: 'active' }]
    const channels = [{
      status: 'active',
      group_ids: [1],
      model_pricing: [
        { platform: 'anthropic', models: ['claude-sonnet-4.5'] },
        { platform: 'openai', models: ['qwen3.7-max', 'kimi-k2.7-code'] }
      ]
    }]

    expect(deriveEnsembleModelOptions([1], channels, groups)).toEqual([
      { model: 'kimi-k2.7-code', platform: 'openai' },
      { model: 'qwen3.7-max', platform: 'openai' }
    ])
  })

  it('rejects incomplete drafts before any group can be created', () => {
    expect(validateEnsembleDraft({
      name: '',
      sourceGroupIds: [],
      proposers: [],
      minProposers: 1
    })).toBe('name-required')

    expect(validateEnsembleDraft({
      name: '新 Ensemble',
      sourceGroupIds: [1],
      proposers: ['gpt-5.6'],
      minProposers: 2
    })).toBe('at-least-two-proposers')
  })

  it('builds a native ensemble group payload with copied account sources', () => {
    expect(buildEnsembleGroupPayload({
      name: '研发 Ensemble',
      description: '组内聚合',
      sourceGroupIds: [4, 2],
      rateMultiplier: 1.2
    })).toMatchObject({
      name: '研发 Ensemble',
      description: '组内聚合',
      platform: 'ensemble',
      rate_multiplier: 1.2,
      copy_accounts_from_group_ids: [4, 2]
    })
  })

  it('finds one active billing channel shared by every source group', () => {
    const channels = [
      { id: 10, name: 'tokenrhythm', status: 'active', group_ids: [1, 2] },
      { id: 11, name: 'other', status: 'active', group_ids: [2, 3] },
      { id: 12, name: 'inactive', status: 'inactive', group_ids: [1, 2] }
    ]

    expect(findSharedEnsembleChannel([1, 2], channels)).toMatchObject({ id: 10 })
    expect(findSharedEnsembleChannel([1, 3], channels)).toBeNull()
    expect(findSharedEnsembleChannel([], channels)).toBeNull()
  })

  it('reuses an existing proposer row when replacing a member at the six-member limit', () => {
    const existing = Array.from({ length: 6 }, (_, index) => ({
      id: index + 1,
      role: 'proposer' as const,
      model: `model-${index + 1}`,
      platform: 'openai',
      priority: 100 + index,
      enabled: true
    }))
    const desired = [
      ...existing.slice(0, 5).map(({ role, model, platform, priority }) => ({ role, model, platform, priority })),
      { role: 'proposer' as const, model: 'replacement', platform: 'openai', priority: 105 }
    ]

    const plan = planEnsembleMemberReconciliation(existing, desired)

    expect(plan.creates).toEqual([])
    expect(plan.deletes).toEqual([])
    expect(plan.updates).toContainEqual({
      id: 6,
      member: { role: 'proposer', model: 'replacement', platform: 'openai', priority: 105 }
    })
  })
})
