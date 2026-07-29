param(
    [string]$Container = "sub2api-postgres",
    [string]$Database = "sub2api",
    [string]$User = "sub2api",
    [string]$OpenAIChannelName = "OpenAI API Pool",
    [string]$AnthropicChannelName = "Anthropic API Pool"
)

$ErrorActionPreference = "Stop"

$sql = @"
BEGIN;

WITH target_channels AS (
    SELECT id, 'openai'::varchar AS platform
    FROM channels
    WHERE name = '$($OpenAIChannelName.Replace("'", "''"))'
    UNION ALL
    SELECT id, 'anthropic'::varchar AS platform
    FROM channels
    WHERE name = '$($AnthropicChannelName.Replace("'", "''"))'
),
deleted AS (
    DELETE FROM channel_model_pricing cmp
    USING target_channels tc
    WHERE cmp.channel_id = tc.id
      AND cmp.platform = tc.platform
      AND (
          cmp.models ? 'deepseek_v4_pro'
          OR cmp.models ? 'deepseek-v4-pro'
          OR cmp.models ? 'deepseek_v4_flash'
          OR cmp.models ? 'deepseek-v4-flash'
      )
)
INSERT INTO channel_model_pricing
    (channel_id, platform, models, billing_mode, input_price, output_price,
     cache_write_price, cache_read_price, image_output_price, per_request_price)
SELECT
    tc.id,
    tc.platform,
    v.models::jsonb,
    'token',
    v.input_price,
    v.output_price,
    v.cache_write_price,
    v.cache_read_price,
    NULL,
    NULL
FROM target_channels tc
CROSS JOIN (
    VALUES
        ('["deepseek_v4_pro", "deepseek-v4-pro"]',     0.000000435000::numeric, 0.000000870000::numeric, 0.000000435000::numeric, 0.000000003625::numeric),
        ('["deepseek_v4_flash", "deepseek-v4-flash"]', 0.000000140000::numeric, 0.000000280000::numeric, 0.000000140000::numeric, 0.000000002800::numeric)
) AS v(models, input_price, output_price, cache_write_price, cache_read_price);

COMMIT;
"@

$sql | docker exec -i $Container psql -U $User -d $Database

$verifySql = @"
SELECT c.name AS channel, p.platform, p.models, p.input_price, p.output_price, p.cache_write_price, p.cache_read_price
FROM channel_model_pricing p
JOIN channels c ON c.id = p.channel_id
WHERE p.models ? 'deepseek-v4-pro' OR p.models ? 'deepseek-v4-flash'
ORDER BY c.id, p.id;
"@

docker exec $Container psql -U $User -d $Database -c $verifySql
