package config

import (
	"slices"
	"testing"
)

func TestValidateSourceAcceptsLocalCIDRFile(t *testing.T) {
	t.Parallel()
	if err := validateSource(Source{Kind: "cidr-file", Path: "sources/tailscale.ipcidr.txt"}); err != nil {
		t.Fatal(err)
	}
}

func TestValidateSourceRejectsEscapingLocalFile(t *testing.T) {
	t.Parallel()
	if err := validateSource(Source{Kind: "cidr-file", Path: "../tailscale.ipcidr.txt"}); err == nil {
		t.Fatal("escaping local path did not fail validation")
	}
}

func TestProjectConfigUsesTailscaleRuleSet(t *testing.T) {
	t.Parallel()
	cfg, err := Load("../../config/rulesets.json")
	if err != nil {
		t.Fatal(err)
	}

	wantPriority := []string{"tailscale", "lan", "cn", "crypto", "us", "global"}
	if !slices.Equal(cfg.Priority, wantPriority) {
		t.Fatalf("got priority %v, want %v", cfg.Priority, wantPriority)
	}
	wantSourcePaths := []string{"sources/tailscale.domain.txt", "sources/tailscale.ipcidr.txt"}
	gotSourcePaths := []string{cfg.RuleSets[0].Sources[0].Path, cfg.RuleSets[0].Sources[1].Path}
	if !slices.Equal(gotSourcePaths, wantSourcePaths) {
		t.Fatalf("got Tailscale source paths %v, want %v", gotSourcePaths, wantSourcePaths)
	}

	wantAssertions := map[string]struct{}{
		"server.example.ts.net": {},
		"100.64.0.1":            {},
		"fd7a:115c:a1e0::1":     {},
	}
	for _, assertion := range cfg.Assertions {
		target := assertion.Domain
		if target == "" {
			target = assertion.IP
		}
		if _, expected := wantAssertions[target]; !expected {
			continue
		}
		if assertion.RuleSet != "tailscale" || !assertion.Exclusive {
			t.Errorf("Tailscale assertion for %q is not exclusive to tailscale", target)
		}
		delete(wantAssertions, target)
	}
	if len(wantAssertions) != 0 {
		t.Fatalf("missing Tailscale assertions: %v", wantAssertions)
	}
}
