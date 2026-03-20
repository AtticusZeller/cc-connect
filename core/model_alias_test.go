package core

import "testing"

func TestResolveModelAlias_CaseInsensitive(t *testing.T) {
	models := []ModelOption{{Name: "gemini-1.5-pro", Alias: "Gemini"}}

	got := resolveModelAlias(models, "gemini")
	if got != "gemini-1.5-pro" {
		t.Fatalf("resolveModelAlias() = %q, want %q", got, "gemini-1.5-pro")
	}
}

func TestResolveModelAlias_NoMatchFallsBackToInput(t *testing.T) {
	models := []ModelOption{{Name: "gemini-1.5-pro", Alias: "gemini"}}

	got := resolveModelAlias(models, "gpt-5.4")
	if got != "gpt-5.4" {
		t.Fatalf("resolveModelAlias() = %q, want original input", got)
	}
}
