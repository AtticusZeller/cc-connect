# cc-connect Installation & Configuration Guide

> **This document is designed to be read by AI coding agents (Claude Code, Gemini CLI, etc.) to help users install and configure cc-connect.** You can feed this file directly to your AI agent.

## What is cc-connect?

cc-connect bridges local AI coding assistants to Telegram, enabling bidirectional chat with your local AI agent from anywhere.

GitHub: https://github.com/AtticusZeller/cc-connect

**IMPORTANT: If you are an AI agent helping a user install cc-connect, use interactive tools (like AskUserQuestion) to guide them through configuration choices:**
- Agent selection (Claude Code, Gemini)
- API keys and authentication tokens
- Project paths and preferences

Don't guess values—always ask the user to choose via interactive prompts.

## Step 1: Install cc-connect

### Option A: npm (recommended for most users)

```bash
npm install -g @atticux/cc-connect
```

After installation, `cc-connect` binary will be available globally.

### Option B: Download binary from GitHub Releases

Go to https://github.com/AtticusZeller/cc-connect/releases and download binary for your platform:

- `cc-connect-linux-amd64` / `cc-connect-linux-arm64`
- `cc-connect-darwin-amd64` / `cc-connect-darwin-arm64`
- `cc-connect-windows-amd64.exe`

```bash
# Example for Linux amd64:
curl -L -o cc-connect https://github.com/AtticusZeller/cc-connect/releases/latest/download/cc-connect-linux-amd64
chmod +x cc-connect
sudo mv cc-connect /usr/local/bin/
```

On macOS, you may need to remove the quarantine attribute:

```bash
xattr -d com.apple.quarantine cc-connect
```

### Option C: Build from source

Requires Go 1.22+.

```bash
git clone https://github.com/AtticusZeller/cc-connect.git
cd cc-connect
make build
# Binary will be at ./cc-connect
```

## Step 2: Install your AI Agent

cc-connect supports Claude Code and Gemini CLI. Install at least one:

```bash
# Claude Code
npm install -g @anthropic-ai/claude-code

# Gemini CLI
npm install -g @google/gemini-cli
```

Verify your selected agent works:

```bash
claude --version
gemini --version
```

## Step 3: Create config.toml

cc-connect looks for config in this order:
1. `-config <path>` flag (explicit)
2. `./config.toml` (current directory)
3. `~/.cc-connect/config.toml` (global, **recommended**)

If no config file exists, running `cc-connect` will auto-create a starter template at `~/.cc-connect/config.toml`.

**Recommended: use the global config location:**

```bash
mkdir -p ~/.cc-connect
# If you cloned the repo, copy the example:
cp config.example.toml ~/.cc-connect/config.toml
# Or just run cc-connect once — it will create a starter config automatically
```

The configuration has this structure:

```toml
# Optional global settings
# language = "en"  # "en", "zh", or "" (auto-detect)

[log]
level = "info"  # debug, info, warn, error

# Each [[projects]] entry connects one code folder to one or more messaging platforms
[[projects]]
name = "my-project"

[projects.agent]
type = "claudecode"  # or "gemini"

[projects.agent.options]
work_dir = "/absolute/path/to/your/project"
mode = "default"

# --- Claude Code mode options ---
# "default", "acceptEdits" (alias: "edit"), "plan", "bypassPermissions" (alias: "yolo")
# allowed_tools = ["Read", "Grep", "Glob"]  # optional: pre-approve specific tools

# --- Gemini CLI mode options ---
# "default", "auto_edit", "yolo", "plan"
```

## Step 4: Configure Telegram

### Telegram — No public IP needed

Connection: Long Polling

**Setup steps:**
1. Message @BotFather on Telegram → send `/newbot`
2. Follow the prompts to set a bot name and username (must end with `bot`)
3. Copy the bot token

**Config:**

```toml
[[projects.platforms]]
type = "telegram"

[projects.platforms.options]
token = "1234567890:ABCdefGHIjklMNOpqrsTUVwxyz"
allow_from = "*"  # Allowed Telegram user IDs, e.g. "123456789,987654321"; "*" = all
# group_reply_all = false  # If true, respond to ALL group messages without @mention
# share_session_in_channel = false  # If true, all users in a group share one agent session
```

**Detailed guide:** [docs/telegram.md](docs/telegram.md)

## Step 5: Run cc-connect

**Important: If you are running inside a Claude Code session** (e.g., Claude Code helped you install and configure cc-connect), you must unset the `CLAUDECODE` environment variable before starting, otherwise Claude Code will refuse to launch as a subprocess:

```bash
unset CLAUDECODE && cc-connect
```

Alternatively, open a **separate terminal** and run cc-connect there — this avoids the issue entirely.

**Normal startup:**

