package rules

import (
	"fmt"
	"net/netip"
	"sort"
)

type PrefixSet []netip.Prefix

func (set PrefixSet) Contains(address netip.Addr) bool {
	for _, prefix := range set {
		if prefix.Contains(address) {
			return true
		}
	}
	return false
}

func (set PrefixSet) Optimized() PrefixSet {
	unique := make(map[netip.Prefix]struct{}, len(set))
	for _, prefix := range set {
		if !prefix.IsValid() {
			panic("invalid IP prefix")
		}
		unique[prefix.Masked()] = struct{}{}
	}

	prefixes := prefixMapToSortedSlice(unique)
	unique = make(map[netip.Prefix]struct{}, len(prefixes))
	for _, prefix := range prefixes {
		if !coveredByPrefixMap(unique, prefix) {
			unique[prefix] = struct{}{}
		}
	}

	for _, addressBits := range []int{32, 128} {
		for bits := addressBits; bits > 0; bits-- {
			parents := make(map[netip.Prefix]uint8)
			for prefix := range unique {
				if prefix.Addr().BitLen() != addressBits || prefix.Bits() != bits {
					continue
				}
				parent := netip.PrefixFrom(prefix.Addr(), bits-1).Masked()
				if childIndex(prefix) == 0 {
					parents[parent] |= 1
				} else {
					parents[parent] |= 2
				}
			}
			for parent, children := range parents {
				if children != 3 {
					continue
				}
				left, right := prefixChildren(parent)
				delete(unique, left)
				delete(unique, right)
				unique[parent] = struct{}{}
			}
		}
	}
	return prefixMapToSortedSlice(unique)
}

func (set PrefixSet) Excluding(earlier PrefixSet) PrefixSet {
	work := append(PrefixSet(nil), set.Optimized()...)
	cuts := earlier.Optimized()
	for _, cut := range cuts {
		next := make(PrefixSet, 0, len(work))
		for _, prefix := range work {
			next = append(next, subtractPrefix(prefix, cut)...)
		}
		work = next
	}
	return work.Optimized()
}

func subtractPrefix(base, cut netip.Prefix) PrefixSet {
	base = base.Masked()
	cut = cut.Masked()
	if base.Addr().BitLen() != cut.Addr().BitLen() || !prefixesOverlap(base, cut) {
		return PrefixSet{base}
	}
	if prefixContains(cut, base) {
		return nil
	}
	if !prefixContains(base, cut) {
		return PrefixSet{base}
	}
	left, right := prefixChildren(base)
	result := subtractPrefix(left, cut)
	return append(result, subtractPrefix(right, cut)...)
}

func prefixChildren(prefix netip.Prefix) (netip.Prefix, netip.Prefix) {
	bits := prefix.Bits()
	if bits < 0 || bits >= prefix.Addr().BitLen() {
		panic(fmt.Sprintf("prefix %s cannot be split", prefix))
	}
	childBits := bits + 1
	left := netip.PrefixFrom(prefix.Addr(), childBits).Masked()
	rightAddress := setAddressBit(prefix.Addr(), bits)
	right := netip.PrefixFrom(rightAddress, childBits).Masked()
	return left, right
}

func setAddressBit(address netip.Addr, bit int) netip.Addr {
	if address.Is4() {
		bytes := address.As4()
		bytes[bit/8] |= 1 << (7 - bit%8)
		return netip.AddrFrom4(bytes)
	}
	bytes := address.As16()
	bytes[bit/8] |= 1 << (7 - bit%8)
	return netip.AddrFrom16(bytes)
}

func childIndex(prefix netip.Prefix) uint8 {
	bit := prefix.Bits() - 1
	address := prefix.Addr()
	if address.Is4() {
		bytes := address.As4()
		return (bytes[bit/8] >> (7 - bit%8)) & 1
	}
	bytes := address.As16()
	return (bytes[bit/8] >> (7 - bit%8)) & 1
}

func coveredByPrefixMap(prefixes map[netip.Prefix]struct{}, prefix netip.Prefix) bool {
	for bits := prefix.Bits(); bits >= 0; bits-- {
		candidate := netip.PrefixFrom(prefix.Addr(), bits).Masked()
		if _, exists := prefixes[candidate]; exists {
			return true
		}
	}
	return false
}

func prefixesOverlap(left, right netip.Prefix) bool {
	return left.Contains(right.Addr()) || right.Contains(left.Addr())
}

func prefixContains(outer, inner netip.Prefix) bool {
	return outer.Bits() <= inner.Bits() && outer.Contains(inner.Addr())
}

func prefixMapToSortedSlice(prefixes map[netip.Prefix]struct{}) PrefixSet {
	result := make(PrefixSet, 0, len(prefixes))
	for prefix := range prefixes {
		result = append(result, prefix)
	}
	sort.Slice(result, func(i, j int) bool {
		left, right := result[i], result[j]
		if left.Addr().BitLen() != right.Addr().BitLen() {
			return left.Addr().BitLen() < right.Addr().BitLen()
		}
		if left.Addr() != right.Addr() {
			return left.Addr().Less(right.Addr())
		}
		return left.Bits() < right.Bits()
	})
	return result
}
