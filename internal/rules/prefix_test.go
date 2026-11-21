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
	assertStrings(t, set.Optimized().Texts(), []string{"192.0.2.0/24", "2001:db8::/32"})
}

func TestPrefixSetOptimizedEdgeCases(t *testing.T) {
	t.Parallel()
	assertStrings(t, PrefixSet{}.Optimized().Texts(), nil)

	cascadeToZero := PrefixSet{
		netip.MustParsePrefix("0.0.0.1/2"),
		netip.MustParsePrefix("64.0.0.2/2"),
		netip.MustParsePrefix("128.0.0.3/1"),
	}
	assertStrings(t, cascadeToZero.Optimized().Texts(), []string{"0.0.0.0/0"})

	mixedFamilies := PrefixSet{
		netip.MustParsePrefix("::ffff:192.0.2.0/120"),
		netip.MustParsePrefix("192.0.2.0/24"),
		netip.MustParsePrefix("255.255.255.255/32"),
		netip.MustParsePrefix("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff/128"),
	}
	assertStrings(t, mixedFamilies.Optimized().Texts(), []string{
		"192.0.2.0/24",
		"255.255.255.255/32",
		"::ffff:192.0.2.0/120",
		"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff/128",
	})
}

func TestPrefixSetExcludingSplitsCIDR(t *testing.T) {
	t.Parallel()
	set := PrefixSet{netip.MustParsePrefix("192.0.2.0/24")}
	excluded := PrefixSet{netip.MustParsePrefix("192.0.2.0/26")}
	assertStrings(t, set.Excluding(excluded).Texts(), []string{"192.0.2.64/26", "192.0.2.128/25"})
}

func TestPrefixSetExcludingEdgeCases(t *testing.T) {
	t.Parallel()
	base := PrefixSet{
		netip.MustParsePrefix("192.0.2.0/24"),
		netip.MustParsePrefix("::ffff:192.0.2.0/120"),
	}
	// IPv4 cut must not remove IPv4-mapped IPv6 range.
	assertStrings(t, base.Excluding(PrefixSet{netip.MustParsePrefix("192.0.2.0/24")}).Texts(), []string{
		"::ffff:192.0.2.0/120",
	})
	// Full removal across both families leaves an empty set.
	assertStrings(t, base.Excluding(PrefixSet{
		netip.MustParsePrefix("0.0.0.0/0"),
		netip.MustParsePrefix("::/0"),
	}).Texts(), nil)
	// Subtracting /1 from /0 leaves the complementary /1.
	assertStrings(t, PrefixSet{netip.MustParsePrefix("0.0.0.0/0")}.Excluding(PrefixSet{
		netip.MustParsePrefix("128.0.0.0/1"),
	}).Texts(), []string{"0.0.0.0/1"})
	// Empty base or empty cuts.
	assertStrings(t, PrefixSet{}.Excluding(base).Texts(), nil)
	assertStrings(t, base.Excluding(nil).Texts(), []string{"192.0.2.0/24", "::ffff:192.0.2.0/120"})
}

func TestPrefixOptimizationAndExclusionPreserveMembership(t *testing.T) {
	t.Parallel()
	random := rand.New(rand.NewSource(1))
	for iteration := range 100 {
		original := append(randomIPv4Prefixes(random, 24), randomIPv6Prefixes(random, 24)...)
		cuts := append(randomIPv4Prefixes(random, 12), randomIPv6Prefixes(random, 12)...)
		optimized := original.Optimized()
		excluded := original.Excluding(cuts)
		for host := range 256 {
			for _, addr := range []netip.Addr{
				netip.AddrFrom4([4]byte{192, 0, 2, byte(host)}),
				testIPv6(byte(host)),
			} {
				if original.Contains(addr) != optimized.Contains(addr) {
					t.Fatalf("iteration %d: optimization changed membership for %s", iteration, addr)
				}
				expected := original.Contains(addr) && !cuts.Contains(addr)
				if excluded.Contains(addr) != expected {
					t.Fatalf("iteration %d: exclusion membership for %s is %t, expected %t", iteration, addr, excluded.Contains(addr), expected)
				}
			}
		}
	}
}

func testIPv6(host byte) netip.Addr {
	var raw [16]byte
	raw[0], raw[1], raw[2], raw[3] = 0x20, 0x01, 0x0d, 0xb8
	raw[15] = host
	return netip.AddrFrom16(raw)
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

func randomIPv6Prefixes(random *rand.Rand, count int) PrefixSet {
	result := make(PrefixSet, 0, count)
	for range count {
		bits := 120 + random.Intn(9)
		result = append(result, netip.PrefixFrom(testIPv6(byte(random.Intn(256))), bits).Masked())
	}
	return result
}
