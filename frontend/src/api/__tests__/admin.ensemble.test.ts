import { describe, expect, it, vi } from 'vitest'

vi.mock('@/api/client', () => ({
  apiClient: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn()
  },
  buildGatewayUrl: (path: string) => path
}))

describe('Ensemble diagnostic stream', () => {
  it('parses SSE event boundaries and returns the final response', async () => {
    const encoder = new TextEncoder()
    const chunks = [
      encoder.encode('event: started\ndata: {"proposers_total":2}\n\n'),
      encoder.encode('event: completed\ndata: {"status_code":200,"response":{"choices":[{"message":{"content":"done"}}]}}\n\n')
    ]
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      body: {
        getReader: () => ({
          read: vi.fn()
            .mockResolvedValueOnce({ done: false, value: chunks[0] })
            .mockResolvedValueOnce({ done: false, value: chunks[1] })
            .mockResolvedValueOnce({ done: true, value: undefined })
        })
      }
    })
    vi.stubGlobal('fetch', fetchMock)

    const { default: ensembleAPI } = await import('@/api/admin/ensemble')
    const events: Array<{ type: string; proposers_total?: number }> = []
    const response = await ensembleAPI.testStream('test-key', [], event => events.push(event))

    expect(fetchMock).toHaveBeenCalledWith('/v1/ensemble/test', expect.objectContaining({ method: 'POST' }))
    expect(events.map(event => event.type)).toEqual(['started', 'completed'])
    expect(events[0].proposers_total).toBe(2)
    expect(response?.choices?.[0]?.message?.content).toBe('done')
  })
})
