package rules

import (
	"fmt"
	"testing"
)

func TestDomainSetOptimized(t *testing.T) {
	t.Parallel()
	set := DomainSet{
		mustDomain(t, DomainSuffix, "example.com"),
		mustDomain(t, DomainSuffix, "api.example.com"),
		mustDomain(t, DomainExact, "www.example.com"),
		mustDomain(t, DomainExact, "only.example.net"),
		mustDomain(t, DomainWildcard, "*.example.net"),
		mustDomain(t, DomainWildcard, "*.covered.example.com"),
	}
	actual := domainRuleTexts(set.Optimized())
	want := []string{"*.example.net", "+.example.com"}
	assertStrings(t, actual, want)
}

func TestDomainSetExcludingDropsOnlyFullyCoveredRules(t *testing.T) {
	t.Parallel()
	earlier := DomainSet{mustDomain(t, DomainExact, "api.example.com")}
	later := DomainSet{
		mustDomain(t, DomainExact, "api.example.com"),
		mustDomain(t, DomainSuffix, "example.com"),
	}
	actual := domainRuleTexts(later.Excluding(earlier))
	want := []string{"+.example.com"}
	assertStrings(t, actual, want)
}

func TestPartialWildcardIsNotMRSCompatible(t *testing.T) {
	t.Parallel()
	if IsWholeLabelWildcard("*.qhimgs?.com") {
		t.Fatal("partial wildcard was incorrectly accepted as an MRS domain rule")
	}
	if !IsWholeLabelWildcard("images.*.example.com") {
		t.Fatal("whole-label wildcard was rejected")
	}
}

func TestDomainOptimizationAndExclusionPreserveFirstMatchSemantics(t *testing.T) {
	t.Parallel()
	earlier := DomainSet{
		mustDomain(t, DomainExact, "api.example.com"),
		mustDomain(t, DomainWildcard, "*.internal.example"),
		mustDomain(t, DomainSuffix, "direct.test"),
	}
	later := DomainSet{
		mustDomain(t, DomainSuffix, "example.com"),
		mustDomain(t, DomainExact, "api.example.com"),
		mustDomain(t, DomainSuffix, "sub.example.com"),
		mustDomain(t, DomainWildcard, "*.internal.example"),
		mustDomain(t, DomainExact, "other.test"),
	}
	optimizedEarlier := earlier.Optimized()
	optimizedLater := later.Excluding(optimizedEarlier)
	for _, domain := range sampledDomains() {
		before := classifyDomain(earlier, later, domain)
		after := classifyDomain(optimizedEarlier, optimizedLater, domain)
		if before != after {
			t.Fatalf("%s classified as %s before optimization and %s after", domain, before, after)
		}
	}
}

func mustDomain(t *testing.T, kind DomainKind, value string) DomainRule {
	t.Helper()
	rule, err := NewDomainRule(kind, value)
	if err != nil {
		t.Fatal(err)
	}
	return rule
}

func domainRuleTexts(set DomainSet) []string {
	result := make([]string, 0, len(set))
	for _, rule := range set {
		result = append(result, rule.Text())
	}
	return result
}

func assertStrings(t *testing.T, actual, expected []string) {
	t.Helper()
	if len(actual) != len(expected) {
		t.Fatalf("got %v, expected %v", actual, expected)
	}
	for index := range expected {
		if actual[index] != expected[index] {
			t.Fatalf("got %v, expected %v", actual, expected)
		}
	}
}

func classifyDomain(earlier, later DomainSet, domain string) string {
	if earlier.Matches(domain) {
		return "earlier"
	}
	if later.Matches(domain) {
		return "later"
	}
	return "fallback"
}

func sampledDomains() []string {
	labels := []string{"api", "sub", "www", "internal", "example", "com", "test", "direct", "other"}
	result := append([]string(nil), labels...)
	for _, first := range labels {
		for _, second := range labels {
			result = append(result, fmt.Sprintf("%s.%s", first, second))
			for _, third := range labels {
				result = append(result, fmt.Sprintf("%s.%s.%s", first, second, third))
			}
		}
	}
	return result
}
