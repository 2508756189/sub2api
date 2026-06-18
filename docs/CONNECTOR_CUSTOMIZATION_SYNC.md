# Connector Customization Sync

This fork keeps TokenPort-specific connector behavior as small, reusable frontend modules instead of scattering custom logic through upstream files.

## What Is Custom

- `frontend/src/components/keys/ConnectorOptions.vue` renders selectable Claude/Codex connector parameters.
- `frontend/src/components/keys/SkillMarketSelector.vue` renders the Skill Market picker.
- `frontend/src/components/keys/connectorTemplates.ts` generates Claude Code, Codex, and skill install scripts.
- `frontend/src/constants/connectorPresets.ts` owns Claude plugin/model and Codex model/MCP presets.
- `frontend/src/api/skillMarket.ts` loads remote registries and falls back to the bundled `/skill-market/index.json`.
- `frontend/src/components/keys/UseKeyModal.vue` is the only upstream component with mount-point changes.

## Why It Survives Upstream Updates

Most TokenPort logic lives in new files. Upstream updates should only conflict where `UseKeyModal.vue`, `Dockerfile`, or `.dockerignore` changed. Resolve those by re-adding the small mount points and keeping the new modules unchanged.

The runtime image must still be rebuilt. Sub2API embeds the frontend into the Go binary, so the stock upstream image cannot load these UI components dynamically.

## Skill Market Refresh

After updating `state-of-art-skills`, rebuild the market there, then refresh the bundled fallback:

```powershell
Set-Location D:\TokenPort智能应用与技能接入平台\state-of-art-skills
python scripts\test_build_market.py
python scripts\build_market.py

Set-Location D:\TokenPort智能应用与技能接入平台\sub2api
.\scripts\sync-skill-market.ps1
```

The preferred live registry is jsDelivr:

```text
https://cdn.jsdelivr.net/gh/2508756189/state-of-art-skills@main/market/index.json
```

If CDN/raw GitHub is unavailable, the UI falls back to the bundled same-origin registry:

```text
/skill-market/index.json
```

## Upstream Sync Workflow

```powershell
Set-Location D:\TokenPort智能应用与技能接入平台\sub2api
git status --short
git fetch origin
git rebase origin/main
.\scripts\sync-skill-market.ps1
docker build -t sub2api-fork:dev .

Set-Location D:\TokenPort智能应用与技能接入平台\token-platform
docker compose -f docker-compose.local.yml up -d --force-recreate sub2api
curl.exe -s http://127.0.0.1:8080/health
curl.exe -s http://127.0.0.1:8080/skill-market/index.json
```

Do not run `docker compose down -v` for normal source updates. That can remove the database volume or local data directories depending on compose configuration.
