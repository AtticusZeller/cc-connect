package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{
			name:    "requires at least one project",
			cfg:     Config{},
			wantErr: "at least one [[projects]] entry is required",
		},
		{
			name: "requires project name",
			cfg: Config{
				Projects: []ProjectConfig{
					validProject(""),
				},
			},
			wantErr: `projects[0].name is required`,
		},
		{
			name: "requires agent type",
			cfg: Config{
				Projects: []ProjectConfig{
					func() ProjectConfig {
						p := validProject("demo")
						p.Agent.Type = ""
						return p
					}(),
				},
			},
			wantErr: `projects[0].agent.type is required`,
		},
		{
			name: "requires at least one platform",
			cfg: Config{
				Projects: []ProjectConfig{
					func() ProjectConfig {
						p := validProject("demo")
						p.Platforms = nil
						return p
					}(),
				},
			},
			wantErr: `projects[0] needs at least one [[projects.platforms]]`,
		},
		{
			name: "requires platform type",
			cfg: Config{
				Projects: []ProjectConfig{
					func() ProjectConfig {
						p := validProject("demo")
						p.Platforms[0].Type = ""
						return p
					}(),
				},
			},
			wantErr: `projects[0].platforms[0].type is required`,
		},
		{
			name: "multi workspace requires base dir",
			cfg: Config{
				Projects: []ProjectConfig{
					func() ProjectConfig {
						p := validProject("demo")
						p.Mode = "multi-workspace"
						return p
					}(),
				},
			},
			wantErr: `project "demo": multi-workspace mode requires base_dir`,
		},
		{
			name: "multi workspace rejects work dir",
			cfg: Config{
				Projects: []ProjectConfig{
					func() ProjectConfig {
						p := validProject("demo")
						p.Mode = "multi-workspace"
						p.BaseDir = "~/workspace"
						p.Agent.Options["work_dir"] = "/tmp/demo"
						return p
					}(),
				},
			},
			wantErr: `project "demo": multi-workspace mode conflicts with agent work_dir`,
		},
		{
			name: "accepts valid config",
			cfg: Config{
				Projects: []ProjectConfig{validProject("demo")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("validate() unexpected error: %v", err)
				}
				return
			}
			assertErrContains(t, err, tt.wantErr)
		})
	}
}

func TestLoad_DefaultsDataDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	cfgPath := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(cfgPath, []byte(baseConfigTOML), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	want := filepath.Join(dir, ".cc-connect")
	if cfg.DataDir != want {
		t.Fatalf("Load() data_dir = %q, want %q", cfg.DataDir, want)
	}
}

func TestListProjects(t *testing.T) {
	writeTestConfig(t, baseConfigTOML)

	names, err := ListProjects()
	if err != nil {
		t.Fatalf("ListProjects() error: %v", err)
	}
	if len(names) != 1 || names[0] != "demo" {
		t.Fatalf("ListProjects() = %#v, want [demo]", names)
	}
}

func TestSaveLanguage(t *testing.T) {
	writeTestConfig(t, baseConfigTOML)

	if err := SaveLanguage("zh"); err != nil {
		t.Fatalf("SaveLanguage() error: %v", err)
	}

	cfg := readTestConfig(t)
	if cfg.Language != "zh" {
		t.Fatalf("Language = %q, want zh", cfg.Language)
	}
}

func TestProviderConfig_SaveActiveProviderAndGetProjectProviders(t *testing.T) {
	writeTestConfig(t, providerConfigTOML)

	if err := SaveActiveProvider("demo", "backup"); err != nil {
		t.Fatalf("SaveActiveProvider() error: %v", err)
	}

	providers, active, err := GetProjectProviders("demo")
	if err != nil {
		t.Fatalf("GetProjectProviders() error: %v", err)
	}
	if active != "backup" {
		t.Fatalf("active provider = %q, want backup", active)
	}
	if len(providers) != 2 {
		t.Fatalf("provider count = %d, want 2", len(providers))
	}
}

func TestProviderConfig_AddAndRemove(t *testing.T) {
	writeTestConfig(t, providerConfigTOML)

	newProvider := ProviderConfig{Name: "relay", APIKey: "sk-relay", BaseURL: "https://example.com"}
	if err := AddProviderToConfig("demo", newProvider); err != nil {
		t.Fatalf("AddProviderToConfig() error: %v", err)
	}
	if err := AddProviderToConfig("demo", newProvider); err == nil {
		t.Fatal("AddProviderToConfig() duplicate provider: expected error")
	}

	cfg := readTestConfig(t)
	if len(cfg.Projects[0].Agent.Providers) != 3 {
		t.Fatalf("provider count after add = %d, want 3", len(cfg.Projects[0].Agent.Providers))
	}

	if err := RemoveProviderFromConfig("demo", "relay"); err != nil {
		t.Fatalf("RemoveProviderFromConfig() error: %v", err)
	}
	if err := RemoveProviderFromConfig("demo", "relay"); err == nil {
		t.Fatal("RemoveProviderFromConfig() missing provider: expected error")
	}
}

