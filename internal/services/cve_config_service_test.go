package services

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseLanguagesSupportsFrontendEcosystems(t *testing.T) {
	ecosystems := []string{
		"Maven",
		"npm",
		"PyPI",
		"Go",
		"crates.io",
		"NuGet",
		"RubyGems",
		"Packagist",
		"Pub",
		"SwiftPM",
		"Alpine",
		"Debian",
		"Ubuntu",
		"Red Hat",
		"Rocky Linux",
		"AlmaLinux",
		"SUSE",
		"openSUSE",
		"Oracle Linux",
		"Amazon Linux",
		"Photon OS",
		"GitHub Actions",
		"Kubernetes",
		"Android",
		"Bitnami",
		"OSS-Fuzz",
		"Chainguard",
	}

	parts := make([]string, 0, len(ecosystems))
	for i, ecosystem := range ecosystems {
		parts = append(parts, fmt.Sprintf("%s:package-%d@1.0.0", ecosystem, i))
	}

	queries, err := parseLanguages(strings.Join(parts, ","))
	if err != nil {
		t.Fatalf("parseLanguages() error = %v", err)
	}

	if len(queries) != len(ecosystems) {
		t.Fatalf("parseLanguages() parsed %d ecosystems, want %d", len(queries), len(ecosystems))
	}

	for i, ecosystem := range ecosystems {
		if queries[i].Package.Ecosystem != ecosystem {
			t.Fatalf("query[%d] ecosystem = %q, want %q", i, queries[i].Package.Ecosystem, ecosystem)
		}
	}
}

func TestMapEcosystemCanonicalizesAliases(t *testing.T) {
	tests := map[string]string{
		"python":         "PyPI",
		"pip":            "PyPI",
		"node":           "npm",
		"nodejs":         "npm",
		"java":           "Maven",
		"maven":          "Maven",
		"go":             "Go",
		"rust":           "crates.io",
		"dotnet":         "NuGet",
		"nuget":          "NuGet",
		"ruby":           "RubyGems",
		"php":            "Packagist",
		"red hat":        "Red Hat",
		"github actions": "GitHub Actions",
		"opensuse":       "openSUSE",
	}

	for input, want := range tests {
		if got := mapEcosystem(input); got != want {
			t.Fatalf("mapEcosystem(%q) = %q, want %q", input, got, want)
		}
		if !isValidEcosystem(want) {
			t.Fatalf("isValidEcosystem(%q) = false, want true", want)
		}
	}

	if isValidEcosystem("unsupported") {
		t.Fatal("isValidEcosystem(unsupported) = true, want false")
	}
}
