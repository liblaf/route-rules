package rules

import "net/netip"

type Collection struct {
	Domains   DomainSet
	Classical ClassicalSet
	Prefixes  PrefixSet
}

func (collection Collection) Optimized() Collection {
	collection.Domains = collection.Domains.Optimized()
	collection.Classical = collection.Classical.Optimized()
	collection.Prefixes = collection.Prefixes.Optimized()
	return collection
}

func (collection Collection) Excluding(earlier Collection) Collection {
	collection.Domains = collection.Domains.Excluding(earlier.Domains)
	collection.Classical = collection.Classical.Excluding(earlier.Classical)
	collection.Prefixes = collection.Prefixes.Excluding(earlier.Prefixes)
	return collection.Optimized()
}

func (collection Collection) MatchesDomain(domain string) bool {
	return collection.Domains.Matches(domain) || collection.Classical.Matches(domain)
}

// Matches evaluates the available domain, destination IP, and destination port.
// A zero destination port means that no port metadata was supplied.
func (collection Collection) Matches(domain string, address netip.Addr, destinationPort uint16) bool {
	return domain != "" && collection.MatchesDomain(domain) ||
		address.IsValid() && collection.Prefixes.Contains(address) ||
		destinationPort != 0 && collection.Classical.MatchesDestinationPort(destinationPort)
}

func (collection *Collection) Append(other Collection) {
	collection.Domains = append(collection.Domains, other.Domains...)
	collection.Classical = append(collection.Classical, other.Classical...)
	collection.Prefixes = append(collection.Prefixes, other.Prefixes...)
}
