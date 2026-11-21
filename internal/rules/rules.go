package rules

import "net/netip"

type Collection struct {
	Domains   DomainSet
	Classical ClassicalSet
	Prefixes  PrefixSet
}

func (c Collection) Optimized() Collection {
	return Collection{
		Domains:   c.Domains.Optimized(),
		Classical: c.Classical.Optimized(),
		Prefixes:  c.Prefixes.Optimized(),
	}
}

func (c Collection) Excluding(earlier Collection) Collection {
	return Collection{
		Domains:   c.Domains.Excluding(earlier.Domains),
		Classical: c.Classical.Excluding(earlier.Classical),
		Prefixes:  c.Prefixes.Excluding(earlier.Prefixes),
	}
}

func (c Collection) MatchesDomain(domain string) bool {
	return c.Domains.Matches(domain) || c.Classical.Matches(domain)
}

// Matches evaluates the available domain, destination IP, and destination port.
// A zero destination port means that no port metadata was supplied.
func (c Collection) Matches(domain string, addr netip.Addr, dstPort uint16) bool {
	return domain != "" && c.MatchesDomain(domain) ||
		addr.IsValid() && c.Prefixes.Contains(addr) ||
		dstPort != 0 && c.Classical.MatchesDestinationPort(dstPort)
}

func (c *Collection) Append(other Collection) {
	c.Domains = append(c.Domains, other.Domains...)
	c.Classical = append(c.Classical, other.Classical...)
	c.Prefixes = append(c.Prefixes, other.Prefixes...)
}
