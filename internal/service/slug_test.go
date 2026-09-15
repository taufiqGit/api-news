package service

import "testing"

func TestCleanSlug(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"lowercase and trim", "  Hello World  ", "hello-world"},
		{"punct to dash", "Hello, World! 2026", "hello-world-2026"},
		{"multiple spaces", "a   b\tc", "a-b-c"},
		{"strip leading/trailing dash", "--hello--", "hello"},
		{"empty fallback", "", "article"},
		{"only symbols", "!!!", "article"},
		{"unicode removed", "Ünïcode Tëst", "n-code-t-st"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanSlug(tt.in)
			if got != tt.want {
				t.Fatalf("cleanSlug(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestItoa(t *testing.T) {
	tests := []struct {
		in   int
		want string
	}{
		{0, "0"},
		{1, "1"},
		{2, "2"},
		{10, "10"},
		{99, "99"},
		{100, "100"},
		{12345, "12345"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := itoa(tt.in)
			if got != tt.want {
				t.Fatalf("itoa(%d) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
