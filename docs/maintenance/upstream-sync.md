# 上游项目同步与合入规范 (Upstream Sync Specification)

## 1. 概述 (Overview)
由于本项目已针对特定场景进行了深度精简（仅保留 Telegram 平台及 Gemini/Claude Code 代理），直接进行 `git merge` 会导致大量的冲突（因物理删除文件引起）。本规范旨在指导维护者如何高效、安全地从上游仓库 (`github.com/chenhg5/cc-connect`) 同步核心功能改进与 Bug 修复，同时保持本项目的精简性。

## 2. 环境配置 (Setup)
确保已添加上游仓库源：
```bash
git remote add upstream https://github.com/chenhg5/cc-connect.git
git fetch upstream
```

## 3. 同步策略 (Sync Strategy)

### 3.1 核心原则
- **按需合入 (Selective Pick)**：不建议直接 `merge` 整个分支。优先使用 `git cherry-pick` 合入具体的 Feature 或 Fix。
- **排除无关目录**：忽略所有关于 `feishu`, `dingtalk`, `slack`, `discord`, `wecom`, `qq`, `line` 等平台的提交。
- **自动更正导入路径**：合入后必须立即运行脚本，将 `github.com/chenhg5/cc-connect` 修正为 `github.com/AtticusZeller/cc-connect`。

### 3.2 重点关注范围
合入时应重点关注以下路径的变更：
- `core/`：核心引擎、接口定义、i18n 等。
- `agent/claudecode/` & `agent/gemini/`：代理逻辑优化。
- `platform/telegram/`：Telegram 适配器更新。
- `cmd/cc-connect/`：CLI 命令与守护进程逻辑。
- `config/` & `daemon/`：配置解析与系统服务支持。

---

## 4. 执行流程 (Workflow)

### 第一步：获取更新
```bash
git fetch upstream main
```

### 第二步：识别提交
查看上游近期提交，筛选出与核心引擎或保留组件相关的 Commit ID：
```bash
git log upstream/main --oneline --grep="core" --grep="telegram" --grep="gemini" --grep="claudecode"
```

### 第三步：执行 Cherry-pick
```bash
git cherry-pick <COMMIT_ID>
```

**冲突处理 (Conflict Handling)：**
- **Deleted by us**：如果冲突提示文件在本项目中已被删除（例如 `platform/feishu.go`），请直接使用 `git rm <file>` 保持删除状态。
- **Import Path Conflict**：如果因为导入路径不同（`chenhg5` vs `AtticusZeller`）产生冲突，请接受上游更改，并在下一步统一修正。

### 第四步：修正与清理 (Post-pick Cleanup)
合入后，务必执行以下自动化清理命令：

1. **修正导入路径 (Import Path Fix)**:
   ```bash
   # 批量替换所有 .go 文件中的模块路径
   find . -name "*.go" -type f -exec sed -i 's|github.com/chenhg5/cc-connect|github.com/AtticusZeller/cc-connect|g' {} +
   ```

2. **整理依赖 (Tidy)**:
   ```bash
   go mod tidy
   ```

3. **格式化 (Format)**:
   ```bash
   go run golang.org/x/tools/cmd/goimports@latest -w .
   ```

### 第五步：验证构建
确保精简后的项目依然保持高可用：
```bash
go build ./...
go test ./...
```

---

## 5. 发布流程 (Release Process)

同步并验证成功后，按照以下步骤发布新版本：

1. **更新版本号**: 修改 `npm/package.json` 中的 `version`（例如从 `1.2.2-beta.3` 升至 `1.2.2-beta.4`）。
2. **提交并标记**:
   ```bash
   git add .
   git commit -m "chore: sync with upstream <COMMIT_ID_OR_DATE>"
   git tag v1.2.2-beta.4
   ```
3. **推送触发 CI**:
   ```bash
   git push origin main
   git push origin v1.2.2-beta.4
   ```

## 6. 注意事项 (Notes)
- **拒绝反向合并**：严禁将本仓库的 `main` 分支直接合并回上游。
- **NPM 权限**：确保 GitHub Secrets 中的 `NPM_TOKEN` 依然有效。
