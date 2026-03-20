# Gemini CLI Mandates for CC-Connect (Slim Version)

This repository is a deeply customized and slimmed-down version of the original `cc-connect`. You MUST strictly follow these mandates when performing any research or code modification tasks.

## 1. Core Architectural Mandates

- **Platform Scope**: ONLY `telegram` is supported. 
  - NEVER attempt to re-add or reference `feishu`, `discord`, `slack`, `dingtalk`, `wecom`, `qq`, or `line`.
- **Agent Scope**: ONLY `gemini` and `claudecode` (Claude Code) are supported.
  - NEVER attempt to re-add or reference `cursor`, `codex`, `qoder`, `opencode`, or `iflow`.
- **Module Path**: The module path is `github.com/AtticusZeller/cc-connect`.
  - ALWAYS use this path for internal imports. NEVER use `github.com/chenhg5/cc-connect`.

## 2. Maintenance & Synchronization Mandates

- **Upstream Sync**: Follow the guidelines in `docs/maintenance/upstream-sync.md`.
  - Use `git cherry-pick` for specific fixes/features from `upstream/main`.
  - ALWAYS run `go mod tidy` and fix import paths immediately after a cherry-pick.
- **Dependency Cleanliness**: Keep `go.mod` lean. Do not add dependencies for unused platforms or agents.

## 3. Deployment & Release Mandates

- **NPM Package**: The package name is `@atticux/cc-connect`.
  - ALWAYS verify `npm/package.json` before tagging a release.
- **Tagging**: Releases are triggered by `v*` tags (e.g., `v1.2.2-beta.4`).
- **Secrets**: Use `NPM_TOKEN` (configured with 2FA bypass) and `GITHUB_TOKEN` in CI.

## 4. Documentation & Communication

- **Tone**: Professional, direct, and concise.
- **Technical Integrity**: Prioritize the stability of the Telegram/Gemini/Claude bridge. 
- **Inquiry vs. Directive**: If a user reports a bug without an explicit fix request, perform an Inquiry first. Do not modify files until a Directive is issued.

## 5. Development Workspace

- **Build**: Use `go build ./...` to verify changes.
- **Test**: Use `go test ./...` to ensure no regressions in core engine or Telegram/Gemini/Claude components.
