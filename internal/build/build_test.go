package build

import (
	"context"
	"net/netip"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/liblaf/route-rules/internal/config"
	"github.com/liblaf/route-rules/internal/rules"
)

func TestVerifyAssertionsUsesPriorityAndFallback(t *testing.T) {
	t.Parallel()
	lanRule, err := rules.NewDomainRule(rules.DomainSuffix, "local")
	if err != nil {
		t.Fatal(err)
	}
	usRule, err := rules.NewDomainRule(rules.DomainSuffix, "openai.com")
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		Priority: []string{"lan", "us", "global"},
		Fallback: "global",
		Assertions: []config.Assertion{
			{Domain: "router.local", RuleSet: "lan"},
			{Domain: "api.openai.com", RuleSet: "us"},
			{Domain: "unknown.example", RuleSet: "global"},
		},
	}
	built := Result{byName: map[string]rules.Collection{
		"lan":    {Domains: rules.DomainSet{lanRule}},
		"us":     {Domains: rules.DomainSet{usRule}},
		"global": {},
	}}
	if err := verifyAssertions(cfg, built); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyAssertionsSupportsExclusiveIPMembership(t *testing.T) {
	t.Parallel()
	tailnetPrefix := netip.MustParsePrefix("100.64.0.0/10")
	cfg := config.Config{
		Priority: []string{"tailnet", "lan"},
		Fallback: "lan",
		Assertions: []config.Assertion{
			{IP: "100.64.0.1", RuleSet: "tailnet", Exclusive: true},
		},
	}
	built := Result{byName: map[string]rules.Collection{
		"tailnet": {Prefixes: rules.PrefixSet{tailnetPrefix}},
		"lan":     {},
	}}
	if err := verifyAssertions(cfg, built); err != nil {
		t.Fatal(err)
	}

	built.byName["lan"] = rules.Collection{Prefixes: rules.PrefixSet{tailnetPrefix}}
	err := verifyAssertions(cfg, built)
	if err == nil || !strings.Contains(err.Error(), "expected exclusive membership") {
		t.Fatalf("got error %v, expected exclusive-membership failure", err)
	}
}

