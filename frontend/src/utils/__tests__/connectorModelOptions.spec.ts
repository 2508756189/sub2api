import { describe, expect, it } from 'vitest'
import { extractConnectorModelOptions } from '../connectorModelOptions'
import type { UserAvailableChannel } from '@/api/channels'

describe('connectorModelOptions', () => {
  it('extracts models with pricing for the selected platform and group', () => {
    const channels: UserAvailableChannel[] = [
      {
        name: 'primary',
        description: '',
        platforms: [
          {
            platform: 'openai',
            groups: [
              {
                id: 1,
                name: 'soft-dev',
                platform: 'openai',
                subscription_type: 'standard',
                rate_multiplier: 1,
                peak_rate_enabled: false,
                peak_start: '',
                peak_end: '',
                peak_rate_multiplier: 1,
                is_exclusive: false,
              },
            ],
            supported_models: [
              {
                name: 'gpt-test',
                platform: 'openai',
                pricing: {
                  billing_mode: 'token',
                  input_price: 0.000001,
                  output_price: 0.000004,
                  cache_write_price: null,
                  cache_read_price: null,
                  image_output_price: null,
                  per_request_price: null,
                  intervals: [],
                },
              },
            ],
          },
          {
            platform: 'anthropic',
            groups: [],
            supported_models: [
              { name: 'claude-test', platform: 'anthropic', pricing: null },
            ],
          },
        ],
      },
    ]

    const options = extractConnectorModelOptions({
      channels,
      platform: 'openai',
      groupId: 1,
    })

    expect(options).toHaveLength(1)
    expect(options[0].name).toBe('gpt-test')
    expect(options[0].pricingSummary).toContain('输入 $1')
    expect(options[0].pricingSummary).toContain('输出 $4')
  })
})
