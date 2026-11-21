package build

import (
	"context"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/liblaf/route-rules/internal/rules"
)

func TestVerifyAssertionsUsesPriorityAndFallback(t *testing.T) {
	t.Parallel()
	lanRule, err := rules.SuffixDomain("local")
	if err != nil {
		t.Fatal(err)
	}
	usRule, err := rules.SuffixDomain("openai.com")
	if err != nil {
		t.Fatal(err)
	}
	cfg := Config{
		Priority: []string{"lan", "us", "global"},
		Fallback: "global",
		Assertions: []Assertion{
			{Domain: "router.local", RuleSet: "lan"},
			{Domain: "api.openai.com", RuleSet: "us"},
			{Domain: "unknown.example", RuleSet: "global"},
		},
	}
	built := Result{RuleSets: []RuleSet{
		{Name: "lan", Rules: rules.Collection{Domains: rules.DomainSet{lanRule}}},
		{Name: "us", Rules: rules.Collection{Domains: rules.DomainSet{usRule}}},
		{Name: "global"},
	}}
	if err := verifyAssertions(cfg, built); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyAssertionsSupportsExclusiveIPMembership(t *testing.T) {
	t.Parallel()
	tailnetPrefix := netip.MustParsePrefix("100.64.0.0/10")
	cfg := Config{
		Priority:   []string{"tailnet", "lan"},
		Fallback:   "lan",
		Assertions: []Assertion{{IP: "100.64.0.1", RuleSet: "tailnet", Exclusive: true}},
	}
	built := Result{RuleSets: []RuleSet{
		{Name: "tailnet", Rules: rules.Collection{Prefixes: rules.PrefixSet{tailnetPrefix}}},
		{Name: "lan"},
	}}
	if err := verifyAssertions(cfg, built); err != nil {
		t.Fatal(err)
	}
	built.RuleSets[1].Rules = rules.Collection{Prefixes: rules.PrefixSet{tailnetPrefix}}
	if err := verifyAssertions(cfg, built); err == nil || !strings.Contains(err.Error(), "expected exclusive membership") {
		t.Fatalf("got error %v, expected exclusive-membership failure", err)
	}
}

func TestVerifyAssertionsUsesDestinationPortAndPreservesFallback(t *testing.T) {
	t.Parallel()
	port, err := rules.NewClassical("DST-PORT", "7844")
	if err != nil {
		t.Fatal(err)
	}
	globalDomain, err := rules.SuffixDomain("argotunnel.com")
	if err != nil {
		t.Fatal(err)
	}
	cfg := Config{
		Priority: []string{"direct", "global"},
		Fallback: "global",
		Assertions: []Assertion{
			{IP: "198.41.192.7", DestinationPort: 7844, RuleSet: "direct", Exclusive: true},
			{IP: "2606:4700:a0::1", DestinationPort: 7844, RuleSet: "direct", Exclusive: true},
			{Domain: "region1.v2.argotunnel.com", DestinationPort: 7844, RuleSet: "direct"},
			{IP: "198.41.192.7", DestinationPort: 443, RuleSet: "global"},
			{Domain: "api.cloudflare.com", DestinationPort: 443, RuleSet: "global"},
			{Domain: "api.cloudflare.com", RuleSet: "global"},
		},
	}
	built := Result{RuleSets: []RuleSet{
		{Name: "direct", Rules: rules.Collection{Classical: rules.ClassicalSet{port}}},
		{Name: "global", Rules: rules.Collection{Domains: rules.DomainSet{globalDomain}}},
	}}
	if err := verifyAssertions(cfg, built); err != nil {
		t.Fatal(err)
	}
	built.RuleSets[0].Rules = rules.Collection{}
	if err := verifyAssertions(cfg, built); err == nil {
		t.Fatal("missing destination port rule passed tunnel assertions")
	}
	built.RuleSets[0].Rules = rules.Collection{Classical: rules.ClassicalSet{port}}
	built.RuleSets[1].Rules = rules.Collection{Classical: rules.ClassicalSet{port}}
	if err := verifyAssertions(cfg, built); err == nil || !strings.Contains(err.Error(), "expected exclusive membership") {
		t.Fatalf("got %v, want overlapping destination port membership failure", err)
	}
}

func TestProjectTailnetAndDirectExclusions(t *testing.T) {
	t.Parallel()
	cfg, err := LoadConfig("../../config/rulesets.json")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(cfg.Priority, []string{"tailnet", "direct", "us", "global"}) {
		t.Fatalf("unexpected project priority: %v", cfg.Priority)
	}
	originalBase := cfg.BaseDir
	cfg.BaseDir = t.TempDir()
	write := func(path, content string) {
		t.Helper()
		full := filepath.Join(cfg.BaseDir, path)
		if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	privateDomains := []string{"ts.net", "server.example.ts.net", "forgejo.liblaf.me", "jellyfin.liblaf.me", "oplist.liblaf.me", "qbt.liblaf.me"}
	publicDomains := []string{"tailscale.com", "login.tailscale.com", "controlplane.tailscale.com", "derp1-all.tailscale.com", "tailscale.io", "controlplane.tailscale.io"}
	privateIPs := []string{"100.64.0.1", "100.127.255.254", "fd7a:115c:a1e0::1", "fd7a:115c:a1e0:ffff::1"}
	directDomains := slices.Concat(publicDomains, []string{"localhost", "router.local", "www.elsevier.com", "shop.pockyt.io"})
	var unrelated []Assertion
	for i := range cfg.RuleSets {
		set := &cfg.RuleSets[i]
		var fixtures []SourceConfig
		for j, src := range set.Sources {
			if src.Path != "" {
				payload, err := os.ReadFile(filepath.Join(originalBase, src.Path))
				if err != nil {
					t.Fatal(err)
				}
				write(src.Path, string(payload))
				fixtures = append(fixtures, src)
				continue
			}
			path := fmt.Sprintf("fixtures/%s-%d.domain.txt", set.Name, j)
			keep := fmt.Sprintf("keep.%s-%d.example", set.Name, j)
			payload := strings.Join(privateDomains, "\n") + "\n+.tailscale.com\n+.tailscale.io\n+.local\nlocalhost\n+.elsevier.com\n+.pockyt.io\n" + keep + "\n"
			if i > 1 {
				payload += "+.liblaf.me\nshared.us.example\n"
			}
			write(path, payload)
			fixtures = append(fixtures, SourceConfig{Name: src.Name, Kind: "domain-file", Path: path})
			unrelated = append(unrelated, Assertion{Domain: keep, RuleSet: set.Name, Exclusive: true})
		}
		if i > 0 {
			path := "fixtures/" + set.Name + ".ipcidr.txt"
			write(path, "100.0.0.0/8\nfd7a:115c::/32\n10.0.0.0/8\n1.0.1.0/24\n")
			fixtures = append(fixtures, SourceConfig{Name: "fixture-ips", Kind: "cidr-file", Path: path})
		}
		set.Sources = fixtures
	}
	cfg.Assertions = slices.DeleteFunc(cfg.Assertions, func(a Assertion) bool {
		return a.DestinationPort == 0 && a.RuleSet != "tailnet" && !slices.Contains(directDomains, a.Domain) && a.Domain != "claude.app" && a.Domain != "api.statsig.com" && a.IP != "160.79.104.1" && a.IP != "2607:6bc0::1"
	})
	cfg.Assertions = append(cfg.Assertions, unrelated...)
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	built, err := Run(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	tailnet := findRuleSet(t, built, "tailnet")
	direct := findRuleSet(t, built, "direct")
	portRule, err := rules.NewClassical("DST-PORT", "7844")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(direct.Classical, portRule) {
		t.Fatal("build lost maintained Cloudflare Tunnel destination port rule")
	}
	var tunnelSourceFound bool
	for _, set := range built.RuleSets {
		for _, src := range set.Sources {
			if src.Name == "Cloudflare:tunnel" {
				tunnelSourceFound = true
				if set.Name != "direct" || src.InputRules != 1 || src.Accepted != (Counts{Classical: 1}) || len(src.Ignored) != 0 {
					t.Errorf("unexpected tunnel source classification in %s: %+v", set.Name, src)
				}
			}
		}
		if set.Name != "direct" && slices.Contains(set.Rules.Classical, portRule) {
			t.Errorf("Cloudflare Tunnel port appeared in %s", set.Name)
		}
	}
	if !tunnelSourceFound {
		t.Fatal("build did not report maintained tunnel source")
	}
	wantDomains := []string{"+.ts.net", "forgejo.liblaf.me", "jellyfin.liblaf.me", "oplist.liblaf.me", "qbt.liblaf.me"}
	if !slices.Equal(tailnet.Domains, wantDomains) || len(tailnet.Classical) != 0 {
		t.Errorf("tailnet domain rules: got %v and classical %v, want %v", tailnet.Domains, tailnet.Classical, wantDomains)
	}
	wantPrefixes := rules.PrefixSet{netip.MustParsePrefix("100.64.0.0/10"), netip.MustParsePrefix("fd7a:115c:a1e0::/48")}
	if !slices.Equal(tailnet.Prefixes, wantPrefixes) {
		t.Errorf("tailnet prefixes: got %v, want %v", tailnet.Prefixes, wantPrefixes)
	}
	for _, domain := range publicDomains {
		if tailnet.MatchesDomain(domain) {
			t.Errorf("tailnet matches public %s", domain)
		}
	}
	for _, domain := range privateDomains[2:] {
		if tailnet.MatchesDomain("child." + domain) {
			t.Errorf("exact custom domain matches child of %s", domain)
		}
		if !direct.MatchesDomain(domain) || !direct.MatchesDomain("child."+domain) {
			t.Errorf("direct lost the broader suffix covering %s", domain)
		}
	}
	for _, domain := range directDomains {
		if !direct.MatchesDomain(domain) {
			t.Errorf("direct lacks merged domain %s", domain)
		}
	}
	for _, set := range built.RuleSets[1:] {
		for _, domain := range privateDomains[:2] {
			if set.Rules.MatchesDomain(domain) {
				t.Errorf("%s retains private domain %s", set.Name, domain)
			}
		}
		for _, ip := range privateIPs {
			if set.Rules.Prefixes.Contains(netip.MustParseAddr(ip)) {
				t.Errorf("%s retains private IP %s", set.Name, ip)
			}
		}
	}
	for _, set := range built.RuleSets[2:] {
		for _, domain := range slices.Concat(privateDomains[2:], directDomains) {
			if set.Rules.MatchesDomain(domain) {
				t.Errorf("%s retains excluded direct domain %s", set.Name, domain)
			}
		}
		for _, ip := range []string{"10.0.0.1", "1.0.1.1", "100.0.0.1", "fd7a:115c:a1df::1"} {
			if set.Rules.Prefixes.Contains(netip.MustParseAddr(ip)) {
				t.Errorf("%s retains direct IP %s", set.Name, ip)
			}
		}
		if !set.Rules.MatchesDomain("shared.us.example") {
			t.Errorf("%s lost the allowed US/global overlap", set.Name)
		}
	}
	for _, ip := range []string{"10.0.0.1", "1.0.1.1", "100.0.0.1", "100.128.0.1", "fd7a:115c:a1df::1", "fd7a:115c:a1e1::1"} {
		if !direct.Prefixes.Contains(netip.MustParseAddr(ip)) {
			t.Errorf("direct lost unrelated IP %s", ip)
		}
	}
	cfg.Assertions = nil
	for i := range cfg.RuleSets {
		cfg.RuleSets[i].Exclude = nil
	}
	unexcluded, err := Run(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, set := range unexcluded.RuleSets[1:] {
		for _, domain := range slices.Concat(privateDomains, directDomains) {
			if !set.Rules.MatchesDomain(domain) {
				t.Errorf("control %s lacks fixture domain %s", set.Name, domain)
			}
		}
		for _, ip := range privateIPs {
			if !set.Rules.Prefixes.Contains(netip.MustParseAddr(ip)) {
				t.Errorf("control %s lacks fixture IP %s", set.Name, ip)
			}
		}
	}
}

func findRuleSet(t *testing.T, built Result, name string) rules.Collection {
	t.Helper()
	for _, set := range built.RuleSets {
		if set.Name == name {
			return set.Rules
		}
	}
	t.Fatalf("ruleset %q not found in build result", name)
	return rules.Collection{}
}
