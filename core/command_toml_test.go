package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCommandRegistry_TomlIntegration tests TOML command file discovery with various scenarios
func TestCommandRegistry_TomlIntegration(t *testing.T) {
	dir := t.TempDir()

	// Create nested directory structure for namespace testing
	gitDir := filepath.Join(dir, "git")
	os.MkdirAll(gitDir, 0755)
	releaseDir := filepath.Join(dir, "release")
	os.MkdirAll(releaseDir, 0755)

	// Create various TOML command files
	testFiles := map[string]string{
		"simple.toml": `prompt = "Simple prompt"
description = "Simple command"`,

		"with_prefix.toml": `prompt = "With prefix"
description = "[Cmd] Command with prefix"`,

		"skill.toml": `prompt = "Skill prompt"
description = "[Skill] A skill command"`,

		"git/commit.toml": `prompt = "Generate commit message"
description = "[Cmd] Git commit helper"`,

		"git/push.toml": `prompt = "Push to remote"
description = "[Cmd] Git push"`,

		"git/branch.toml": `prompt = "Manage branches"
description = "[Cmd] Git branch management"`,

		"release/publish.toml": `prompt = "Publish release"
description = "[Skill] Release publisher"`,

		"release/notes.toml": `prompt = "Generate release notes"
description = "[Cmd] Release notes"`,
	}

	for path, content := range testFiles {
		fullPath := filepath.Join(dir, path)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create %s: %v", path, err)
		}
	}

	r := NewCommandRegistry()
	r.SetAgentDirs([]string{dir})

	tests := []struct {
		name     string
		cmd      string
		wantDesc string
	}{
		{"simple command", "simple", "Simple command"},
		{"with prefix", "with_prefix", "[Cmd] Command with prefix"},
		{"skill command", "skill", "[Skill] A skill command"},
		{"namespaced git:commit", "git:commit", "[Cmd] Git commit helper"},
		{"namespaced git:push", "git:push", "[Cmd] Git push"},
		{"namespaced git:branch", "git:branch", "[Cmd] Git branch management"},
		{"namespaced release:publish", "release:publish", "[Skill] Release publisher"},
		{"namespaced release:notes", "release:notes", "[Cmd] Release notes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, ok := r.Resolve(tt.cmd)
			if !ok {
				t.Fatalf("Failed to resolve %q", tt.cmd)
			}
			if cmd.Description != tt.wantDesc {
				t.Errorf("Description = %q, want %q", cmd.Description, tt.wantDesc)
			}
			if cmd.Source != "agent" {
				t.Errorf("Source = %q, want 'agent'", cmd.Source)
			}
		})
	}

	// Test ListAll includes all commands
	all := r.ListAll()
	if len(all) != 8 {
		t.Errorf("ListAll returned %d commands, want 8", len(all))
	}

}

// TestCommandRegistry_TomlOnlyFilter tests that SetAcceptedExts filters correctly
func TestCommandRegistry_TomlOnlyFilter(t *testing.T) {
	dir := t.TempDir()

	// Create both .md and .toml files
	os.WriteFile(filepath.Join(dir, "test.md"), []byte("MD content"), 0644)
	os.WriteFile(filepath.Join(dir, "test.toml"), []byte(`prompt = "TOML prompt"
description = "TOML command"`), 0644)

	r := NewCommandRegistry()
	r.SetAcceptedExts([]string{".toml"}) // Only accept .toml
	r.SetAgentDirs([]string{dir})

	cmd, ok := r.Resolve("test")
	if !ok {
		t.Fatal("Expected to resolve 'test'")
	}
	// Should get TOML content, not MD
	if cmd.Prompt != "TOML prompt" {
		t.Errorf("Prompt = %q, want 'TOML prompt' (.md should be filtered)", cmd.Prompt)
	}

	// ListAll should only show 1 command
	all := r.ListAll()
	if len(all) != 1 {
		t.Errorf("ListAll returned %d commands, want 1", len(all))
	}
}

// TestCommandRegistry_MdTomlPriority tests that .md overrides .toml
func TestCommandRegistry_MdTomlPriority(t *testing.T) {
	dir := t.TempDir()

	// Create both files with same name
	os.WriteFile(filepath.Join(dir, "test.md"), []byte("MD wins"), 0644)
	os.WriteFile(filepath.Join(dir, "test.toml"), []byte(`prompt = "TOML loses"
description = "TOML command"`), 0644)

	r := NewCommandRegistry()
	r.SetAgentDirs([]string{dir})

	cmd, ok := r.Resolve("test")
	if !ok {
		t.Fatal("Expected to resolve 'test'")
	}
	// MD should win
	if cmd.Prompt != "MD wins" {
		t.Errorf("Prompt = %q, want 'MD wins' (.md should override .toml)", cmd.Prompt)
	}
}

