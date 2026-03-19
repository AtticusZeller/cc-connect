# Gemini CLI Stream-JSON Reference

This document provides a comprehensive reference for Gemini CLI's `--output-format stream-json` output format, used by cc-connect to drive the Gemini agent.

---

## Command Line

```bash
gemini -p "your prompt" --output-format stream-json
```

### Additional Flags

| Flag | Description |
|-------|-------------|
| `-y` / `--approval-mode yolo` | Auto-approve all tool calls (YOLO mode) |
| `--approval-mode auto_edit` | Auto-approve edit tools, ask for others |
| `--approval-mode plan` | Read-only plan mode |
| `--resume <session_id>` | Resume an existing session |
| `-m <model>` | Specify model (e.g., `gemini-2.5-flash`) |
| `--timeout <minutes>` | Session timeout (default: no timeout) |

---

## Stream-JSON Output Format

The output is newline-delimited JSON (JSONL). Each line is a complete JSON object representing an event.

### Event Types

There are 6 event types:

1. **`init`** - Session initialization
2. **`message`** - Text content from user or assistant
3. **`tool_use`** - Tool invocation
4. **`tool_result`** - Tool execution result
5. **`error`** - Error/warning message
6. **`result`** - Session completion with stats

---

### Base Fields (All Events)

| Field | Type | Description |
|--------|-------|-------------|
| `type` | string | Event type (see below) |
| `timestamp` | string | ISO 8601 timestamp (e.g., `2025-03-19T12:00:00.000Z`) |

---

## 1. Init Event

```json
{
  "type": "init",
  "timestamp": "2025-03-19T12:00:00.000Z",
  "session_id": "abc123-def456-...",
  "model": "gemini-2.5-flash"
}
```

| Field | Type | Description |
|--------|-------|-------------|
| `session_id` | string | Unique session identifier (UUID prefix) |
| `model` | string | Model name being used |

**Usage in cc-connect**: Store `session_id` for session resumption (`--resume` flag).

---

## 2. Message Event

```json
{
  "type": "message",
  "timestamp": "2025-03-19T12:00:01.000Z",
  "role": "assistant",
  "content": "Hello",
  "delta": true
}
```

| Field | Type | Description |
|--------|-------|-------------|
| `role` | string | Either `"user"` or `"assistant"` |
| `content` | string | Text content fragment |
| `delta` | boolean | `true` = streaming fragment, `false` = complete message |

### Delta Behavior

- **`delta: true`** - Incremental streaming fragment, emit immediately as `EventText`
- **`delta: false` (or missing)** - Complete message, buffered for later classification

**Usage in cc-connect**:
- Delta messages are sent immediately as text for real-time streaming
- Non-delta messages are buffered and classified as `EventThinking` or `EventText` based on following events (tool_use triggers thinking, result triggers text)

---

## 3. Tool Use Event

```json
{
  "type": "tool_use",
  "timestamp": "2025-03-19T12:00:02.000Z",
  "tool_name": "read_file",
  "tool_id": "read-abc123",
  "parameters": {
    "file_path": "/path/to/file.txt"
  }
}
```

| Field | Type | Description |
|--------|-------|-------------|
| `tool_name` | string | Tool identifier (see Tool Names below) |
| `tool_id` | string | Unique invocation ID (for matching with tool_result) |
| `parameters` | object | Tool-specific parameters (see Tool Parameters) |

**Usage in cc-connect**: Emit as `EventToolUse` with human-readable parameter summary.

---

## 4. Tool Result Event

```json
{
  "type": "tool_result",
  "timestamp": "2025-03-19T12:00:03.000Z",
  "tool_id": "read-abc123",
  "status": "success",
  "output": "File contents here..."
}
```

Or on error:

```json
{
  "type": "tool_result",
  "timestamp": "2025-03-19T12:00:03.000Z",
  "tool_id": "read-abc123",
  "status": "error",
  "error": {
    "type": "FILE_NOT_FOUND",
    "message": "File not found"
  }
}
```

| Field | Type | Description |
|--------|-------|-------------|
| `tool_id` | string | Matches `tool_id` from corresponding `tool_use` event |
| `status` | string | Either `"success"` or `"error"` |
| `output` | string | Tool output (present on success) |
| `error` | object (optional) | Error details on failure |

**Usage in cc-connect**: Emit as `EventToolResult`. On error, prefix output with `"Error: "`.

---

## 5. Error Event

```json
{
  "type": "error",
  "timestamp": "2025-03-19T12:00:04.000Z",
  "severity": "error",
  "message": "Something went wrong"
}
```

| Field | Type | Description |
|--------|-------|-------------|
| `severity` | string | Either `"warning"` or `"error"` |
| `message` | string | Error/warning message |

**Usage in cc-connect**: Emit as `EventError` with formatted message `[severity] message`.

---

## 6. Result Event

```json
{
  "type": "result",
  "timestamp": "2025-03-19T12:00:10.000Z",
  "status": "success",
  "stats": {
    "total_tokens": 100,
    "input_tokens": 50,
    "output_tokens": 50,
    "cached": 0,
    "input": 50,
    "duration_ms": 1200,
    "tool_calls": 2,
    "models": {
      "gemini-2.5-flash": {
        "total_tokens": 100,
        "input_tokens": 50,
        "output_tokens": 50,
        "cached": 0,
        "input": 50
      }
    }
  }
}
```

Or on error:

```json
{
  "type": "result",
  "timestamp": "2025-03-19T12:00:10.000Z",
  "status": "error",
  "error": {
    "type": "MaxSessionTurnsError",
    "message": "Maximum session turns exceeded"
  }
}
```