```bash
# Run with config.toml in the current directory
cc-connect

# Or specify config path
cc-connect -config /path/to/config.toml

# Check version
cc-connect --version
```

You should see logs like:

```
level=INFO msg="platform started" project=my-project platform=telegram
level=INFO msg="engine started" project=my-project agent=claudecode platforms=1
level=INFO msg="cc-connect is running" projects=1
```

## Step 6: Chat Commands

Once running, send messages to your bot on Telegram. Available slash commands:

```
/new [name]      — Start a new session
/list            — List agent sessions
/switch <id>     — Resume an existing session
/current         — Show current active session
/history [n]     — Show last n messages (default 10)
/mode [name]     — View/switch permission mode (default/edit/plan/yolo)
/quiet           — Toggle thinking/tool progress messages
/allow <tool>    — Pre-allow a tool (next session)
/provider [...]  — Manage API providers (list/add/remove/switch)
/stop            — Stop current execution
/help            — Show available commands
```

During a session, Claude may ask for tool permissions. Reply:
- `allow` or `允许` — approve this request
- `deny` or `拒绝` — reject this request
- `allow all` or `允许所有` — auto-approve all remaining requests this session

## Multi-Project Setup

A single cc-connect process can manage multiple projects. Each project has its own agent, work directory, and platforms:

```toml
# First project — using Claude Code
[[projects]]
name = "backend"

[projects.agent]
type = "claudecode"

[projects.agent.options]
work_dir = "/path/to/backend"
mode = "default"

[[projects.platforms]]
type = "telegram"

[projects.platforms.options]
token = "xxx"

# Second project — using Gemini CLI
[[projects]]
name = "my-gemini-project"

[projects.agent]
type = "gemini"

[projects.agent.options]
work_dir = "/path/to/gemini-project"
mode = "yolo"    # "default" | "auto_edit" | "yolo" | "plan"

[[projects.platforms]]
type = "telegram"

[projects.platforms.options]
token = "xxx"
```

## Upgrade

### Check current version

```bash
cc-connect --version
```

### npm users

```bash
npm install -g @atticux/cc-connect
```

### Binary users

Check the latest release at https://github.com/AtticusZeller/cc-connect/releases and compare with your local version. To upgrade:

```bash
# Linux/macOS — replace with your platform suffix
curl -L -o /usr/local/bin/cc-connect https://github.com/AtticusZeller/cc-connect/releases/latest/download/cc-connect-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/')
chmod +x /usr/local/bin/cc-connect
```

### Source users

```bash
cd cc-connect
git pull
make build
```

After upgrading, restart the running cc-connect process.

## Step 7: Run as Background Service (Optional)

You can run cc-connect as a daemon managed by the OS init system (Linux systemd user service, macOS launchd LaunchAgent).

### Install daemon

```bash
cc-connect daemon install --config ~/.cc-connect/config.toml
```

You can also point the daemon at a directory that contains `config.toml`:

```bash
cc-connect daemon install --work-dir ~/.cc-connect
```

Optional flags: `--config PATH`, `--log-file PATH`, `--log-max-size N` (MB), `--work-dir DIR`, `--force` (overwrite existing unit). `--config` points to a config file, while `--work-dir` points to the directory containing `config.toml`.

### Control the service

```bash
cc-connect daemon start
cc-connect daemon stop
cc-connect daemon restart
cc-connect daemon status
```

### View logs

```bash
cc-connect daemon logs           # tail current log
cc-connect daemon logs -f         # follow (like tail -f)
cc-connect daemon logs -n 100     # last 100 lines
cc-connect daemon logs --log-file /path/to/log  # custom log file
```

Logs auto-rotate at the configured max size and keep one backup.

### Uninstall

```bash
cc-connect daemon uninstall
```

## Additional Features

The following additional features are available:

- **Claude Code**: Anthropic Claude Code CLI integration
- **Gemini CLI**: Google Gemini CLI integration
- **Voice Messages (STT)**: Speech-to-text via Whisper API (OpenAI / Groq / Qwen). Requires `ffmpeg` and `[speech]` config.
- **Voice Reply (TTS)**: Text-to-speech via Qwen TTS / OpenAI TTS. Requires `ffmpeg` and `[tts]` config.
- **Image Messages**: Send images to Claude Code for multimodal analysis
- **API Provider Management**: Runtime switching between API providers via `/provider` command or CLI
- **CLI Send**: `cc-connect send` to inject messages into active sessions from external processes

## Troubleshooting

- **"session already in use"** — A previous Claude Code process may still be running. Use `/new` to start a fresh session.
- **No response from bot** — Check `cc-connect` logs. Set `level = "debug"` in `[log]` for verbose output.
- **macOS binary won't open** — Run `xattr -d com.apple.quarantine cc-connect` to remove the quarantine flag.
