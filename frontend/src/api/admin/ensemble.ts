import { apiClient, buildGatewayUrl } from '../client'

export type EnsembleMemberRole = 'proposer' | 'aggregator'

export interface EnsembleProposer {
  id: number
  group_id: number
  role: EnsembleMemberRole
  model: string
  platform: string
  priority: number
  enabled: boolean
  created_at?: string
  updated_at?: string
}

export interface EnsembleConfig {
  aggregator_enabled: boolean
  min_proposers: number
  timeout_seconds: number
  max_tokens: number
  expose_metadata: boolean
  source_group_ids?: number[]
}

export interface EnsembleMemberStat {
  model: string
  role: EnsembleMemberRole
  status: 'ok' | 'failed' | string
  duration_ms: number
  prompt_tokens?: number
  completion_tokens?: number
  content?: string
  cost?: number
  cost_source?: string
  error?: string
}

export interface EnsembleTestResponse {
  id?: string
  model?: string
  choices?: Array<{ message?: { content?: string } }>
  usage?: {
    prompt_tokens?: number
    completion_tokens?: number
    total_tokens?: number
  }
  ensemble_metadata?: {
    members?: EnsembleMemberStat[]
    members_total?: number
    members_succeeded?: number
    aggregated?: boolean
    duration_ms?: number
    proposer_results?: EnsembleMemberStat[]
    aggregator_result?: EnsembleMemberStat & { fallback?: boolean }
    total_cost?: number
  }
}

export interface SaveEnsembleMemberRequest {
  role: EnsembleMemberRole
  model: string
  platform: string
  priority: number
  enabled: boolean
}

export const ensembleAPI = {
  listMembers: async (groupId: number): Promise<EnsembleProposer[]> => {
    const { data } = await apiClient.get<EnsembleProposer[]>(
      `/admin/groups/${groupId}/ensemble-proposers`
    )
    return data
  },

  createMember: async (
    groupId: number,
    request: SaveEnsembleMemberRequest
  ): Promise<EnsembleProposer> => {
    const { data } = await apiClient.post<EnsembleProposer>(
      `/admin/groups/${groupId}/ensemble-proposers`,
      request
    )
    return data
  },

  updateMember: async (
    groupId: number,
    memberId: number,
    request: SaveEnsembleMemberRequest
  ): Promise<EnsembleProposer> => {
    const { data } = await apiClient.put<EnsembleProposer>(
      `/admin/groups/${groupId}/ensemble-proposers/${memberId}`,
      request
    )
    return data
  },

  deleteMember: async (groupId: number, memberId: number): Promise<void> => {
    await apiClient.delete(`/admin/groups/${groupId}/ensemble-proposers/${memberId}`)
  },

  getConfig: async (groupId: number): Promise<EnsembleConfig> => {
    const { data } = await apiClient.get<EnsembleConfig>(
      `/admin/groups/${groupId}/ensemble-config`
    )
    return data
  },

  updateConfig: async (groupId: number, config: EnsembleConfig): Promise<EnsembleConfig> => {
    const { data } = await apiClient.put<EnsembleConfig>(
      `/admin/groups/${groupId}/ensemble-config`,
      config
    )
    return data
  },

  /**
   * Send a real gateway request through the selected Ensemble group's API key.
   * This is deliberately a one-off test credential, not a Router configuration.
   */
  test: async (
    apiKey: string,
    messages: Array<{ role: string; content: string }>
  ): Promise<EnsembleTestResponse> => {
    const response = await fetch(buildGatewayUrl('/v1/chat/completions'), {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${apiKey}`
      },
      body: JSON.stringify({
        model: 'ensemble',
        messages,
        stream: false
      })
    })
    const data = await response.json().catch(() => ({}))
    if (!response.ok) {
      throw new Error(data?.error?.message ?? data?.message ?? response.statusText)
    }
    return data as EnsembleTestResponse
  }
}

export default ensembleAPI
