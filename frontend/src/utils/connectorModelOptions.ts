import {
  BILLING_MODE_IMAGE,
  BILLING_MODE_PER_REQUEST,
  BILLING_MODE_TOKEN,
} from '@/constants/channel'
import type { UserAvailableChannel, UserSupportedModel, UserSupportedModelPricing } from '@/api/channels'
import { formatScaled } from '@/utils/pricing'

export interface ConnectorModelOption {
  name: string
  platform: string
  pricing: UserSupportedModelPricing | null
  pricingSummary: string
  label: string
  compactLabel: string
}

function summarizePricing(pricing: UserSupportedModelPricing | null): string {
  if (!pricing) return '暂无定价'

  switch (pricing.billing_mode) {
    case BILLING_MODE_PER_REQUEST:
      return `${formatScaled(pricing.per_request_price, 1)} / 次`
    case BILLING_MODE_IMAGE:
      return `${formatScaled(pricing.image_output_price, 1)} / 张`
    case BILLING_MODE_TOKEN:
    default: {
      const input = formatScaled(pricing.input_price, 1_000_000)
      const output = formatScaled(pricing.output_price, 1_000_000)
      return `输入 ${input} / 输出 ${output} / 1M tokens`
    }
  }
}

function summarizePricingCompact(pricing: UserSupportedModelPricing | null): string {
  if (!pricing) return 'no price'

  switch (pricing.billing_mode) {
    case BILLING_MODE_PER_REQUEST:
      return `${formatScaled(pricing.per_request_price, 1)} / req`
    case BILLING_MODE_IMAGE:
      return `${formatScaled(pricing.image_output_price, 1)} / image`
    case BILLING_MODE_TOKEN:
    default: {
      const input = formatScaled(pricing.input_price, 1_000_000)
      const output = formatScaled(pricing.output_price, 1_000_000)
      return `${input}/${output} per 1M`
    }
  }
}

export function toConnectorModelOption(model: UserSupportedModel): ConnectorModelOption {
  const pricingSummary = summarizePricing(model.pricing)
  const compactPricingSummary = summarizePricingCompact(model.pricing)
  return {
    name: model.name,
    platform: model.platform,
    pricing: model.pricing,
    pricingSummary,
    label: `${model.name} · ${pricingSummary}`,
    compactLabel: `${model.name} · ${compactPricingSummary}`,
  }
}

export function extractConnectorModelOptions(params: {
  channels: UserAvailableChannel[]
  platform?: string | null
  groupId?: number | null
}): ConnectorModelOption[] {
  const { channels, platform, groupId } = params
  if (!platform) return []

  const byName = new Map<string, ConnectorModelOption>()
  channels.forEach((channel) => {
    channel.platforms
      .filter((section) => section.platform === platform)
      .filter((section) => {
        if (!groupId) return true
        return section.groups.some((group) => group.id === groupId)
      })
      .forEach((section) => {
        section.supported_models.forEach((model) => {
          if (!model.name || byName.has(model.name)) return
          byName.set(model.name, toConnectorModelOption(model))
        })
      })
  })

  return Array.from(byName.values()).sort((a, b) => a.name.localeCompare(b.name))
}
