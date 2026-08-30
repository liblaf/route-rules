package build

import (
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
