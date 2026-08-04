-- Store the concrete upstream platform for model identifiers that do not
-- encode their provider in the public name (for example qwen, kimi, glm).
ALTER TABLE ensemble_proposers
    ADD COLUMN IF NOT EXISTS platform VARCHAR(32) NOT NULL DEFAULT '';
