# TokenPort Runtime Repository

本仓库同时保留官方上游镜像和 TokenPort 产品代码，两者通过分支隔离。

## 分支职责

| 分支 | 用途 | 是否直接开发 |
| --- | --- | --- |
| `main` | 镜像 `Wei-Shaw/sub2api` | 否 |
| `tokenport/main` | TokenPort 产品、连接器、品牌层和 Skill Market | 是 |

`origin` 指向 `2508756189/sub2api`，`upstream` 指向 `Wei-Shaw/sub2api`。

## 产品代码位置

- `frontend/src/tokenport/brand/`: TokenPort 品牌常量。
- `frontend/src/tokenport/home/`: 商业首页。
- `frontend/src/tokenport/market/`: Skill Market 页面与详情。
- `frontend/src/tokenport/access-center/`: 统一接入配置中心。
- `tokenport-overlay.json`: 上游挂载点清单。
- `scripts/check-tokenport-overlay.ps1`: 挂载点完整性检查。

## 上游同步

官方更新先同步到 `main`，再通过 `.github/workflows/tokenport-upstream-sync.yml` 创建合并到 `tokenport/main` 的审查分支。不要在产品分支执行 `git pull origin main`。

```powershell
git switch tokenport/main
git pull --ff-only
powershell -ExecutionPolicy Bypass -File scripts\check-tokenport-overlay.ps1
```

运行部署与项目资料位于 `2508756189/token-platform`，Skill 市场源位于 `2508756189/state-of-art-skills`。本仓库保留 Sub2API 原项目的许可证、合规声明和技术身份。
