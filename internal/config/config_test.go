package config

import (
	"slices"
	"strings"
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

func TestValidateSourceClashFile(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		source Source
		valid  bool
	}{
		{"local", Source{Kind: "clash-file", Path: "sources/cloudflared.classical.txt"}, true},
		{"missing path", Source{Kind: "clash-file"}, false},
		{"parent", Source{Kind: "clash-file", Path: "../cloudflared.txt"}, false},
		{"nested parent", Source{Kind: "clash-file", Path: "sources/../../cloudflared.txt"}, false},
		{"absolute", Source{Kind: "clash-file", Path: "/tmp/cloudflared.txt"}, false},
		{"remote url", Source{Kind: "clash-file", Path: "sources/cloudflared.txt", URL: "https://example.com/rules"}, false},
		{"selection", Source{Kind: "clash-file", Path: "sources/cloudflared.txt", AllLists: true}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateSource(tc.source); (err == nil) != tc.valid {
				t.Fatalf("validateSource(%+v) = %v, want valid=%t", tc.source, err, tc.valid)
			}
		})
	}
}

func TestProjectConfigUsesTailnetAndDirectRuleSets(t *testing.T) {
	t.Parallel()
	cfg, err := Load("../../config/rulesets.json")
	if err != nil {
		t.Fatal(err)
	}
	wantPriority := []string{"tailnet", "direct", "us", "global"}
	if !slices.Equal(cfg.Priority, wantPriority) {
		t.Fatalf("got priority %v, want %v", cfg.Priority, wantPriority)
	}
	if len(cfg.RuleSets) != len(wantPriority) {
		t.Fatalf("got %d rulesets, want %d", len(cfg.RuleSets), len(wantPriority))
	}
	wantDirectSources := []string{
		"Tailscale:public-services",
		"SukkaW/Surge:lan-domain", "SukkaW/Surge:lan-ip",
		"v2fly/domain-list-community:private", "IANA:local-addresses",
		"felixonmars/dnsmasq-china-list:accelerated-domains.china.conf",
		"gaoyifan/china-operator-ip:china46.txt", "SukkaW/Surge:apple_cn",
		"SukkaW/Surge:direct", "SukkaW/Surge:domestic-domain",
		"SukkaW/Surge:domestic-ip", "SukkaW/Surge:microsoft_cdn",
		"v2fly/domain-list-community:{cn,elsevier,sciencedirect}", "custom:cn", "Cloudflare:tunnel",
	}
	var directSources []string
	for _, src := range cfg.RuleSets[1].Sources {
		directSources = append(directSources, src.Name)
	}
	if !slices.Equal(directSources, wantDirectSources) {
		t.Errorf("direct sources: got %v, want merged sources in order %v", directSources, wantDirectSources)
	}
	wantLocalSources := map[string][]Source{
		"tailnet": {{Kind: "domain-file", Path: "sources/tailnet.domain.txt"}, {Kind: "cidr-file", Path: "sources/tailnet.ipcidr.txt"}},
		"direct":  {{Kind: "domain-file", Path: "sources/tailscale.domain.txt"}, {Kind: "domain-file", Path: "overrides/cn.domain.txt", OptionalIfCovered: true}, {Kind: "clash-file", Path: "sources/cloudflared.classical.txt"}},
		"us":      {{Kind: "domain-file", Path: "sources/anthropic.domain.txt"}, {Kind: "cidr-file", Path: "sources/anthropic.ipcidr.txt"}},
	}
	for i, set := range cfg.RuleSets {
		if set.Name != wantPriority[i] {
			t.Errorf("ruleset %d: got %s, want %s", i, set.Name, wantPriority[i])
		}
		var local []Source
		for _, src := range set.Sources {
			if src.Path != "" {
				local = append(local, Source{Kind: src.Kind, Path: src.Path, OptionalIfCovered: src.OptionalIfCovered})
			}
		}
		if !slices.EqualFunc(local, wantLocalSources[set.Name], func(a, b Source) bool {
			return a.Kind == b.Kind && a.Path == b.Path && a.OptionalIfCovered == b.OptionalIfCovered
		}) {
			t.Errorf("%s local sources: got %v, want %v", set.Name, local, wantLocalSources[set.Name])
		}
		var wantExclude []string
		if i == 1 {
			wantExclude = []string{"tailnet"}
		} else if i > 1 {
			wantExclude = []string{"tailnet", "direct"}
		}
		if !slices.Equal(set.Exclude, wantExclude) {
			t.Errorf("%s exclusions: got %v, want %v", set.Name, set.Exclude, wantExclude)
		}
	}
	wantAssertions := map[string]string{
		"ts.net": "tailnet", "server.example.ts.net": "tailnet",
		"forgejo.liblaf.me": "tailnet", "jellyfin.liblaf.me": "tailnet", "oplist.liblaf.me": "tailnet", "qbt.liblaf.me": "tailnet",
		"100.64.0.1": "tailnet", "fd7a:115c:a1e0::1": "tailnet",
		"tailscale.com": "direct", "login.tailscale.com": "direct", "controlplane.tailscale.com": "direct", "derp1-all.tailscale.com": "direct",
		"tailscale.io": "direct", "controlplane.tailscale.io": "direct",
		"localhost": "direct", "dl.acm.org": "direct", "www.elsevier.com": "direct",
		"www.sciencedirect.com": "direct", "ieeexplore.ieee.org": "direct", "www.scopus.com": "direct", "shop.pockyt.io": "direct",
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
		// The direct +.liblaf.me override partially overlaps the four exact custom
		// hosts. Those assertions verify priority; suffixes and IPs are exclusive.
		wantExclusive := want == "tailnet" && !slices.Contains([]string{"forgejo.liblaf.me", "jellyfin.liblaf.me", "oplist.liblaf.me", "qbt.liblaf.me"}, target) || strings.HasSuffix(target, "tailscale.com") || strings.HasSuffix(target, "tailscale.io")
		if assertion.RuleSet != want || assertion.Exclusive != wantExclusive {
			t.Errorf("assertion for %q: got ruleset %s and exclusive=%t, want %s and exclusive=%t", target, assertion.RuleSet, assertion.Exclusive, want, wantExclusive)
		}
		delete(wantAssertions, target)
	}
	if len(wantAssertions) != 0 {
		t.Fatalf("missing assertions: %v", wantAssertions)
	}
}
