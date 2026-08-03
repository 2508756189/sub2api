-- Grok Composer is billed on the same rate card as Grok Build 0.1.
-- Keep this as a separate migration because migration 192 may already be applied.
INSERT INTO channel_model_pricing (
    channel_id, platform, models, billing_mode,
    input_price, output_price, cache_read_price
)
SELECT
    8,
    'grok',
    '["grok-composer-2.5-fast"]'::jsonb,
    'token',
    0.000001,
    0.000002,
    0.0000002
WHERE NOT EXISTS (
    SELECT 1
    FROM channel_model_pricing
    WHERE channel_id = 8
      AND platform = 'grok'
      AND models = '["grok-composer-2.5-fast"]'::jsonb
);