func TestCommandConfig_AddAndRemove(t *testing.T) {
	writeTestConfig(t, baseConfigTOML)

	cmd := CommandConfig{Name: "review", Description: "code review", Prompt: "review {{args}}"}
	if err := AddCommand(cmd); err != nil {
		t.Fatalf("AddCommand() error: %v", err)
	}
	if err := AddCommand(cmd); err == nil {
		t.Fatal("AddCommand() duplicate command: expected error")
	}

	cfg := readTestConfig(t)
	if len(cfg.Commands) != 1 || cfg.Commands[0].Name != "review" {
		t.Fatalf("commands after add = %#v, want one review command", cfg.Commands)
	}

	if err := RemoveCommand("review"); err != nil {
		t.Fatalf("RemoveCommand() error: %v", err)
	}
	if err := RemoveCommand("review"); err == nil {
		t.Fatal("RemoveCommand() missing command: expected error")
	}
}

func TestAliasConfig_AddAndRemove(t *testing.T) {
	writeTestConfig(t, baseConfigTOML)

	if err := AddAlias(AliasConfig{Name: "帮助", Command: "/help"}); err != nil {
		t.Fatalf("AddAlias() error: %v", err)
	}
	if err := AddAlias(AliasConfig{Name: "帮助", Command: "/list"}); err != nil {
		t.Fatalf("AddAlias() update error: %v", err)
	}

	cfg := readTestConfig(t)
	if len(cfg.Aliases) != 1 || cfg.Aliases[0].Command != "/list" {
		t.Fatalf("aliases after update = %#v, want one updated alias", cfg.Aliases)
	}

	if err := RemoveAlias("帮助"); err != nil {
		t.Fatalf("RemoveAlias() error: %v", err)
	}
	if err := RemoveAlias("帮助"); err == nil {
		t.Fatal("RemoveAlias() missing alias: expected error")
	}
}

func TestDisplayConfig_Save(t *testing.T) {
	writeTestConfig(t, baseConfigTOML)

	thinking := 120
	tool := 240
	if err := SaveDisplayConfig(&thinking, &tool); err != nil {
		t.Fatalf("SaveDisplayConfig() error: %v", err)
	}

	cfg := readTestConfig(t)
	if cfg.Display.ThinkingMaxLen == nil || *cfg.Display.ThinkingMaxLen != 120 {
		t.Fatalf("ThinkingMaxLen = %#v, want 120", cfg.Display.ThinkingMaxLen)
	}
	if cfg.Display.ToolMaxLen == nil || *cfg.Display.ToolMaxLen != 240 {
		t.Fatalf("ToolMaxLen = %#v, want 240", cfg.Display.ToolMaxLen)
	}

	thinking = 360
	if err := SaveDisplayConfig(&thinking, nil); err != nil {
		t.Fatalf("SaveDisplayConfig() second update error: %v", err)
	}

	cfg = readTestConfig(t)
	if cfg.Display.ThinkingMaxLen == nil || *cfg.Display.ThinkingMaxLen != 360 {
		t.Fatalf("ThinkingMaxLen after update = %#v, want 360", cfg.Display.ThinkingMaxLen)
	}
	if cfg.Display.ToolMaxLen == nil || *cfg.Display.ToolMaxLen != 240 {
		t.Fatalf("ToolMaxLen after nil update = %#v, want 240", cfg.Display.ToolMaxLen)
	}
}

func TestTTSConfig_SaveMode(t *testing.T) {
	writeTestConfig(t, baseConfigTOML)

	if err := SaveTTSMode("always"); err != nil {
		t.Fatalf("SaveTTSMode() error: %v", err)
	}

	cfg := readTestConfig(t)
	if cfg.TTS.TTSMode != "always" {
		t.Fatalf("TTSMode = %q, want always", cfg.TTS.TTSMode)
	}
}

const attachmentSendConfigFixture = `
attachment_send = "off"

[[projects]]
name = "alpha"

[projects.agent]
type = "claudecode"

[projects.agent.options]
work_dir = "/tmp/alpha"

[[projects.platforms]]
type = "telegram"

[projects.platforms.options]
bot_token = "token_xxx"
`

