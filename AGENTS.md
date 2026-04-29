# AGENTS.md

This file records repo-specific operating instructions for Codex working in
`/Users/walker2/workspace/sub2api`.

## Deploy Server

When the user says `部署服务器`, `部署到线上`, `发布到服务器`, or equivalent:

1. Use the operator-provided deployment script outside this repository instead
   of re-deriving the whole process.
2. Do not copy server hosts, usernames, passwords, keys, or deployment scripts
   into this repository.
3. Before running the deploy script, make sure the repo still builds locally.
4. After deployment, verify:
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

The external deploy script is the preferred path, but if deployment fails,
inspect the remote health-check step first. Deploy validation must confirm the
actual live port before deciding whether to roll back.
