# RustDesk API (Fork) Context & Instructions

## Repository Overview

- Repo: `quanla93/rustdesk-api` (branch `master`)
- Upstream: `lejianwen/rustdesk-api`
- Companion server repo: `quanla93/rustdesk-server` (branch `forapi`)

## Features & Customizations in this Fork

1. **Default English**:
   - `conf/config.yaml`: `lang: "en"`
   - `conf/admin/hello.html`: English greeting by default
   - Source code comments, docs, and Swagger schemas translated to English.
   - All `resources/i18n/*.toml` files cleaned to remove hardcoded Chinese defaults.
2. **Web Client**:
   - Served at `/webclient/` or `/web/`.
   - Manifest and initial shell updated with English branding and modernized loading animation.
3. **Automated CI/CD**:
   - `.github/workflows/build.yml` triggers automatically on pushes to `master` and version tags (`v*.*.*`).
   - Publishes multi-arch container images to `ghcr.io/quanla93/rustdesk-api:latest`.

## Local Testing

- Build/test locally using Docker:
  `docker run --rm -v $(pwd):/work -w /work golang:1.23 sh -lc '/usr/local/go/bin/go test ./... -skip Redis && /usr/local/go/bin/go build ./cmd/apimain.go'`