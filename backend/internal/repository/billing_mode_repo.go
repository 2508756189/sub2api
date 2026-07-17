package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

type billingModeRepository struct{ db *sql.DB }

func NewBillingModeRepository(db *sql.DB) service.BillingModeRepository {
	return &billingModeRepository{db: db}
}

func (r *billingModeRepository) ReadBillingMode(ctx context.Context) (service.BillingModeSnapshot, error) {
	if r == nil || r.db == nil {
		return service.BillingModeSnapshot{}, nil
	}
	rows, err := r.db.QueryContext(ctx, `SELECT key, value FROM settings WHERE key = ANY($1)`, pq.Array([]string{
		service.SettingKeyBillingCurrency, service.SettingKeyBillingUSDToCNYRate, "tokenport_billing_currency_migration_v1",
	}))
	if err != nil {
		return service.BillingModeSnapshot{}, err
	}
	defer rows.Close()
	values := map[string]string{}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return service.BillingModeSnapshot{}, err
		}
		values[key] = value
	}
	if err := rows.Err(); err != nil {
		return service.BillingModeSnapshot{}, err
	}
	currency := config.NormalizeBillingCurrency(values[service.SettingKeyBillingCurrency])
	rate, _ := strconv.ParseFloat(values[service.SettingKeyBillingUSDToCNYRate], 64)
	if values[service.SettingKeyBillingCurrency] == "" {
		var marker struct {
			Currency string  `json:"currency"`
			Rate     float64 `json:"rate"`
		}
		if json.Unmarshal([]byte(values["tokenport_billing_currency_migration_v1"]), &marker) == nil && marker.Currency != "" {
			currency = config.NormalizeBillingCurrency(marker.Currency)
			if marker.Rate > 0 {
				rate = marker.Rate
			}
		}
	}
	return service.BillingModeSnapshot{Currency: currency, USDToCNYRate: rate, Configured: values[service.SettingKeyBillingCurrency] != "" || values["tokenport_billing_currency_migration_v1"] != ""}, nil
}

