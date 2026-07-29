# TokenPort Product Overlay

This directory contains user-facing TokenPort product modules that sit on top of the Sub2API runtime.

- `brand/`: product name, copy, supported clients, and upstream attribution.
- `home/`: public commercial home experience.
- `market/`: reusable Skill Market catalog and detail UI.
- `access-center/`: unified client configuration and CCS import workspace.

Upstream files should contain only small mount points. Before merging an upstream update, run:

```powershell
pwsh scripts/check-tokenport-overlay.ps1
cd frontend
npm run typecheck
npm run test:run
npm run build
```

Internal Go modules, routes, database names, environment variables, and protocol identifiers retain their Sub2API names. User-facing defaults resolve to TokenPort, while deployment settings can override the brand.
