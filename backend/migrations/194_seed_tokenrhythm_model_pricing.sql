-- Seed the TokenRhythm model rates exposed by its /v1/models endpoint.
--
-- The endpoint reports effective CNY prices per 1M tokens.  ChannelModelPricing
-- stores USD per token, so these values use the platform's current 7.2 CNY/USD
-- exchange-rate setting.  Do not overwrite an administrator's edited row.
-- qwen3.7-max and kimi-k2.7-code are only seeded for the OpenAI-compatible
-- channel because the upstream model metadata does not advertise Anthropic
-- protocol support for them.
--
-- Channel rows are tenant data rather than schema data.  Skip the seed when a
-- configured channel does not exist, so a fresh installation can complete its
-- migrations and still use the billing fallback rates until the channel exists.
WITH seed(channel_id, platform, model_name, input_price, output_price, cache_read_price) AS (
    VALUES
        (5, 'anthropic', 'glm-5',            0.000000833333333333, 0.000003055555555556, 0.000000208333333333),
        (5, 'anthropic', 'kimi-k2.5',        0.000000555555555556, 0.000002916666666667, 0.000000111111111111),
        (5, 'anthropic', 'minimax-m2.7',     0.000000291666666667, 0.000001166666666667, NULL),
        (5, 'anthropic', 'mimo-v2.5-pro',    0.000000416666666667, 0.000000833333333333, NULL),
        (6, 'openai',    'glm-5',            0.000000833333333333, 0.000003055555555556, 0.000000208333333333),
        (6, 'openai',    'kimi-k2.5',        0.000000555555555556, 0.000002916666666667, 0.000000111111111111),
        (6, 'openai',    'qwen3.7-max',      0.000000833333333333, 0.000002500000000000, 0.000000166666666667),
        (6, 'openai',    'minimax-m2.7',     0.000000291666666667, 0.000001166666666667, NULL),
        (6, 'openai',    'mimo-v2.5-pro',    0.000000416666666667, 0.000000833333333333, NULL),
        (6, 'openai',    'kimi-k2.7-code',   0.000000902777777778, 0.000003750000000000, 0.000000180555555556)
)
INSERT INTO channel_model_pricing (
    channel_id, platform, models, billing_mode,
    input_price, output_price, cache_read_price
)
SELECT
    c.id,
    s.platform,
    jsonb_build_array(s.model_name),
    'token',
    s.input_price,
    s.output_price,
    s.cache_read_price
FROM seed s
JOIN channels AS c ON c.id = s.channel_id
WHERE NOT EXISTS (
    SELECT 1
    FROM channel_model_pricing p
    WHERE p.channel_id = s.channel_id
      AND p.platform = s.platform
      AND p.models = jsonb_build_array(s.model_name)
);