func (r *billingModeRepository) SaveBillingMode(ctx context.Context, currency string, rate float64) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO settings (key, value, updated_at) VALUES
		($1, $2, NOW()), ($3, $4, NOW())
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at`,
		service.SettingKeyBillingCurrency, currency, service.SettingKeyBillingUSDToCNYRate, strconv.FormatFloat(rate, 'f', 8, 64))
	return err
}

func (r *billingModeRepository) ConvertBillingMode(ctx context.Context, conversion service.BillingModeConversion) (service.BillingModeConversionResult, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return service.BillingModeConversionResult{}, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(2508756189)"); err != nil {
		return service.BillingModeConversionResult{}, err
	}
	for _, table := range []string{"settings", "users", "api_keys", "groups", "user_platform_quotas", "user_subscriptions", "usage_logs", "usage_dashboard_daily", "usage_dashboard_hourly", "billing_usage_entries", "promo_codes", "promo_code_usages", "redeem_codes", "user_affiliates", "user_affiliate_ledger", "batch_image_jobs", "batch_image_items"} {
		if _, err := tx.ExecContext(ctx, fmt.Sprintf("LOCK TABLE %s IN ACCESS EXCLUSIVE MODE", table)); err != nil {
			return service.BillingModeConversionResult{}, err
		}
	}
	var active int64
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM batch_image_jobs WHERE status NOT IN ('completed','failed','cancelled','output_deleted')`).Scan(&active); err != nil {
		return service.BillingModeConversionResult{}, err
	}
	if active > 0 {
		return service.BillingModeConversionResult{}, service.ErrBillingModeActiveJobs
	}

	factor := conversion.Factor
	updates := []struct{ table, sql string }{
		{"users", `UPDATE users SET balance = balance * $1, frozen_balance = frozen_balance * $1, total_recharged = total_recharged * $1, balance_notify_threshold = balance_notify_threshold * $1`},
		{"api_keys", `UPDATE api_keys SET quota = quota * $1, quota_used = quota_used * $1, rate_limit_5h = rate_limit_5h * $1, rate_limit_1d = rate_limit_1d * $1, rate_limit_7d = rate_limit_7d * $1, usage_5h = usage_5h * $1, usage_1d = usage_1d * $1, usage_7d = usage_7d * $1`},
		{"groups", `UPDATE groups SET daily_limit_usd = daily_limit_usd * $1, weekly_limit_usd = weekly_limit_usd * $1, monthly_limit_usd = monthly_limit_usd * $1`},
		{"user_platform_quotas", `UPDATE user_platform_quotas SET daily_limit_usd = daily_limit_usd * $1, weekly_limit_usd = weekly_limit_usd * $1, monthly_limit_usd = monthly_limit_usd * $1, daily_usage_usd = daily_usage_usd * $1, weekly_usage_usd = weekly_usage_usd * $1, monthly_usage_usd = monthly_usage_usd * $1`},
		{"user_subscriptions", `UPDATE user_subscriptions SET daily_usage_usd = daily_usage_usd * $1, weekly_usage_usd = weekly_usage_usd * $1, monthly_usage_usd = monthly_usage_usd * $1`},
		{"usage_logs", `UPDATE usage_logs SET input_cost = input_cost * $1, output_cost = output_cost * $1, cache_creation_cost = cache_creation_cost * $1, cache_read_cost = cache_read_cost * $1, total_cost = total_cost * $1, actual_cost = actual_cost * $1, image_output_cost = image_output_cost * $1, image_input_cost = image_input_cost * $1, account_stats_cost = account_stats_cost * $1`},
		{"usage_dashboard_daily", `UPDATE usage_dashboard_daily SET total_cost = total_cost * $1, actual_cost = actual_cost * $1, account_cost = account_cost * $1`},
		{"usage_dashboard_hourly", `UPDATE usage_dashboard_hourly SET total_cost = total_cost * $1, actual_cost = actual_cost * $1, account_cost = account_cost * $1`},
		{"billing_usage_entries", `UPDATE billing_usage_entries SET delta_usd = delta_usd * $1`},
		{"promo_codes", `UPDATE promo_codes SET bonus_amount = bonus_amount * $1`},
		{"promo_code_usages", `UPDATE promo_code_usages SET bonus_amount = bonus_amount * $1`},
		{"redeem_codes", `UPDATE redeem_codes SET value = value * $1 WHERE type = 'admin_balance'`},
		{"user_affiliates", `UPDATE user_affiliates SET aff_quota = aff_quota * $1, aff_history_quota = aff_history_quota * $1, aff_frozen_quota = aff_frozen_quota * $1`},
		{"user_affiliate_ledger", `UPDATE user_affiliate_ledger SET amount = amount * $1, balance_after = balance_after * $1, aff_quota_after = aff_quota_after * $1, aff_frozen_quota_after = aff_frozen_quota_after * $1, aff_history_quota_after = aff_history_quota_after * $1`},
		{"batch_image_jobs", `UPDATE batch_image_jobs SET estimated_cost = estimated_cost * $1, hold_amount = hold_amount * $1, actual_cost = actual_cost * $1, base_unit_price = base_unit_price * $1, billable_unit_price = billable_unit_price * $1, hold_unit_price = hold_unit_price * $1, currency = $2`},
		{"batch_image_items", `UPDATE batch_image_items SET billed_amount = billed_amount * $1`},
	}
	var converted int64
	tables := make([]string, 0, len(updates))
	for _, update := range updates {
		var result sql.Result
		if update.table == "batch_image_jobs" {
			result, err = tx.ExecContext(ctx, update.sql, factor, conversion.ToCurrency)
		} else {
			result, err = tx.ExecContext(ctx, update.sql, factor)
		}
		if err != nil {
			return service.BillingModeConversionResult{}, err
		}
		count, _ := result.RowsAffected()
		converted += count
		tables = append(tables, update.table)
	}
	if err := convertMoneySettings(ctx, tx, factor); err != nil {
		return service.BillingModeConversionResult{}, err
	}
	historyEntries, err := readBillingHistory(ctx, tx)
	if err != nil {
		return service.BillingModeConversionResult{}, err
	}
	historyEntries = append(historyEntries, normalizeBillingHistoryEntry(conversion.HistoryEntry))
	if len(historyEntries) > 20 {
		historyEntries = historyEntries[len(historyEntries)-20:]
	}
	history, _ := json.Marshal(historyEntries)
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO settings (key, value, updated_at) VALUES
		($1, $2, NOW()), ($3, $4, NOW()), ($5, $6, NOW())
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at`,
		service.SettingKeyBillingCurrency, conversion.ToCurrency,
		service.SettingKeyBillingUSDToCNYRate, strconv.FormatFloat(conversion.ToRate, 'f', 8, 64),
		service.SettingKeyBillingCurrencyHistory, string(history)); err != nil {
		return service.BillingModeConversionResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return service.BillingModeConversionResult{}, err
	}
	return service.BillingModeConversionResult{RowsConverted: converted, Tables: tables}, nil
}

func readBillingHistory(ctx context.Context, tx *sql.Tx) ([]map[string]any, error) {
	var raw string
	err := tx.QueryRowContext(ctx, `SELECT value FROM settings WHERE key = $1`, service.SettingKeyBillingCurrencyHistory).Scan(&raw)
	if err == sql.ErrNoRows {
		return []map[string]any{}, nil
	}
	if err != nil {
		return nil, err
	}
	var entries []map[string]any
	if err := json.Unmarshal([]byte(raw), &entries); err != nil {
		return nil, fmt.Errorf("invalid billing currency history: %w", err)
	}
	return entries, nil
}

func normalizeBillingHistoryEntry(entry map[string]any) map[string]any {
	return map[string]any{"at": time.Now().UTC().Format(time.RFC3339), "from": entry["from"], "to": entry["to"], "from_rate": entry["from_rate"], "to_rate": entry["to_rate"], "factor": entry["factor"]}
}

func convertMoneySettings(ctx context.Context, tx *sql.Tx, factor float64) error {
	keys := []string{
		"default_balance", "balance_low_notify_threshold",
		"auth_source_default_email_balance", "auth_source_default_linuxdo_balance", "auth_source_default_oidc_balance",
		"auth_source_default_wechat_balance", "auth_source_default_github_balance", "auth_source_default_google_balance", "auth_source_default_dingtalk_balance",
		"default_platform_quotas", "auth_source_default_email_platform_quotas", "auth_source_default_linuxdo_platform_quotas", "auth_source_default_oidc_platform_quotas",
		"auth_source_default_wechat_platform_quotas", "auth_source_default_github_platform_quotas", "auth_source_default_google_platform_quotas", "auth_source_default_dingtalk_platform_quotas",
	}
	rows, err := tx.QueryContext(ctx, `SELECT key, value FROM settings WHERE key = ANY($1) FOR UPDATE`, pq.Array(keys))
	if err != nil {
		return err
	}
	defer rows.Close()
	values := map[string]string{}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return err
		}
		values[key] = value
	}
	for key, value := range values {
		var converted string
		if strings.Contains(key, "platform_quotas") {
			var raw map[string]map[string]*float64
			if err := json.Unmarshal([]byte(value), &raw); err != nil {
				return fmt.Errorf("invalid monetary setting %s: %w", key, err)
			}
			for _, dimensions := range raw {
				for name, amount := range dimensions {
					if amount != nil {
						v := *amount * factor
						dimensions[name] = &v
					}
				}
			}
			encoded, err := json.Marshal(raw)
			if err != nil {
				return err
			}
			converted = string(encoded)
		} else {
			amount, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
			if err != nil {
				return fmt.Errorf("invalid monetary setting %s: %w", key, err)
			}
			converted = strconv.FormatFloat(amount*factor, 'f', 8, 64)
		}
		if _, err := tx.ExecContext(ctx, `UPDATE settings SET value = $1, updated_at = NOW() WHERE key = $2`, converted, key); err != nil {
			return err
		}
	}
	return rows.Err()
}