| Field | Type | Description |
|--------|-------|-------------|
| `status` | string | Either `"success"` or `"error"` |
| `error` | object (optional) | Error details on failure |
| `stats` | object (optional) | Session statistics on success |

### Stats Fields

| Field | Type | Description |
|--------|-------|-------------|
| `total_tokens` | number | Total tokens consumed across all models |
| `input_tokens` | number | Input/prompt tokens |
| `output_tokens` | number | Output/candidate tokens |
| `cached` | number | Cached tokens (cached prompts) |
| `input` | number | Number of input messages |
| `duration_ms` | number | Session duration in milliseconds |
| `tool_calls` | number | Total tool calls made |
| `models` | object | Per-model breakdown |

### Model Stats (within `models` object)

| Field | Type | Description |
|--------|-------|-------------|
| `total_tokens` | number | Total tokens for this model |
| `input_tokens` | number | Input tokens (alias: `prompt`) |
| `output_tokens` | number | Output tokens (alias: `candidates`) |
| `cached` | number | Cached tokens |
| `input` | number | Number of input messages |

**Usage in cc-connect**: Emit as `EventResult` with `Done: true` and `SessionID` set. Store stats for telemetry.

---

## Tool Names and Parameters

### Core Tools

| Tool Name | Primary Parameter | Additional Parameters | Description |
|------------|-------------------|----------------------|-------------|
| `read_file` | `file_path` / `path` | - | Read file content |
| `write_file` | `file_path` / `path`, `content` | - | Write file |
| `replace` | `file_path` / `path`, `old_string` / `old_str`, `new_string` / `new_str` | Replace text in file |
| `list_directory` | `dir_path` / `path`, `directory` | List directory contents |
| `run_shell_command` | `command` | - | Execute shell command |
| `Bash` | `command` | - | Alternative shell command tool |
| `shell` | `command` | - | Alternative shell command tool |

### Search Tools

| Tool Name | Primary Parameter | Additional Parameters | Description |
|------------|-------------------|----------------------|-------------|
| `glob` | `pattern` | - | Find files by pattern |
| `Grep` | `pattern` | - | Search text in files |
| `grep_search` | `pattern` | - | Search text in files (alternative) |

### Web Tools

| Tool Name | Primary Parameter | Additional Parameters | Description |
|------------|-------------------|----------------------|-------------|
| `google_web_search` | `query` | - | Google web search |
| `web_fetch` | `prompt` / `url` | - | Fetch web content |

### Memory Tools

| Tool Name | Primary Parameter | Additional Parameters | Description |
|------------|-------------------|----------------------|-------------|
| `save_memory` | `fact` | - | Save to memory |
| `write_todos` | `todos` | - | Write TODOs |

### Plan Mode Tools

| Tool Name | Primary Parameter | Additional Parameters | Description |
|------------|-------------------|----------------------|-------------|
| `enter_plan_mode` | `reason` | - | Enter planning mode |
| `exit_plan_mode` | `plan_path` | - | Exit planning mode |

### Other Tools

| Tool Name | Primary Parameter | Additional Parameters | Description |
|------------|-------------------|----------------------|-------------|
| `activate_skill` | `name` | - | Activate a skill |
| `ask_user` | `questions` | - | Ask user a question |
| `get_internal_docs` | - | - | Get internal documentation |

---

## Event Flow Example

A typical interaction looks like:

```
{"type":"init","timestamp":"...","session_id":"xxx","model":"gemini-2.5-flash"}
{"type":"message","timestamp":"...","role":"assistant","content":"Let me check...","delta":false}
{"type":"tool_use","timestamp":"...","tool_name":"read_file","tool_id":"read-123","parameters":{"file_path":"/tmp/test.txt"}}
{"type":"tool_result","timestamp":"...","tool_id":"read-123","status":"success","output":"file content"}
{"type":"message","timestamp":"...","role":"assistant","content":"Here is","delta":true}
{"type":"message","timestamp":"...","role":"assistant","content":"the result.","delta":true}
{"type":"result","timestamp":"...","status":"success","stats":{...}}
```

---

## Session Management

### Session Storage

Gemini CLI stores sessions in:
```
~/.gemini/tmp/<project-slug>/chats/session-<timestamp>-<uuid-prefix>.json
```

### Project Slug

The project slug is derived from the project path:
1. Read `~/.gemini/projects.json` (registry)
2. If project path matches, use registered slug
3. Otherwise, generate from directory name (lowercase, non-alphanumeric → hyphens)

### Session File Format

```json
{
  "sessionId": "session-id-here",
  "projectHash": "...",
  "startTime": "2025-03-19T12:00:00.000Z",
  "lastUpdated": "2025-03-19T12:05:00.000Z",
  "messages": [
    {
      "type": "user",
      "content": "Hello"
    },
    {
      "type": "user",
      "content": [{"text": "world"}]
    }
  ],
  "kind": "chat"
}
```

### Session Listing

cc-connect reads session files and filters:
- Excludes `kind: "subagent"` sessions (internal subagent sessions)
- Excludes sessions with no user messages
- Summarizes using first meaningful user message (first non-empty, non-braced line)
- Truncates summary to 60 characters for display

### Session Deletion

To delete a session, find the session file by matching `sessionId` and delete it.

---

## Environment Variables

| Variable | Description |
|-----------|-------------|
| `GOOGLE_API_KEY` | API key for Google Generative AI |
| `GEMINI_API_KEY` | Alternative API key variable |
| `GOOGLE_GENAI_USE_GCA` | Force GCA authentication (default: false) |
