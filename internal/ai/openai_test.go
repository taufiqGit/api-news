package ai

import (
	"strings"
	"testing"
)

func TestStripFences(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"json fence", "```json\n{\"a\":1}\n```", "{\"a\":1}"},
		{"plain fence", "```\n{\"a\":1}\n```", "{\"a\":1}"},
		{"no fence", `{"a":1}`, `{"a":1}`},
		{"whitespace around", "  \n{\"a\":1}\n  ", `{"a":1}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripFences(tt.in)
			if got != tt.want {
				t.Fatalf("stripFences(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseResult(t *testing.T) {
	validJSON := `{
		"title": "World markets rally",
		"slug": "world-markets-rally",
		"excerpt": "Markets surged on Tuesday.",
		"content": "Markets surged on Tuesday as investors reacted.",
		"category": "economy",
		"tags": ["economy", "markets"]
	}`

	res, err := parseResult(validJSON)
	if err != nil {
		t.Fatalf("parseResult valid: %v", err)
	}
	if res.Title != "World markets rally" {
		t.Fatalf("title salah: %q", res.Title)
	}
	if res.Slug != "world-markets-rally" {
		t.Fatalf("slug salah: %q", res.Slug)
	}
	if res.Content == "" {
		t.Fatal("content kosong")
	}
	if res.Category == nil || *res.Category != "economy" {
		t.Fatalf("category salah: %v", res.Category)
	}
	if len(res.Tags) != 2 {
		t.Fatalf("tags length salah: %d", len(res.Tags))
	}

	// Dengan markdown fence.
	fenced := "```json\n" + validJSON + "\n```"
	if _, err := parseResult(fenced); err != nil {
		t.Fatalf("parseResult fenced: %v", err)
	}

	// Field wajib hilang → error.
	missing := `{"slug":"x","content":"body"}`
	if _, err := parseResult(missing); err == nil {
		t.Fatal("parseResult seharusnya error saat title hilang")
	}

	missingContent := `{"title":"t","slug":"x"}`
	if _, err := parseResult(missingContent); err == nil {
		t.Fatal("parseResult seharusnya error saat content hilang")
	}

	// JSON tidak valid → error.
	if _, err := parseResult("not json at all"); err == nil {
		t.Fatal("parseResult seharusnya error untuk non-JSON")
	}

	// Kategori null → Category nil.
	nullCategory := `{"title":"t","slug":"s","content":"body","category":null,"tags":[]}`
	r2, err := parseResult(nullCategory)
	if err != nil {
		t.Fatalf("parseResult null category: %v", err)
	}
	if r2.Category != nil {
		t.Fatalf("category harus nil, got %v", *r2.Category)
	}
}

func TestSystemPromptAntiSlop(t *testing.T) {
	// Pastikan prompt memuat kata terlarang yang diharapkan.
	banned := []string{"delve", "unleash", "game-changer", "landscape", "in today's fast-paced world"}
	for _, b := range banned {
		if !strings.Contains(systemPrompt, b) {
			t.Errorf("systemPrompt harus memuat kata terlarang %q", b)
		}
	}
}