// TestCommandRegistry_TomlMalformed tests handling of malformed TOML files
func TestCommandRegistry_TomlMalformed(t *testing.T) {
	dir := t.TempDir()

	// Create a malformed TOML file
	os.WriteFile(filepath.Join(dir, "bad.toml"), []byte(`prompt = "missing closing quote`), 0644)

	r := NewCommandRegistry()
	r.SetAgentDirs([]string{dir})

	// Should not crash; malformed files should be silently skipped
	all := r.ListAll()
	if len(all) != 0 {
		t.Errorf("Malformed TOML should be skipped, got %d commands", len(all))
	}
}

// TestCommandRegistry_TomlEmptyDescription tests commands with optional description
func TestCommandRegistry_TomlEmptyDescription(t *testing.T) {
	dir := t.TempDir()

	// Create TOML with only prompt (description is optional)
	os.WriteFile(filepath.Join(dir, "no_desc.toml"), []byte(`prompt = "Just a prompt"`), 0644)

	r := NewCommandRegistry()
	r.SetAgentDirs([]string{dir})

	cmd, ok := r.Resolve("no_desc")
	if !ok {
		t.Fatal("Expected to resolve 'no_desc'")
	}
	if cmd.Prompt != "Just a prompt" {
		t.Errorf("Prompt = %q, want 'Just a prompt'", cmd.Prompt)
	}
	// Description should be empty when not specified
	if cmd.Description != "" {
		t.Errorf("Description = %q, want empty string", cmd.Description)
	}
}

// TestCommandRegistry_MultipleAgentDirs tests scanning multiple directories
func TestCommandRegistry_MultipleAgentDirs(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	os.WriteFile(filepath.Join(dir1, "cmd1.toml"), []byte(`prompt = "Dir1"`), 0644)
	os.WriteFile(filepath.Join(dir2, "cmd2.toml"), []byte(`prompt = "Dir2"`), 0644)

	r := NewCommandRegistry()
	r.SetAgentDirs([]string{dir1, dir2})

	// Both commands should be found
	_, ok1 := r.Resolve("cmd1")
	_, ok2 := r.Resolve("cmd2")

	if !ok1 {
		t.Error("cmd1 should be resolved from dir1")
	}
	if !ok2 {
		t.Error("cmd2 should be resolved from dir2")
	}
}

// TestCommandRegistry_TomlDeepNesting tests deeply nested namespace paths
func TestCommandRegistry_TomlDeepNesting(t *testing.T) {
	dir := t.TempDir()

	// Create deeply nested directory
	deepDir := filepath.Join(dir, "a", "b", "c", "d")
	os.MkdirAll(deepDir, 0755)
	os.WriteFile(filepath.Join(deepDir, "deep.toml"), []byte(`prompt = "Deep"
description = "Deep command"`), 0644)

	r := NewCommandRegistry()
	r.SetAgentDirs([]string{dir})

	cmd, ok := r.Resolve("a:b:c:d:deep")
	if !ok {
		t.Fatal("Expected to resolve 'a:b:c:d:deep'")
	}
	if cmd.Name != "a:b:c:d:deep" {
		t.Errorf("Name = %q, want 'a:b:c:d:deep'", cmd.Name)
	}
}

// TestCommandRegistry_TomlSpecialCharsInDescription tests handling of special characters
func TestCommandRegistry_TomlSpecialCharsInDescription(t *testing.T) {
	dir := t.TempDir()

	// TOML with special characters
	content := `prompt = "Test"
description = "Command with [brackets] and <special> chars"`
	os.WriteFile(filepath.Join(dir, "special.toml"), []byte(content), 0644)

	r := NewCommandRegistry()
	r.SetAgentDirs([]string{dir})

	cmd, ok := r.Resolve("special")
	if !ok {
		t.Fatal("Expected to resolve 'special'")
	}
	expectedDesc := "Command with [brackets] and <special> chars"
	if cmd.Description != expectedDesc {
		t.Errorf("Description = %q, want %q", cmd.Description, expectedDesc)
	}
}

// TestCommandRegistry_TomlMultilinePrompt tests multiline TOML strings
func TestCommandRegistry_TomlMultilinePrompt(t *testing.T) {
	dir := t.TempDir()

	// TOML with multiline prompt
	content := `prompt = """
This is a
multiline prompt
with multiple lines
"""
description = "Multiline test"
`
	os.WriteFile(filepath.Join(dir, "multiline.toml"), []byte(content), 0644)

	r := NewCommandRegistry()
	r.SetAgentDirs([]string{dir})

	cmd, ok := r.Resolve("multiline")
	if !ok {
		t.Fatal("Expected to resolve 'multiline'")
	}
	if !strings.Contains(cmd.Prompt, "multiline prompt") {
		t.Errorf("Prompt should contain 'multiline prompt', got %q", cmd.Prompt)
	}
}
