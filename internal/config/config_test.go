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
		{"require attributes", Source{Kind: "clash-file", Path: "sources/cloudflared.txt", RequireAttributes: []string{"cn"}}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateSource(tc.source); (err == nil) != tc.valid {
				t.Fatalf("validateSource(%+v) = %v, want valid=%t", tc.source, err, tc.valid)
			}
		})
	}
}

func TestValidateSourceV2FlyAndRequireAttributes(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		source Source
		valid  bool
	}{
		{"named lists", Source{Kind: "v2fly", Lists: []string{"cn"}}, true},
		{"all lists", Source{Kind: "v2fly", AllLists: true}, true},
		{"all lists with attribute", Source{Kind: "v2fly", AllLists: true, RequireAttributes: []string{"cn"}}, true},
		{"missing selection", Source{Kind: "v2fly"}, false},
		{"lists and all_lists", Source{Kind: "v2fly", Lists: []string{"cn"}, AllLists: true}, false},
		{"empty attribute", Source{Kind: "v2fly", AllLists: true, RequireAttributes: []string{""}}, false},
		{"prefixed attribute", Source{Kind: "v2fly", AllLists: true, RequireAttributes: []string{"@cn"}}, false},
		{"comma in attribute", Source{Kind: "v2fly", AllLists: true, RequireAttributes: []string{"cn,ads"}}, false},
		{"whitespace attribute", Source{Kind: "v2fly", AllLists: true, RequireAttributes: []string{" cn "}}, false},
		{"clash with require_attributes", Source{Kind: "clash", URL: "https://example.com/rules.txt", RequireAttributes: []string{"cn"}}, false},
		{"iana-local with require_attributes", Source{Kind: "iana-local", URLs: []string{"https://example.com/a.csv"}, RequireAttributes: []string{"cn"}}, false},
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
	const allCNSource = "v2fly/domain-list-community:*@cn"
	wantDirectSources := []string{
		"Tailscale:public-services",
		"SukkaW/Surge:lan-domain", "SukkaW/Surge:lan-ip",
		"v2fly/domain-list-community:private", "IANA:local-addresses",
		"felixonmars/dnsmasq-china-list:accelerated-domains.china.conf",
		"gaoyifan/china-operator-ip:china46.txt", "SukkaW/Surge:apple_cn",
		"SukkaW/Surge:direct", "SukkaW/Surge:domestic-domain",
		"SukkaW/Surge:domestic-ip", "SukkaW/Surge:microsoft_cdn",
		"v2fly/domain-list-community:{cn,elsevier,sciencedirect}",
		allCNSource,
		"custom:cn", "Cloudflare:tunnel", "custom:direct",
	}
	var directSources []string
	for _, src := range cfg.RuleSets[1].Sources {
		directSources = append(directSources, src.Name)
		if src.Name == allCNSource && (!src.AllLists || !slices.Equal(src.RequireAttributes, []string{"cn"})) {
			t.Errorf("%s must select all lists and require the cn attribute", src.Name)
		}
	}
	if !slices.Equal(directSources, wantDirectSources) {
		t.Errorf("direct sources: got %v, want merged sources in order %v", directSources, wantDirectSources)
	}
	wantLocalSources := map[string][]Source{
		"tailnet": {{Kind: "domain-file", Path: "sources/tailnet.domain.txt"}, {Kind: "cidr-file", Path: "sources/tailnet.ipcidr.txt"}},
		"direct":  {{Kind: "domain-file", Path: "sources/tailscale.domain.txt"}, {Kind: "domain-file", Path: "overrides/cn.domain.txt", OptionalIfCovered: true}, {Kind: "clash-file", Path: "sources/cloudflared.classical.txt"}, {Kind: "clash-file", Path: "sources/direct.classical.txt"}},
		"us":      {{Kind: "clash-file", Path: "sources/us.classical.txt"}, {Kind: "domain-file", Path: "sources/anthropic.domain.txt"}, {Kind: "cidr-file", Path: "sources/anthropic.ipcidr.txt"}},
	}
	for i, set := range cfg.RuleSets {
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
	// The direct +.liblaf.me override partially overlaps the four exact custom
	// hosts. Those assertions verify priority; suffixes and IPs are exclusive.
	wantAssertions := map[string]struct {
		ruleset   string
		exclusive bool
	}{
		"ts.net":                     {"tailnet", true},
		"server.example.ts.net":      {"tailnet", true},
		"forgejo.liblaf.me":          {"tailnet", false},
		"jellyfin.liblaf.me":         {"tailnet", false},
		"oplist.liblaf.me":           {"tailnet", false},
		"qbt.liblaf.me":              {"tailnet", false},
		"100.64.0.1":                 {"tailnet", true},
		"fd7a:115c:a1e0::1":          {"tailnet", true},
		"tailscale.com":              {"direct", true},
		"login.tailscale.com":        {"direct", true},
		"controlplane.tailscale.com": {"direct", true},
		"derp1-all.tailscale.com":    {"direct", true},
		"tailscale.io":               {"direct", true},
		"controlplane.tailscale.io":  {"direct", true},
		"localhost":                  {"direct", false},
		"dl.acm.org":                 {"direct", false},
		"www.elsevier.com":           {"direct", false},
		"www.sciencedirect.com":      {"direct", false},
		"ieeexplore.ieee.org":        {"direct", false},
		"www.scopus.com":             {"direct", false},
		"shop.pockyt.io":             {"direct", false},
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
		if assertion.RuleSet != want.ruleset || assertion.Exclusive != want.exclusive {
			t.Errorf("assertion for %q: got ruleset %s and exclusive=%t, want %s and exclusive=%t", target, assertion.RuleSet, assertion.Exclusive, want.ruleset, want.exclusive)
		}
		delete(wantAssertions, target)
	}
	if len(wantAssertions) != 0 {
		t.Fatalf("missing assertions: %v", wantAssertions)
	}
}
