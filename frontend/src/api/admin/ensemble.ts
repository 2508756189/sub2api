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
  stream_trace?: boolean
  source_group_ids?: number[]
}

export interface EnsembleMemberStat {
  model: string
  platform?: string
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

export interface EnsembleProgressEvent {
  type: 'started' | 'member_started' | 'member_finished' | 'proposers_finished' | 'fallback' | 'error' | 'completed' | string
  model?: string
  platform?: string
  role?: EnsembleMemberRole | string
  status?: string
  error?: string
  member?: EnsembleMemberStat
  proposers_total?: number
  proposers_succeeded?: number
  aggregator?: string
  aggregated?: boolean
  duration_ms?: number
  status_code?: number
  response?: EnsembleTestResponse
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
  },

  /**
   * Run the same gateway request as test(), but receive execution diagnostics
   * as server-sent events so the admin page can show where a slow call stops.
   */
  testStream: async (
    apiKey: string,
    messages: Array<{ role: string; content: string }>,
    onEvent: (event: EnsembleProgressEvent) => void,
    signal?: AbortSignal
  ): Promise<EnsembleTestResponse | null> => {
    const response = await fetch(buildGatewayUrl('/v1/ensemble/test'), {
      method: 'POST',
      headers: {
        Accept: 'text/event-stream',
        'Content-Type': 'application/json',
        Authorization: `Bearer ${apiKey}`
      },
      body: JSON.stringify({
        model: 'ensemble',
        messages,
        stream: false
      }),
      signal
    })

    if (!response.ok) {
      const data = await response.json().catch(() => ({}))
      throw new Error(data?.error?.message ?? data?.message ?? response.statusText)
    }
    if (!response.body) throw new Error('测试接口没有返回进度流')

    const reader = response.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''
    let eventName = ''
    let dataLines: string[] = []
    let finalResponse: EnsembleTestResponse | null = null

    const consumeEvent = () => {
      const currentEvent = eventName || 'message'
      const payload = dataLines.join('\n')
      if (payload) {
        try {
          const parsed = JSON.parse(payload) as Record<string, unknown>
          const event = { ...parsed, type: currentEvent } as EnsembleProgressEvent
          onEvent(event)
          if ((currentEvent === 'completed' || currentEvent === 'error') && event.response) {
            finalResponse = event.response
          }
        } catch {
          onEvent({ type: currentEvent, error: payload })
        }
      }
      eventName = ''
      dataLines = []
    }

    while (true) {
      const { done, value } = await reader.read()
      buffer += decoder.decode(value ?? new Uint8Array(), { stream: !done })
      const lines = buffer.split(/\r?\n/)
      buffer = lines.pop() ?? ''
      for (const line of lines) {
        if (line === '') {
          consumeEvent()
        } else if (line.startsWith('event:')) {
          eventName = line.slice(6).trim()
        } else if (line.startsWith('data:')) {
          dataLines.push(line.slice(5).trimStart())
        }
      }
      if (done) break
    }
    if (buffer || dataLines.length) consumeEvent()
    return finalResponse
  }
}

export default ensembleAPI
