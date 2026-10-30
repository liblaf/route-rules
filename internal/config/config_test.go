package config

import (
	"slices"
	"testing"
)

func TestValidateSourceAcceptsLocalCIDRFile(t *testing.T) {
	t.Parallel()
	if err := validateSource(Source{Kind: "cidr-file", Path: "sources/tailnet.ipcidr.txt"}); err != nil {
		t.Fatal(err)
	}
}

func TestValidateSourceRejectsEscapingLocalFile(t *testing.T) {
	t.Parallel()
	if err := validateSource(Source{Kind: "cidr-file", Path: "../tailnet.ipcidr.txt"}); err == nil {
		t.Fatal("escaping local path did not fail validation")
	}
}

func TestProjectConfigUsesTailnetAndTailscaleRuleSets(t *testing.T) {
	t.Parallel()
	cfg, err := Load("../../config/rulesets.json")
	if err != nil {
		t.Fatal(err)
	}
	wantPriority := []string{"tailnet", "tailscale", "lan", "cn", "us", "global"}
	if !slices.Equal(cfg.Priority, wantPriority) {
		t.Fatalf("got priority %v, want %v", cfg.Priority, wantPriority)
	}
	for i, want := range [][]Source{
		{{Kind: "domain-file", Path: "sources/tailnet.domain.txt"}, {Kind: "cidr-file", Path: "sources/tailnet.ipcidr.txt"}},
		{{Kind: "domain-file", Path: "sources/tailscale.domain.txt"}},
	} {
		got := cfg.RuleSets[i].Sources
		if len(got) != len(want) {
			t.Fatalf("%s sources: got %v, want %v", cfg.RuleSets[i].Name, got, want)
		}
		for j := range want {
			if got[j].Kind != want[j].Kind || got[j].Path != want[j].Path {
				t.Errorf("%s source %d: got %v, want %v", cfg.RuleSets[i].Name, j, got[j], want[j])
			}
		}
	}
	for _, set := range cfg.RuleSets[2:] {
		for _, excluded := range []string{"tailnet", "tailscale"} {
			if !slices.Contains(set.Exclude, excluded) {
				t.Errorf("%s must explicitly exclude %s", set.Name, excluded)
			}
		}
	}
	wantAssertions := map[string]string{
		"ts.net": "tailnet", "server.example.ts.net": "tailnet",
		"forgejo.liblaf.me": "tailnet", "jellyfin.liblaf.me": "tailnet", "oplist.liblaf.me": "tailnet", "qbt.liblaf.me": "tailnet",
		"100.64.0.1": "tailnet", "fd7a:115c:a1e0::1": "tailnet",
		"tailscale.com": "tailscale", "login.tailscale.com": "tailscale", "controlplane.tailscale.com": "tailscale", "derp1-all.tailscale.com": "tailscale",
		"tailscale.io": "tailscale", "controlplane.tailscale.io": "tailscale",
	}
	for _, assertion := range cfg.Assertions {
		target := assertion.Domain
		if target == "" {
			target = assertion.IP
		}
		want, expected := wantAssertions[target]
		if !expected {
			continue
		}
		// The CN +.liblaf.me override partially overlaps the four exact custom
		// hosts. Those assertions verify priority; suffixes and IPs are exclusive.
		wantExclusive := !slices.Contains([]string{"forgejo.liblaf.me", "jellyfin.liblaf.me", "oplist.liblaf.me", "qbt.liblaf.me"}, target)
		if assertion.RuleSet != want || assertion.Exclusive != wantExclusive {
			t.Errorf("assertion for %q: got ruleset %s and exclusive=%t, want %s and exclusive=%t", target, assertion.RuleSet, assertion.Exclusive, want, wantExclusive)
		}
		delete(wantAssertions, target)
	}
	if len(wantAssertions) != 0 {
		t.Fatalf("missing assertions: %v", wantAssertions)
	}
}