const relayConfigFixture = `
[relay]
timeout_secs = 300

[[projects]]
name = "alpha"

[projects.agent]
type = "claudecode"

[projects.agent.options]
work_dir = "/tmp/alpha"

[[projects.platforms]]
type = "telegram"

[projects.platforms.options]
bot_token = "token_xxx"
`

const relayConfigNegativeFixture = `
[relay]
timeout_secs = -1

[[projects]]
name = "alpha"

[projects.agent]
type = "claudecode"

[projects.agent.options]
work_dir = "/tmp/alpha"

[[projects.platforms]]
type = "telegram"

[projects.platforms.options]
bot_token = "token_xxx"
`







func TestLoad_DefaultsAttachmentSendToOn(t *testing.T) {
	configPath := writeConfigFixture(t, relayConfigFixture)

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.AttachmentSend != "on" {
		t.Fatalf("cfg.AttachmentSend = %q, want %q", cfg.AttachmentSend, "on")
	}
}

func TestLoad_ParsesAttachmentSendOff(t *testing.T) {
	configPath := writeConfigFixture(t, attachmentSendConfigFixture)

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.AttachmentSend != "off" {
		t.Fatalf("cfg.AttachmentSend = %q, want %q", cfg.AttachmentSend, "off")
	}
}

func validProject(name string) ProjectConfig {
	return ProjectConfig{
		Name: name,
		Agent: AgentConfig{
			Type:    "claudecode",
			Options: map[string]any{"mode": "default"},
		},
		Platforms: []PlatformConfig{
			{Type: "telegram", Options: map[string]any{"token": "test-token"}},
		},
	}
}

func assertErrContains(t *testing.T, err error, want string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error containing %q, got nil", want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("error = %q, want substring %q", err.Error(), want)
	}
}

func writeTestConfig(t *testing.T, content string) {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	oldPath := ConfigPath
	ConfigPath = path
	t.Cleanup(func() {
		ConfigPath = oldPath
	})
}

func readTestConfig(t *testing.T) Config {
	t.Helper()

	data, err := os.ReadFile(ConfigPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("parse config: %v", err)
	}
	return cfg
}

func TestLoadRelayTimeoutConfig(t *testing.T) {
	configPath := writeConfigFixture(t, relayConfigFixture)

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.Relay.TimeoutSecs == nil {
		t.Fatal("cfg.Relay.TimeoutSecs = nil, want non-nil")
	}
	if *cfg.Relay.TimeoutSecs != 300 {
		t.Fatalf("cfg.Relay.TimeoutSecs = %d, want 300", *cfg.Relay.TimeoutSecs)
	}
}

func TestLoadRejectsNegativeRelayTimeout(t *testing.T) {
	configPath := writeConfigFixture(t, relayConfigNegativeFixture)

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected error for negative relay timeout, got nil")
	}
	if !strings.Contains(err.Error(), "relay.timeout_secs must be >= 0") {
		t.Fatalf("error = %q, want contains %q", err.Error(), "relay.timeout_secs must be >= 0")
	}
}
func writeConfigFixture(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config fixture: %v", err)
	}
	return path
}

func patchConfigPath(t *testing.T, path string) {
	t.Helper()
	prev := ConfigPath
	ConfigPath = path
	t.Cleanup(func() {
		ConfigPath = prev
	})
}

func readConfigFixture(t *testing.T, path string) *Config {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read config fixture: %v", err)
	}
	cfg := &Config{}
	if err := toml.Unmarshal(data, cfg); err != nil {
		t.Fatalf("parse config fixture: %v", err)
	}
	return cfg
}

func stringMapValue(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

const baseConfigTOML = `
[[projects]]
name = "demo"

[projects.agent]
type = "claudecode"

[projects.agent.options]
mode = "default"

[[projects.platforms]]
type = "telegram"

[projects.platforms.options]
token = "test-token"
`

const providerConfigTOML = `
[[projects]]
name = "demo"

[projects.agent]
type = "claudecode"

[projects.agent.options]
mode = "default"
provider = "primary"

[[projects.agent.providers]]
name = "primary"
api_key = "sk-primary"

[[projects.agent.providers]]
name = "backup"
api_key = "sk-backup"

[[projects.platforms]]
type = "telegram"

[projects.platforms.options]
token = "test-token"
`



const preserveFormatFixture = `# top comment should stay
custom_top = "keep_me"

[[projects]]
name = "alpha"

[projects.agent]
type = "claudecode"

[projects.agent.options]
work_dir = "/tmp/alpha"

[[projects.platforms]]
type = "telegram"

[projects.platforms.options]
app_id = "old_app" # keep inline comment
app_secret = "old_secret"
custom_option = "still_here"
`
