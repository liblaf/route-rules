package build

import (
	"net/netip"
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
		Priority: []string{"lan", "us", "proxy"},
		Fallback: "proxy",
		Assertions: []config.Assertion{
			{Domain: "router.local", RuleSet: "lan"},
			{Domain: "api.openai.com", RuleSet: "us"},
			{Domain: "unknown.example", RuleSet: "proxy"},
		},
	}
	built := Result{byName: map[string]rules.Collection{
		"lan":   {Domains: rules.DomainSet{lanRule}},
		"us":    {Domains: rules.DomainSet{usRule}},
		"proxy": {},
	}}
	if err := verifyAssertions(cfg, built); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyAssertionsSupportsExclusiveIPMembership(t *testing.T) {
	t.Parallel()
	tailscalePrefix := netip.MustParsePrefix("100.64.0.0/10")
	cfg := config.Config{
		Priority: []string{"tailscale", "lan"},
		Fallback: "lan",
		Assertions: []config.Assertion{
			{IP: "100.64.0.1", RuleSet: "tailscale", Exclusive: true},
		},
	}
	built := Result{byName: map[string]rules.Collection{
		"tailscale": {Prefixes: rules.PrefixSet{tailscalePrefix}},
		"lan":       {},
	}}
	if err := verifyAssertions(cfg, built); err != nil {
		t.Fatal(err)
	}

	built.byName["lan"] = rules.Collection{Prefixes: rules.PrefixSet{tailscalePrefix}}
	err := verifyAssertions(cfg, built)
	if err == nil || !strings.Contains(err.Error(), "expected exclusive membership") {
		t.Fatalf("got error %v, expected exclusive-membership failure", err)
	}
}
