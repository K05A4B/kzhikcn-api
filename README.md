# kzhikcn-api

kzhikcn-api is a lightweight headless CMS backend written in Go, built around Markdown content, predictable admin workflows, and simple deployment.

## Highlights

- Article lifecycle: draft/hidden/published, soft delete & restore
- Markdown content with file-based assets
- Categories & tags with expression filters
- JWT auth + optional TOTP MFA
- Rate limiting with high‑quota bypass keys
- RSS & sitemap endpoints
- Local storage and cache (Redis optional)
- Zero‑dependency startup (local SQLite + local cache by default)

## Requirements

- Go 1.25+ (toolchain 1.26.2)
- SQLite3 (default) or MySQL
- Optional: Redis (cache)

## Quick start (local)

1. Generate config: `go run . gen-config`
2. Start server: `go run . serve -a 0.0.0.0:5083`
3. First boot auto‑migrates the DB and creates a default admin (`admin` / `admin`). Change it immediately.

You can also run directly without `gen-config`. If `config.yml` is missing, it will be created from the default template without interactive prompts (no website/JWT questions).

Use `-c` to specify a config file: `go run . -c ./config.yml serve`.

## Configuration

- Default config file: `config.yml` (auto‑generated if missing).
- Environment placeholders are supported inside config, e.g. `${WEBSITE_DESCRIPTION}` reads from env (see `docs/configuration.md`).
- To enable HTTPS, set `cert_file` and `key_file`.

## CLI

See `docs/cli.md` for admin/user management commands.

## API

Base path: `/api/v1`

See `docs/introduct.md` and `docs/api/*` for details.

## Docs

- `docs/introduct.md` (API overview, Chinese)
- `docs/configuration.md`
- `docs/cli.md`
- `docs/deployment.md`

## Development

For hot reload, use Air with `.air.toml` (optional).

## Docker

The Docker image listens on `${ADDRESS}` (example: `0.0.0.0:5083`). Mount config and data directories in production.

## 中文版本

请阅读 `README.zh-CN.md`。
