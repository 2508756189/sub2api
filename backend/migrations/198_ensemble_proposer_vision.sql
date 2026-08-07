-- Ensemble 成员视觉能力标记：默认 true 保持既有行为，管理员显式把无视觉
-- 模型置 false 后，图片请求会跳过该成员（能力补齐）而不是让它调用失败。
ALTER TABLE ensemble_proposers
    ADD COLUMN IF NOT EXISTS vision BOOLEAN NOT NULL DEFAULT TRUE;
