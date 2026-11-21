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
			set := ClassicalSet{mustClassical(t, "DST-PORT", tc.value)}
			for _, port := range tc.matches {
				if !set.MatchesDestinationPort(port) {
					t.Errorf("rule %s does not match %d", set[0], port)
				}
			}
			for _, port := range tc.misses {
				if set.MatchesDestinationPort(port) {
					t.Errorf("rule %s matches %d", set[0], port)
				}
			}
			for _, domain := range []string{"", "7844", "api.cloudflare.com"} {
				if set.Matches(domain) {
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
			if _, err := NewClassical("DST-PORT", value); err == nil {
				t.Fatalf("accepted invalid destination port expression %q", value)
			}
		})
	}
}

func TestMixedClassicalRulesKeepDomainAndPortMatchingIndependent(t *testing.T) {
	t.Parallel()
	port := mustClassical(t, "DST-PORT", "7844")
	domain := mustClassical(t, "DOMAIN-KEYWORD", "openai")
	set := ClassicalSet{port, domain}
	if !set.Matches("api.openai.com") || set.Matches("api.cloudflare.com") {
		t.Fatal("mixed classical rules changed domain matching")
	}
	if !set.MatchesDestinationPort(7844) || set.MatchesDestinationPort(443) || (ClassicalSet{domain}).MatchesDestinationPort(7844) {
		t.Fatal("mixed classical rules changed destination port matching")
	}
}

func TestCollectionMatchesDestinationPortBesideDomainAndIPRules(t *testing.T) {
	t.Parallel()
	collection := Collection{
		Domains:   DomainSet{mustParseDomain(t, "+.example.com")},
		Prefixes:  PrefixSet{netip.MustParsePrefix("192.0.2.0/24")},
		Classical: ClassicalSet{mustClassical(t, "DST-PORT", "0/7844")},
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

func TestClassicalGlobAndRegexMatching(t *testing.T) {
	t.Parallel()
	collection := Collection{Classical: ClassicalSet{
		mustClassical(t, "DOMAIN-KEYWORD", "OpenAI"),
		mustClassical(t, "DOMAIN-WILDCARD", "*.qhimgs?.com"),
		mustClassical(t, "DOMAIN-REGEX", `^api[0-9]{1,2}\.example\.com$`),
		mustClassical(t, "DST-PORT", "7844,no-resolve"),
	}}
	for _, tc := range []struct {
		domain string
		port   uint16
		want   bool
	}{
		{"api.openai.com", 0, true},
		{"cdn.qhimgs7.com", 0, true},
		{"cdn.qhimgs12.com", 0, false},
		{"cdn.qhimgs\n.com", 0, false},
		{"api42.example.com", 0, true},
		{"api123.example.com", 0, false},
		{"unrelated.example", 7844, false},
	} {
		if got := collection.Matches(tc.domain, netip.Addr{}, tc.port); got != tc.want {
			t.Errorf("Collection.Matches(%q, %d) = %t, want %t", tc.domain, tc.port, got, tc.want)
		}
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
			rule := mustClassical(t, tc.kind, tc.value)
			if rule != tc.kind+","+tc.value {
				t.Fatalf("classical payload changed: %q", rule)
			}
			set := ClassicalSet{rule}
			if set.Matches("cloudflared.example") || set.MatchesDestinationPort(7844) {
				t.Fatal("contextual rule matched without the required connection metadata")
			}
		})
	}
	if _, err := NewClassical("PROCESS-NMAE", "cloudflared"); err == nil {
		t.Fatal("unknown rule type accepted")
	}
}

func TestCollectionExcludingAndMatchesPreserveInvariants(t *testing.T) {
	t.Parallel()
	earlier := Collection{
		Domains:   DomainSet{mustParseDomain(t, "+.covered.test")},
		Classical: ClassicalSet{mustClassical(t, "DST-PORT", "7844")},
		Prefixes:  PrefixSet{netip.MustParsePrefix("192.0.2.0/26")},
	}.Optimized()
	later := Collection{
		Domains: DomainSet{
			mustParseDomain(t, "a.covered.test"),
			mustParseDomain(t, "+.keep.test"),
		},
		Classical: ClassicalSet{
			mustClassical(t, "DST-PORT", "7844"),
			mustClassical(t, "DOMAIN-KEYWORD", "openai"),
		},
		Prefixes: PrefixSet{netip.MustParsePrefix("192.0.2.0/24")},
	}.Excluding(earlier)

	assertStrings(t, later.Domains, []string{"+.keep.test"})
	assertStrings(t, later.Classical, []string{"DOMAIN-KEYWORD,openai"})
	assertStrings(t, later.Prefixes.Texts(), []string{"192.0.2.64/26", "192.0.2.128/25"})

	if later.Matches("a.covered.test", netip.Addr{}, 0) {
		t.Fatal("matcher retained excluded domain")
	}
	if !later.Matches("sub.keep.test", netip.Addr{}, 0) || !later.Matches("api.openai.com", netip.Addr{}, 0) {
		t.Fatal("matcher lost retained domain or keyword rule")
	}
	if later.Matches("", netip.Addr{}, 7844) {
		t.Fatal("matcher retained excluded port")
	}
	if later.Matches("", netip.MustParseAddr("192.0.2.1"), 0) || !later.Matches("", netip.MustParseAddr("192.0.2.100"), 0) {
		t.Fatal("matcher returned wrong IP prefix membership")
	}
}

func mustClassical(t *testing.T, kind, value string) string {
	t.Helper()
	rule, err := NewClassical(kind, value)
	if err != nil {
		t.Fatal(err)
	}
	return rule
}