func TestProjectTailnetAndTailscaleExclusions(t *testing.T) {
	t.Parallel()
	cfg, err := config.Load("../../config/rulesets.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Priority) < 2 || cfg.Priority[0] != "tailnet" || cfg.Priority[1] != "tailscale" {
		t.Fatal("project config must define tailnet then tailscale")
	}
	originalBase := cfg.BaseDir
	cfg.BaseDir = t.TempDir()
	// Keep the maintained split sources; use fixtures for all lower-set inputs.
	// Each lower set gets unique unrelated entries so earlier exclusions cannot hide loss.
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
	for i := range cfg.RuleSets {
		set := &cfg.RuleSets[i]
		var local []config.Source
		for _, src := range set.Sources {
			if i >= 2 || (src.Kind != "domain-file" && src.Kind != "cidr-file") {
				continue
			}
			payload, err := os.ReadFile(filepath.Join(originalBase, src.Path))
			if err != nil {
				t.Fatal(err)
			}
			write(src.Path, string(payload))
			local = append(local, src)
		}
		if i >= 2 {
			domainPath := "fixtures/" + set.Name + ".domain.txt"
			ipPath := "fixtures/" + set.Name + ".ipcidr.txt"
			write(domainPath, "+.tailscale.com\n+.tailscale.io\n"+strings.Join(append(append([]string{}, privateDomains...), publicDomains...), "\n")+"\nkeep."+set.Name+".example\n")
			// Supernets force subtraction of the private ranges while retaining neighbors.
			write(ipPath, "100.0.0.0/8\nfd7a:115c::/32\n")
			local = append(local, config.Source{Name: "fixture-domains", Kind: "domain-file", Path: domainPath}, config.Source{Name: "fixture-ips", Kind: "cidr-file", Path: ipPath})
		}
		set.Sources = local
	}
	// Unrelated production assertions depend on remote data. Keep split assertions.
	cfg.Assertions = slices.DeleteFunc(cfg.Assertions, func(a config.Assertion) bool { return a.RuleSet != "tailnet" && a.RuleSet != "tailscale" })
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	built, err := Run(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string][]string{
		"tailnet":   {"+.ts.net", "forgejo.liblaf.me", "jellyfin.liblaf.me", "oplist.liblaf.me", "qbt.liblaf.me"},
		"tailscale": {"+.tailscale.com", "+.tailscale.io"},
	} {
		got := built.byName[name]
		var texts []string
		for _, domain := range got.Domains {
			texts = append(texts, domain.Text())
		}
		slices.Sort(texts)
		slices.Sort(want)
		if !slices.Equal(texts, want) {
			t.Errorf("%s domains: got %v, want %v", name, texts, want)
		}
		if len(got.Classical) != 0 {
			t.Errorf("%s has classical rules: %v", name, got.Classical)
		}
	}
	tailnet := built.byName["tailnet"]
	wantPrefixes := rules.PrefixSet{netip.MustParsePrefix("100.64.0.0/10"), netip.MustParsePrefix("fd7a:115c:a1e0::/48")}
	if !slices.Equal(tailnet.Prefixes, wantPrefixes) {
		t.Errorf("tailnet prefixes: got %v, want %v", tailnet.Prefixes, wantPrefixes)
	}
	if len(built.byName["tailscale"].Prefixes) != 0 {
		t.Error("public tailscale contains IP rules")
	}
	for _, domain := range publicDomains {
		if tailnet.MatchesDomain(domain) {
			t.Errorf("tailnet matches public %s", domain)
		}
	}
	for _, domain := range privateDomains {
		if built.byName["tailscale"].MatchesDomain(domain) {
			t.Errorf("tailscale matches private %s", domain)
		}
	}
	for _, domain := range privateDomains[2:] {
		if tailnet.MatchesDomain("child." + domain) {
			t.Errorf("exact custom domain matches child of %s", domain)
		}
	}
	for _, set := range built.RuleSets[2:] {
		for _, domain := range append(append([]string{}, privateDomains...), publicDomains...) {
			if set.Rules.MatchesDomain(domain) {
				t.Errorf("%s retains excluded domain %s", set.Name, domain)
			}
		}
		for _, ip := range privateIPs {
			if set.Rules.Prefixes.Contains(netip.MustParseAddr(ip)) {
				t.Errorf("%s retains private IP %s", set.Name, ip)
			}
		}
		if !set.Rules.MatchesDomain("keep." + set.Name + ".example") {
			t.Errorf("%s lost unrelated domain", set.Name)
		}
	}
	// LAN sees the supernets first; subtraction must preserve their unrelated portions.
	for _, ip := range []string{"100.0.0.1", "100.128.0.1", "fd7a:115c:a1df::1", "fd7a:115c:a1e1::1"} {
		if !built.byName["lan"].Prefixes.Contains(netip.MustParseAddr(ip)) {
			t.Errorf("subtraction lost unrelated IP %s", ip)
		}
	}
	// Match the maintained CN override's partial overlap too. Exact custom hosts
	// cannot be subtracted from +.liblaf.me, so Run must verify Tailnet priority.
	cnPath := "fixtures/cn.domain.txt"
	cnPayload, err := os.ReadFile(filepath.Join(cfg.BaseDir, cnPath))
	if err != nil {
		t.Fatal(err)
	}
	write(cnPath, string(cnPayload)+"+.liblaf.me\n")
	overlapping, err := Run(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, domain := range privateDomains[2:] {
		if !overlapping.byName["cn"].MatchesDomain(domain) || !overlapping.byName["cn"].MatchesDomain("child."+domain) {
			t.Errorf("CN lost the broader suffix covering %s", domain)
		}
	}
	write(cnPath, string(cnPayload))
	// A control build without exclusions demonstrates actual overlap in every input,
	// independent of first-match routing or the production assertion list.
	cfg.Assertions = nil
	for i := range cfg.RuleSets {
		cfg.RuleSets[i].Exclude = nil
	}
	unexcluded, err := Run(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, set := range unexcluded.RuleSets[2:] {
		for _, domain := range append(append([]string{}, privateDomains...), publicDomains...) {
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
