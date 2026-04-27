# AGENTS.md

This file records repo-specific operating instructions for Codex working in
`/Users/walker2/workspace/sub2api`.

## Deploy Server

When the user says `部署服务器`, `部署到线上`, `发布到服务器`, or equivalent:

1. Treat it as a request to deploy to the current production host:
   - host: `47.251.68.126`
   - service: `sub2api`
   - app dir: `/opt/sub2api`
2. Use the external deployment script instead of re-deriving the whole process:
   - `/Users/walker2/workspace/.codex-tmp/sub2api_deploy_47.251.68.126.sh`
3. Do not copy server credentials into this repository.
4. Before running the deploy script, make sure the repo still builds locally.
5. After deployment, verify:
   - remote health endpoint returns OK
   - `systemctl status sub2api` is active
   - remote binary version matches the intended build

## Build Notes

This repo currently relies on embedded frontend assets for production builds.

Expected local build sequence:

1. Frontend:
   - `pnpm --dir frontend install --frozen-lockfile`
   - `pnpm --dir frontend run build`
2. Backend:
   - `cd backend && go build -tags embed ./cmd/server`

## Wire Regeneration

If backend compilation fails because `cmd/server/wire_gen.go` is out of date or
provider signatures changed:

1. Regenerate Wire output:
   - `cd backend && go generate ./cmd/server`
2. Re-run backend build.

Do not hand-edit `backend/cmd/server/wire_gen.go` unless there is a strong
reason; prefer regeneration.

## Current Deployment Caveat

The external deploy script exists and is the preferred path, but if deployment
fails, inspect the remote health-check step first. The current server is known
to listen on port `18899`, and deploy validation must confirm the actual live
port before deciding whether to roll back.
