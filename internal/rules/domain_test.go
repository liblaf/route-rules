package rules

import (
	"fmt"
	"slices"
	"testing"
)

func TestDomainSetOptimized(t *testing.T) {
	t.Parallel()
	set := DomainSet{
		mustParseDomain(t, "+.example.com"),
		mustParseDomain(t, "+.api.example.com"),
		mustParseDomain(t, "www.example.com"),
		mustParseDomain(t, "only.example.net"),
		mustParseDomain(t, "*.example.net"),
		mustParseDomain(t, "*.covered.example.com"),
	}
	assertStrings(t, set.Optimized(), []string{"*.example.net", "+.example.com"})
}

func TestDomainSetExcludingDropsOnlyFullyCoveredRules(t *testing.T) {
	t.Parallel()
	earlier := DomainSet{
		mustParseDomain(t, "api.example.com"),
		mustParseDomain(t, "api.example.com"),
		mustParseDomain(t, "+.sub.covered.test"),
		mustParseDomain(t, "+.covered.test"),
		mustParseDomain(t, "*.covered.test"),
	}
	later := DomainSet{
		mustParseDomain(t, "api.example.com"),
		mustParseDomain(t, "+.example.com"),
		mustParseDomain(t, "a.covered.test"),
		mustParseDomain(t, "+.deep.sub.covered.test"),
		mustParseDomain(t, "*.deep.covered.test"),
	}
	assertStrings(t, later.Excluding(earlier), []string{"+.example.com"})
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
		mustParseDomain(t, "api.example.com"),
		mustParseDomain(t, "*.internal.example"),
		mustParseDomain(t, "+.direct.test"),
	}
	later := DomainSet{
		mustParseDomain(t, "+.example.com"),
		mustParseDomain(t, "api.example.com"),
		mustParseDomain(t, "+.sub.example.com"),
		mustParseDomain(t, "*.internal.example"),
		mustParseDomain(t, "other.test"),
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

func mustParseDomain(t *testing.T, line string) string {
	t.Helper()
	rule, err := ParseDomain(line)
	if err != nil {
		t.Fatal(err)
	}
	return rule
}

func assertStrings(t *testing.T, actual, expected []string) {
	t.Helper()
	if !slices.Equal(actual, expected) {
		t.Fatalf("got %v, expected %v", actual, expected)
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
