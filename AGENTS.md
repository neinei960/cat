# Codex Project Notes

## Auto Deploy

This repository already has a Claude deployment hook at:

- `/Users/genglsh/workstation/cat/cat/.claude/hooks/deploy.sh`

When editing this project with Codex, treat remote deployment as required work, not optional cleanup.

Rules:

- After any code change under `web/`, run the deploy hook so the H5 frontend is built and synced to the remote server.
- After any code change under `server/`, run the deploy hook so the Linux binary is rebuilt and restarted on the remote server.
- If both `web/` and `server/` change, deploy both.
- Do not skip deployment unless the user explicitly says not to deploy.
- For any change that affects UI, run a Playwright check after deployment to verify the actual rendered behavior.

Preferred command:

```bash
printf '{"tool_input":{"file_path":"/absolute/path/to/changed/file"}}' | /Users/genglsh/workstation/cat/cat/.codex/hooks/deploy.sh
```

Remote deployment target:

- host: `36.151.144.227`
- user: `root`
- web dir: `/opt/cat/web/`
- server dir: `/opt/cat/server/`
