# DeepSeek Channel Pricing

This project uses channel model pricing as per-token USD prices. The protocol
format of an upstream account, such as OpenAI-compatible or Anthropic-compatible,
does not decide the model cost. DeepSeek upstream accounts should be priced by
their real DeepSeek model names.

The local TokenPort demo currently uses:

| Model aliases | Input USD/token | Output USD/token | Cache write USD/token | Cache read USD/token |
| --- | ---: | ---: | ---: | ---: |
| `deepseek_v4_pro`, `deepseek-v4-pro` | `0.000000435000` | `0.000000870000` | `0.000000435000` | `0.000000003625` |
| `deepseek_v4_flash`, `deepseek-v4-flash` | `0.000000140000` | `0.000000280000` | `0.000000140000` | `0.000000002800` |

Source: DeepSeek Models & Pricing lists prices per 1M tokens:
https://api-docs.deepseek.com/quick_start/pricing

Conversion rule:

```text
USD/token = USD per 1M tokens / 1,000,000
```

`cache_read_price` maps to DeepSeek cache-hit input pricing. `input_price`
and `cache_write_price` map to cache-miss input pricing.

Apply or re-apply the pricing to the local Docker database:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\apply-deepseek-channel-pricing.ps1
docker compose -f E:\token-platform\docker-compose.local.yml restart sub2api
```

Verify new requests are billable:

```sql
SELECT id, model, input_tokens, output_tokens, input_cost, output_cost, total_cost, actual_cost, channel_id, group_id
FROM usage_logs
ORDER BY id DESC
LIMIT 5;
```

Existing zero-cost rows are intentionally not backfilled by the script. Backfill
and wallet compensation should be a separate explicit operation.
