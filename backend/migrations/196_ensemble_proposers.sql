-- Ensemble 分组：成员表 + 分组级配置列
-- 仅 platform='ensemble' 的分组使用。成员模型由该分组自己绑定的账号提供
-- （组内聚合，不跨上游）。

CREATE TABLE IF NOT EXISTS ensemble_proposers (
    id BIGSERIAL PRIMARY KEY,
    group_id BIGINT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL DEFAULT 'proposer',
    model VARCHAR(200) NOT NULL,
    priority INTEGER NOT NULL DEFAULT 100,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,
    CONSTRAINT ensemble_proposers_role_check CHECK (role IN ('proposer', 'aggregator'))
);

-- 同一分组内，同一角色下模型名唯一（软删后可重用）。
CREATE UNIQUE INDEX IF NOT EXISTS idx_ensemble_proposers_unique_active
    ON ensemble_proposers (group_id, role, model)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_ensemble_proposers_group_enabled
    ON ensemble_proposers (group_id, enabled)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_ensemble_proposers_group_role
    ON ensemble_proposers (group_id, role)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_ensemble_proposers_group_priority
    ON ensemble_proposers (group_id, priority, id)
    WHERE deleted_at IS NULL;

-- 分组级 Ensemble 配置（聚合开关、最少成功候选数、超时等）。
ALTER TABLE groups ADD COLUMN IF NOT EXISTS ensemble_config JSONB NOT NULL DEFAULT '{}'::jsonb;
