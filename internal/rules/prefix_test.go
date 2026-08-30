package rules

import (
	"math/rand"
	"net/netip"
	"testing"
)

func TestPrefixSetOptimizedMergesSiblings(t *testing.T) {
	t.Parallel()
	set := PrefixSet{
		netip.MustParsePrefix("192.0.2.0/25"),
		netip.MustParsePrefix("192.0.2.128/25"),
		netip.MustParsePrefix("2001:db8::/33"),
		netip.MustParsePrefix("2001:db8:8000::/33"),
	}
	actual := prefixRuleTexts(set.Optimized())
	want := []string{"192.0.2.0/24", "2001:db8::/32"}
	assertStrings(t, actual, want)
}

func TestPrefixSetExcludingSplitsCIDR(t *testing.T) {
	t.Parallel()
	set := PrefixSet{netip.MustParsePrefix("192.0.2.0/24")}
	excluded := PrefixSet{netip.MustParsePrefix("192.0.2.0/26")}
	actual := prefixRuleTexts(set.Excluding(excluded))
	want := []string{"192.0.2.64/26", "192.0.2.128/25"}
	assertStrings(t, actual, want)
}

func TestPrefixOptimizationAndExclusionPreserveMembership(t *testing.T) {
	t.Parallel()
	random := rand.New(rand.NewSource(1))
	for iteration := 0; iteration < 100; iteration++ {
		original := randomIPv4Prefixes(random, 24)
		cuts := randomIPv4Prefixes(random, 12)
		optimized := original.Optimized()
		excluded := original.Excluding(cuts)
		for host := 0; host < 256; host++ {
			address := netip.AddrFrom4([4]byte{192, 0, 2, byte(host)})
			if original.Contains(address) != optimized.Contains(address) {
				t.Fatalf("iteration %d: optimization changed membership for %s", iteration, address)
			}
			expected := original.Contains(address) && !cuts.Contains(address)
			if excluded.Contains(address) != expected {
				t.Fatalf("iteration %d: exclusion membership for %s is %t, expected %t", iteration, address, excluded.Contains(address), expected)
			}
		}
	}
}

func prefixRuleTexts(set PrefixSet) []string {
	result := make([]string, 0, len(set))
	for _, prefix := range set {
		result = append(result, prefix.String())
	}
	return result
}

func randomIPv4Prefixes(random *rand.Rand, count int) PrefixSet {
	result := make(PrefixSet, 0, count)
	for range count {
		address := netip.AddrFrom4([4]byte{192, 0, 2, byte(random.Intn(256))})
		bits := 24 + random.Intn(9)
		result = append(result, netip.PrefixFrom(address, bits).Masked())
	}
	return result
}
