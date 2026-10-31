package rules

import (
	"net/netip"
	"testing"
)

func TestDestinationPortRulesMatchPortsWithoutMatchingDomains(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		value   string
		matches []uint16
		misses  []uint16
	}{
		{"7844", []uint16{7844}, []uint16{443, 7843, 7845}},
		{"80/443/7844-7846", []uint16{80, 443, 7844, 7845, 7846}, []uint16{0, 81, 7843, 7847, 65535}},
		{"7846-7844", []uint16{7844, 7845, 7846}, []uint16{7843, 7847}},
		{"0-65535", []uint16{0, 1, 7844, 65535}, nil},
	} {
		t.Run(tc.value, func(t *testing.T) {
			rule, err := NewClassicalRule("DST-PORT", tc.value)
			if err != nil {
				t.Fatal(err)
			}
			for _, port := range tc.matches {
				if !rule.MatchesDestinationPort(port) {
					t.Errorf("rule %s does not match %d", rule.Text(), port)
				}
			}
			for _, port := range tc.misses {
				if rule.MatchesDestinationPort(port) {
					t.Errorf("rule %s matches %d", rule.Text(), port)
				}
			}
			for _, domain := range []string{"", "7844", "api.cloudflare.com"} {
				if rule.Matches(domain) {
					t.Errorf("port rule matched domain %q", domain)
				}
			}
		})
	}
}

func TestDestinationPortRulesRejectInvalidExpressions(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"", "*", "https", "-1", "65536", "80-65536", "7844-", "1-2-3", "7844//443", "/7844", "7844/", "7844\n443"} {
		t.Run(value, func(t *testing.T) {
			if _, err := NewClassicalRule("DST-PORT", value); err == nil {
				t.Fatalf("accepted invalid destination port expression %q", value)
			}
		})
	}
}

func TestMixedClassicalRulesKeepDomainAndPortMatchingIndependent(t *testing.T) {
	t.Parallel()
	port, err := NewClassicalRule("DST-PORT", "7844")
	if err != nil {
		t.Fatal(err)
	}
	domain, err := NewClassicalRule("DOMAIN-KEYWORD", "openai")
	if err != nil {
		t.Fatal(err)
	}
	set := ClassicalSet{port, domain}
	if !set.Matches("api.openai.com") || set.Matches("api.cloudflare.com") {
		t.Fatal("mixed classical rules changed domain matching")
	}
	if !set.MatchesDestinationPort(7844) || set.MatchesDestinationPort(443) || domain.MatchesDestinationPort(7844) {
		t.Fatal("mixed classical rules changed destination port matching")
	}
}

func TestCollectionMatchesDestinationPortBesideDomainAndIPRules(t *testing.T) {
	t.Parallel()
	port, err := NewClassicalRule("DST-PORT", "0/7844")
	if err != nil {
		t.Fatal(err)
	}
	domain, err := NewDomainRule(DomainSuffix, "example.com")
	if err != nil {
		t.Fatal(err)
	}
	collection := Collection{
		Domains:   DomainSet{domain},
		Prefixes:  PrefixSet{netip.MustParsePrefix("192.0.2.0/24")},
		Classical: ClassicalSet{port},
	}
	for _, tc := range []struct {
		name   string
		domain string
		ip     netip.Addr
		port   uint16
		want   bool
	}{
		{"port only", "unrelated.example", netip.MustParseAddr("198.51.100.1"), 7844, true},
		{"domain only", "www.example.com", netip.Addr{}, 443, true},
		{"IP only", "", netip.MustParseAddr("192.0.2.1"), 443, true},
		{"no match", "unrelated.example", netip.MustParseAddr("198.51.100.1"), 443, false},
		{"unspecified port", "unrelated.example", netip.Addr{}, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := collection.Matches(tc.domain, tc.ip, tc.port); got != tc.want {
				t.Fatalf("Matches(%q, %s, %d) = %t, want %t", tc.domain, tc.ip, tc.port, got, tc.want)
			}
		})
	}
}

func TestContextualClassicalRulesPreserveFieldsWithoutDomainMatches(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ kind, value string }{
		{"PROCESS-NAME", "cloudflared"},
		{"IP-ASN", "13335,no-resolve,src"},
		{"AND", "((NETWORK,UDP),(DST-PORT,7844))"},
		{"SRC-IP-CIDR", "192.0.2.0/24"},
		{"DST-PORT", "7844,no-resolve"},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			rule, err := NewClassicalRule(tc.kind, tc.value)
			if err != nil {
				t.Fatal(err)
			}
			if rule.Kind != tc.kind || rule.Value != tc.value || rule.Text() != tc.kind+","+tc.value {
				t.Fatalf("classical payload changed: %+v", rule)
			}
			if rule.Matches("cloudflared.example") || rule.MatchesDestinationPort(7844) {
				t.Fatal("contextual rule matched without the required connection metadata")
			}
		})
	}
	if _, err := NewClassicalRule("PROCESS-NMAE", "cloudflared"); err == nil {
		t.Fatal("unknown rule type accepted")
	}
}
